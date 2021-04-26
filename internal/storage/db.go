package storage

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	bundle "github.com/bennycio/bundle/internal"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func InsertUser(user bundle.User) error {
	asJSON, _ := json.Marshal(user)
	validatedUser, err := bundle.ValidateAndReturnUser(string(asJSON))

	if err != nil {
		return err
	}

	bcryptPass, err := bcrypt.GenerateFromPassword([]byte(validatedUser.Password), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	session, err := getMongoSession()
	if err != nil {
		return err
	}
	defer session.Cancel()

	collection := session.Client.Database("main").Collection("users")

	countUserName, err := collection.CountDocuments(session.Ctx, bson.D{{"username", bundle.NewCaseInsensitiveRegex(validatedUser.Username)}})

	if err != nil {
		return err
	}

	if countUserName > 0 {
		err = errors.New("user already exists with given username")
		return err
	}

	countEmail, err := collection.CountDocuments(session.Ctx, bson.D{{"email", bundle.NewCaseInsensitiveRegex(validatedUser.Email)}})

	if err != nil {
		return err
	}

	if countEmail > 0 {
		err = errors.New("user already exists with given email")
		return err
	}

	_, err = collection.InsertOne(session.Ctx, bson.D{{"username", validatedUser.Username}, {"email", validatedUser.Email}, {"password", string(bcryptPass)}})

	if err != nil {
		return err
	}
	return nil
}

func GetUser(user bundle.User) (bundle.User, error) {
	session, err := getMongoSession()
	if err != nil {
		return bundle.User{}, err
	}
	defer session.Cancel()

	collection := session.Client.Database("main").Collection("users")

	decodedUser := &bundle.User{}

	if user.Email == "" {
		err = collection.FindOne(session.Ctx, bson.D{{"username", user.Username}}).Decode(decodedUser)
	} else if user.Username == "" {
		err = collection.FindOne(session.Ctx, bson.D{{"email", bundle.NewCaseInsensitiveRegex(user.Email)}}).Decode(decodedUser)
	} else {
		err = collection.FindOne(session.Ctx, bson.D{{"username", user.Username}, {"email", bundle.NewCaseInsensitiveRegex(user.Email)}}).Decode(decodedUser)
	}
	if err != nil {
		return bundle.User{}, err
	}

	return *decodedUser, nil
}

func InsertPlugin(plugin bundle.Plugin) error {
	session, err := getMongoSession()
	if err != nil {
		return err
	}
	defer session.Cancel()

	collection := session.Client.Database("main").Collection("plugins")

	_, err = collection.InsertOne(session.Ctx, bson.D{{"name", plugin.Name}, {"author", plugin.Author}, {"version", plugin.Version}, {"lastUpdated", time.Now().Unix()}})

	if err != nil {
		return err
	}
	return nil

}

func UpdatePlugin(name string, plugin bundle.Plugin) error {

	session, err := getMongoSession()
	if err != nil {
		return err
	}
	defer session.Cancel()

	collection := session.Client.Database("main").Collection("plugins")
	decodedPluginResult := &bundle.Plugin{}

	updateResult, err := collection.ReplaceOne(session.Ctx, bson.D{{"name", bundle.NewCaseInsensitiveRegex(name)}}, bson.D{{"name", decodedPluginResult.Name}, {"author", plugin.Author}, {"version", plugin.Version}, {"description", plugin.Description}, {"lastUpdated", time.Now().Unix()}})
	if err != nil {
		return err
	}
	if updateResult.MatchedCount < 1 {
		return errors.New("no plugin found")
	}
	return nil

}

func GetPlugin(name string) (bundle.Plugin, error) {
	if name == "" {
		return bundle.Plugin{}, errors.New("no plugin name provided")
	}

	session, err := getMongoSession()
	if err != nil {
		return bundle.Plugin{}, err
	}
	defer session.Cancel()

	collection := session.Client.Database("main").Collection("plugins")
	decodedPluginResult := &bundle.Plugin{}

	err = collection.FindOne(session.Ctx, bson.D{{"name", bundle.NewCaseInsensitiveRegex(name)}}).Decode(decodedPluginResult)
	if err != nil {
		return bundle.Plugin{}, err
	}

	return *decodedPluginResult, nil

}

func PaginatePlugins(page int) ([]bundle.Plugin, error) {
	session, err := getMongoSession()
	if err != nil {
		return nil, err
	}
	defer session.Cancel()

	findOptions := options.Find()
	// Sort by `price` field descending
	findOptions.SetSort(bson.D{{"lastUpdated", -1}})
	if page > 1 {
		findOptions.SetSkip(int64(page*10 - 10))
	}
	findOptions.SetLimit(10)

	collection := session.Client.Database("main").Collection("plugins")

	cur, err := collection.Find(session.Ctx, bson.D{}, findOptions)
	if err != nil {
		return nil, err
	}

	results := []bundle.Plugin{}
	defer cur.Close(session.Ctx)
	for cur.Next(session.Ctx) {
		plugin := bundle.Plugin{}
		if err = cur.Decode(&plugin); err != nil {
			return nil, err
		}
		results = append(results, plugin)
	}

	return results, nil

}

func getMongoSession() (*bundle.Mongo, error) {
	mg := &bundle.Mongo{}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(viper.GetString("MongoURL")))
	mg.Cancel = cancel
	mg.Client = client
	mg.Ctx = ctx
	if err != nil {
		return mg, err
	}
	return mg, nil
}

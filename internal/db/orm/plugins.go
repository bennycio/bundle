package orm

import (
	"errors"
	"time"

	"github.com/bennycio/bundle/api"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PluginsOrm struct{}

func NewPluginsOrm() *PluginsOrm { return &PluginsOrm{} }

func (p *PluginsOrm) Insert(plugin *api.Plugin) error {
	session, err := getMongoSession()
	if err != nil {
		return err
	}
	defer session.Cancel()

	collection := session.Client.Database("plugins").Collection("plugins")

	bs, err := bson.Marshal(plugin)
	if err != nil {
		return err
	}
	newPlugin := bson.D{}
	err = bson.Unmarshal(bs, &newPlugin)
	if err != nil {
		return err
	}
	newPlugin = append(newPlugin, bson.E{"lastUpdated", time.Now().Unix()})

	_, err = collection.InsertOne(session.Ctx, newPlugin)

	if err != nil {
		return err
	}
	return nil

}

func (p *PluginsOrm) Update(req *api.Plugin) error {

	session, err := getMongoSession()
	if err != nil {
		return err
	}
	defer session.Cancel()

	collection := session.Client.Database("plugins").Collection("plugins")

	updatedPlugin := marshallBsonClean(req)
	updatedPlugin = append(updatedPlugin, bson.E{"lastUpdated", time.Now().Unix()})

	if req.Id == "" {
		plorm := NewPluginsOrm()
		pl, err := plorm.Get(req)
		if err != nil {
			return err
		}
		req.Id = pl.Id
	}

	id, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return err
	}

	updateResult, err := collection.UpdateByID(session.Ctx, id, bson.D{{"$set", updatedPlugin}})
	if err != nil {
		return err
	}
	if updateResult.MatchedCount < 1 {
		return errors.New("no plugin found")
	}
	return nil

}

func (p *PluginsOrm) Get(req *api.Plugin) (*api.Plugin, error) {
	if req.Name == "" {
		return nil, errors.New("no plugin name provided")
	}

	session, err := getMongoSession()
	if err != nil {
		return nil, err
	}
	defer session.Cancel()

	collection := session.Client.Database("plugins").Collection("plugins")
	decodedPluginResult := &api.Plugin{}

	if req.Id == "" {
		err = collection.FindOne(session.Ctx, bson.D{{"name", caseInsensitive(req.Name)}}).Decode(decodedPluginResult)
		if err != nil {
			return nil, err
		}
	} else {
		err = collection.FindOne(session.Ctx, bson.D{{"_id", req.Id}}).Decode(decodedPluginResult)
		if err != nil {
			return nil, err
		}
	}

	return decodedPluginResult, nil

}

func (p *PluginsOrm) Paginate(req *api.PaginatePluginsRequest) (*api.PaginatePluginsResponse, error) {
	session, err := getMongoSession()
	if err != nil {
		return nil, err
	}
	defer session.Cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"lastUpdated", -1}})
	if req.Page > 1 {
		findOptions.SetSkip(int64(req.Page*req.Count - req.Count))
	}
	findOptions.SetLimit(int64(req.Count))

	collection := session.Client.Database("plugins").Collection("plugins")

	cur, err := collection.Find(session.Ctx, bson.D{}, findOptions)
	if err != nil {
		return nil, err
	}

	results := []*api.Plugin{}
	defer cur.Close(session.Ctx)
	for cur.Next(session.Ctx) {
		plugin := &api.Plugin{}
		if err = cur.Decode(&plugin); err != nil {
			return nil, err
		}
		results = append(results, plugin)
	}

	return &api.PaginatePluginsResponse{
		Plugins: results,
	}, nil

}
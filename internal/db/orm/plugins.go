package orm

import (
	"errors"
	"time"

	"github.com/bennycio/bundle/api"
	"github.com/bennycio/bundle/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type plugin struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name,omitempty" json:"name"`
	Description string             `bson:"description,omitempty" json:"description"`
	Author      primitive.ObjectID `bson:"author,omitempty" json:"author"`
	Version     string             `bson:"version,omitempty" json:"version"`
	Thumbnail   string             `bson:"thumbnail,omitempty" json:"thumbnail"`
	Category    category           `bson:"category,omitempty" json:"category"`
	Metadata    metadata           `bson:"metadata,omitempty" json:"metadata"`
	Premium     premium            `bson:"premium,omitempty" json:"premium"`
	LastUpdated primitive.DateTime `bson:"lastUpdated,omitempty" json:"lastUpdated"`
}

type premium struct {
	Price     int32 `bson:"price,omitempty" json:"price"`
	Purchases int32 `bson:"purchases,omitempty" json:"purchases"`
}

type metadata struct {
	Downloads int64    `bson:"downloads,omitempty" json:"price"`
	Conflicts []string `bson:"conflicts,omitempty" json:"purchases"`
}
type category int32

const (
	all category = 0 << iota
	prem
	tools
	eco
	chat
	mech
	admin
	bungee
	fun
	misc
	lib
)

type PluginsOrm struct{}

func NewPluginsOrm() *PluginsOrm { return &PluginsOrm{} }

func (p *PluginsOrm) Insert(pl *api.Plugin) error {
	mgses, err := getMongoSession()
	if err != nil {
		logger.ErrLog.Print(err.Error())
		return err
	}
	defer mgses.Cancel()

	collection := mgses.Client.Database("plugins").Collection("plugins")

	countName, err := collection.CountDocuments(mgses.Ctx, bson.D{{"name", pl.Name}}, options.Count().SetCollation(&options.Collation{Locale: "en", Strength: 1}))

	if err != nil {
		return err
	}

	if countName > 0 {
		err = errors.New("plugin already exists with given name")
		logger.ErrLog.Print(err.Error())
		return err
	}

	insertion := apiToOrmPl(pl)
	err = validatePluginInsert(insertion)
	if err != nil {
		logger.ErrLog.Print(err.Error())
		return err
	}

	_, err = collection.InsertOne(mgses.Ctx, insertion)

	if err != nil {
		logger.ErrLog.Print(err.Error())
		return err
	}
	return nil

}

func (p *PluginsOrm) Update(req *api.Plugin) error {

	mgses, err := getMongoSession()
	if err != nil {
		logger.ErrLog.Print(err.Error())
		return err
	}
	defer mgses.Cancel()

	collection := mgses.Client.Database("plugins").Collection("plugins")

	update := apiToOrmPl(req)
	err = validatePluginUpdate(update)
	if err != nil {
		logger.ErrLog.Print(err.Error())
		return err
	}

	var updateResult *mongo.UpdateResult

	if update.Id != primitive.NilObjectID {
		updateResult, err = collection.UpdateByID(mgses.Ctx, update.Id, bson.D{{"$set", update}}, &options.UpdateOptions{Upsert: boolin(true)})
	} else {
		updateResult, err = collection.UpdateOne(mgses.Ctx, bson.D{{"name", update.Name}}, bson.D{{"$set", update}}, &options.UpdateOptions{Upsert: boolin(true)})
	}

	if err != nil {
		logger.ErrLog.Print(err.Error())
		return err
	}
	if updateResult.ModifiedCount < 1 && updateResult.UpsertedCount < 1 {
		err = errors.New("no plugin found")
		logger.ErrLog.Print(err.Error())
		return err
	}

	return nil

}

func (p *PluginsOrm) Get(req *api.Plugin) (*api.Plugin, error) {
	session, err := getMongoSession()
	if err != nil {
		logger.ErrLog.Print(err.Error())
		return nil, err
	}
	defer session.Cancel()

	collection := session.Client.Database("plugins").Collection("plugins")
	decodedPluginResult := &plugin{}

	get := apiToOrmPl(req)
	err = validatePluginGet(get)
	if err != nil {
		logger.ErrLog.Print(err.Error())
		return nil, err
	}

	if get.Id == primitive.NilObjectID {
		res := collection.FindOne(session.Ctx, bson.D{{"name", get.Name}}, options.FindOne().SetCollation(&options.Collation{Locale: "en", Strength: 1}))
		if res.Err() != nil {
			logger.ErrLog.Print(res.Err().Error())
			return nil, res.Err()
		}
		res.Decode(decodedPluginResult)
	} else {
		res := collection.FindOne(session.Ctx, bson.D{{"_id", get.Id}})
		if res.Err() != nil {
			logger.ErrLog.Print(res.Err().Error())
			return nil, res.Err()
		}
		res.Decode(decodedPluginResult)
	}

	return ormToApiPl(*decodedPluginResult), nil

}

func (p *PluginsOrm) Paginate(req *api.PaginatePluginsRequest) ([]*api.Plugin, error) {
	mgses, err := getMongoSession()
	if err != nil {
		logger.ErrLog.Print(err.Error())
		return nil, err
	}
	defer mgses.Cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"lastUpdated", -1}})
	if req.Page > 1 {
		findOptions.SetSkip(int64(req.Page*req.Count - req.Count))
	}
	findOptions.SetLimit(int64(req.Count))

	collection := mgses.Client.Database("plugins").Collection("plugins")

	fil := bson.D{}

	if req.Search != "" {
		fil = append(fil, bson.E{"$text", bson.D{{"$search", req.Search}}})
	}
	if req.Category != api.Category_ALL {
		fil = append(fil, bson.E{"category", req.Category})
	}
	if req.Sort != api.Sort_NONE {
		switch req.Sort {
		case api.Sort_LATEST:
			findOptions.SetSort(bson.D{{"lastUpdated", -1}})
		case api.Sort_PURCHASES:
			findOptions.SetSort(bson.D{{"premium.purchases", -1}})
		case api.Sort_DOWNLOADS:
			findOptions.SetSort(bson.D{{"downloads", -1}})
		}
	}

	cur, err := collection.Find(mgses.Ctx, fil, findOptions)
	if err != nil {
		logger.ErrLog.Print(err.Error())
		return nil, err
	}

	results := []*api.Plugin{}
	defer cur.Close(mgses.Ctx)
	for cur.Next(mgses.Ctx) {
		pl := &plugin{}
		if err = cur.Decode(&pl); err != nil {
			logger.ErrLog.Print(err.Error())
			return nil, err
		}
		results = append(results, ormToApiPl(*pl))
	}

	return results, nil

}

func validatePluginUpdate(pl plugin) error {
	if pl.Id == primitive.NilObjectID && pl.Name == "" {
		return errors.New("id or name required for update")
	}
	return nil
}

func validatePluginInsert(pl plugin) error {
	if pl.Name == "" {
		return errors.New("name required for insertion")
	}
	if pl.Version == "" {
		return errors.New("version required for insertion")
	}
	if pl.Author == primitive.NilObjectID {
		return errors.New("author id required for insertion")
	}
	return nil
}

func validatePluginGet(pl plugin) error {
	if pl.Name == "" && pl.Id == primitive.NilObjectID {
		return errors.New("id or name required for get")
	}
	return nil
}

func ormToApiPl(pl plugin) *api.Plugin {
	p := &api.Plugin{
		Id:          pl.Id.Hex(),
		Name:        pl.Name,
		Description: pl.Description,
		Version:     pl.Version,
		Thumbnail:   pl.Thumbnail,
		Category:    api.Category(pl.Category),
		Metadata: &api.PluginMetadata{
			Downloads: pl.Metadata.Downloads,
			Conflicts: pl.Metadata.Conflicts,
		},
		Premium: &api.Premium{
			Price:     pl.Premium.Price,
			Purchases: pl.Premium.Purchases,
		},
		LastUpdated: pl.LastUpdated.Time().Unix(),
	}
	a, err := NewUsersOrm().Get(&api.User{Id: pl.Author.Hex()})
	if err == nil {
		p.Author = a
	}
	return p
}

func apiToOrmPl(pl *api.Plugin) plugin {

	if pl == nil {
		return plugin{}
	}

	var lastUpdated primitive.DateTime
	if pl.LastUpdated != 0 {
		lastUpdated = primitive.DateTime(pl.LastUpdated)
	} else {
		lastUpdated = primitive.DateTime(time.Now().Unix())
	}
	result := plugin{
		Name:        pl.Name,
		Description: pl.Description,
		Version:     pl.Version,
		Thumbnail:   pl.Thumbnail,
		LastUpdated: lastUpdated,
		Category:    category(pl.Category),
	}
	pluginID, err := primitive.ObjectIDFromHex(pl.Id)
	if pluginID != primitive.NilObjectID && err == nil {
		result.Id = pluginID
	}
	if pl.Author != nil {
		authorID, err := primitive.ObjectIDFromHex(pl.Author.Id)
		if authorID != primitive.NilObjectID && err == nil {
			result.Author = authorID
		}
	}
	if pl.Premium != nil {
		result.Premium = premium{
			Price:     pl.Premium.Price,
			Purchases: pl.Premium.Purchases,
		}
	}
	if pl.Metadata != nil {
		result.Metadata = metadata{
			Downloads: pl.Metadata.Downloads,
			Conflicts: pl.Metadata.Conflicts,
		}
	}

	return result

}

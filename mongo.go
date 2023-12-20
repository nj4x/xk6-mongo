package xk6_mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	k6modules "go.k6.io/k6/js/modules"
)

// Register the extension on module initialization, available to
// import from JS as "k6/x/mongo".
func init() {
	k6modules.Register("k6/x/mongo", new(Mongo))
}

// Mongo is the k6 extension for a Mongo client.
type Mongo struct{}

// Client is the Mongo client wrapper.
type Client struct {
	client                      *mongo.Client
	tolerateUnacknowledgedWrite bool
}

// NewClient represents the Client constructor (i.e. `new mongo.Client()`) and
// returns a new Mongo client object.
// connURI -> mongodb://username:password@address:port/db?connect=direct
func (*Mongo) NewClient(connURI string, unacknowledgedWriteConcern bool) interface{} {

	clientOptions := options.Client().ApplyURI(connURI)
	if unacknowledgedWriteConcern {
		clientOptions.SetWriteConcern(writeconcern.Unacknowledged())
	}
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		panic(err)
		return nil
	}

	return &Client{client: client, tolerateUnacknowledgedWrite: unacknowledgedWriteConcern}
}

func (*Mongo) HexToObjectID(id string) primitive.ObjectID {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		panic(err)
	}
	return oid
}

func (c *Client) Count(database string, collection string, filter interface{}, limit int64, skip int64) int64 {
	db := c.client.Database(database)
	col := db.Collection(collection)
	opts := options.Count().SetSkip(skip)
	if limit > 0 {
		opts.SetLimit(limit)
	}
	cnt, err := col.CountDocuments(context.TODO(), filter, opts)
	c.CheckError(err)
	return cnt
}

// Insert returns primitive.ObjectID
func (c *Client) Insert(database string, collection string, doc map[string]any) interface{} {
	db := c.client.Database(database)
	col := db.Collection(collection)
	res, err := col.InsertOne(context.TODO(), doc)
	c.CheckError(err)
	return res.InsertedID
}

// InsertMany returns []primitive.ObjectID
func (c *Client) InsertMany(database string, collection string, docs []any) []interface{} {
	db := c.client.Database(database)
	col := db.Collection(collection)
	res, err := col.InsertMany(context.TODO(), docs)
	c.CheckError(err)
	return res.InsertedIDs
}

func (c *Client) Find(database string, collection string, filter interface{}, limit int64, skip int64, projection interface{}) []bson.M {
	db := c.client.Database(database)
	col := db.Collection(collection)
	opts := options.Find().SetSkip(skip).SetProjection(projection)
	if limit > 0 {
		opts.SetLimit(limit)
	}
	cur, err := col.Find(context.TODO(), filter, opts)
	c.CheckError(err)
	var results []bson.M
	if err = cur.All(context.TODO(), &results); err != nil {
		c.CheckError(err)
	}
	return results
}

func (c *Client) FindOne(database string, collection string, filter map[string]any, skip int64) bson.M {
	db := c.client.Database(database)
	col := db.Collection(collection)
	var result bson.M
	opts := options.FindOne().SetSort(bson.D{{"_id", 1}}).SetSkip(skip)
	err := col.FindOne(context.TODO(), filter, opts).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return bson.M{}
	}
	c.CheckError(err)
	return result
}

func (c *Client) UpdateOne(database string, collection string, filter interface{}, data map[string]any) int64 {
	db := c.client.Database(database)

	col := db.Collection(collection)

	update := bson.A{bson.D{{"$set", data}}}
	res, err := col.UpdateOne(context.TODO(), filter, update)
	c.CheckError(err)
	return res.MatchedCount
}

func (c *Client) CheckError(err interface{ Error() string }) {
	if err != nil && (err.Error() != "unacknowledged write" || !c.tolerateUnacknowledgedWrite) {
		panic(err)
	}
}

func (c *Client) DeleteOne(database string, collection string, filter map[string]any) error {
	db := c.client.Database(database)
	col := db.Collection(collection)
	opts := options.Delete().SetHint(bson.D{{"_id", 1}})
	_, err := col.DeleteOne(context.TODO(), filter, opts)
	c.CheckError(err)
	return nil
}

func (c *Client) DeleteMany(database string, collection string, filter map[string]any) error {
	db := c.client.Database(database)
	col := db.Collection(collection)
	opts := options.Delete().SetHint(bson.D{{"_id", 1}})
	_, err := col.DeleteMany(context.TODO(), filter, opts)
	c.CheckError(err)
	return nil
}

func (c *Client) DropCollection(database string, collection string) error {
	db := c.client.Database(database)
	col := db.Collection(collection)
	err := col.Drop(context.TODO())
	c.CheckError(err)
	return nil
}

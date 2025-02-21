package resolvers

import (
	"context"
	"log"
	"time"

	"github.com/graphql-go/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func UsersCollection() *mongo.Collection {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongo:27017"))

	if err != nil {
		log.Panic("Error when connecting to mongodb", err)
	}

	collection := client.Database("testing").Collection("users")

	return collection
}

func UserResolver(_ graphql.ResolveParams) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := UsersCollection()
	result, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Print("Error in finding user", err)
		return nil, err
	}
	defer result.Close(ctx)

	var r []bson.M
	err = result.All(ctx, &r)
	if err != nil {
		log.Print("Error in reading users from cursor", err)
	}
	return r, nil
}

func AddUserResolver(p graphql.ResolveParams) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := UsersCollection()
	id, err := collection.InsertOne(ctx, p.Args["input"])
	if err != nil {
		log.Print("Error in inserting user", err)
		return nil, err
	}

	var result bson.M
	err = collection.FindOne(ctx, bson.M{"_id": id.InsertedID}).Decode(&result)
	if err != nil {
		log.Print("Error in finding the inserted user by id", err)
		return nil, err
	}

	return result, nil
}

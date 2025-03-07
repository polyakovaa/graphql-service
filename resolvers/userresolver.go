package resolvers

import (
	"context"
	"errors"
	"grphqlserver/auth"
	"log"
	"time"

	"github.com/graphql-go/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func UserRegisterResolver(p graphql.ResolveParams) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := UsersCollection()

	username := p.Args["username"].(string)
	password := p.Args["password"].(string)

	var existingUser bson.M
	err := collection.FindOne(ctx, bson.M{"username": username}).Decode(&existingUser)
	if err == nil {
		return nil, errors.New("user already exists")
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	res, err := collection.InsertOne(ctx, bson.M{
		"username": username,
		"password": hashedPassword,
	})
	if err != nil {
		return nil, err
	}
	token, err := auth.GenerateToken(res.InsertedID.(primitive.ObjectID).Hex())
	if err != nil {
		return nil, err
	}

	return bson.M{"token": token}, nil

}

func LoginUserResolver(p graphql.ResolveParams) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := UsersCollection()

	username := p.Args["username"].(string)
	password := p.Args["password"].(string)

	var user bson.M
	err := collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	if !auth.CheckPassword(password, user["password"].(string)) {
		return nil, errors.New("invalid username or password")
	}

	token, err := auth.GenerateToken(user["_id"].(primitive.ObjectID).Hex())
	if err != nil {
		return nil, err
	}

	return bson.M{"token": token}, nil
}

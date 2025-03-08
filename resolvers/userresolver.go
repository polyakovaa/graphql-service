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
	"golang.org/x/crypto/bcrypt"
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

func RegisterUserResolver(p graphql.ResolveParams) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := UsersCollection()

	input, _ := p.Args["input"].(map[string]interface{})
	username, _ := input["userName"].(string)
	password, _ := input["password"].(string)
	email, _ := input["email"].(string)

	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	var existingUser bson.M
	err := collection.FindOne(ctx, bson.M{"userName": username}).Decode(&existingUser)
	if err == nil {
		return nil, errors.New("username already exists")
	} else if err != mongo.ErrNoDocuments {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUser := bson.M{
		"userName": username,
		"password": string(hashedPassword),
		"email":    email,
	}

	id, err := collection.InsertOne(ctx, newUser)
	if err != nil {
		return nil, err
	}

	token, err := auth.GenerateToken(id.InsertedID.(primitive.ObjectID).Hex())
	if err != nil {
		return nil, err
	}

	return token, nil
}

func LoginUserResolver(p graphql.ResolveParams) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := UsersCollection()

	input, _ := p.Args["input"].(map[string]interface{})
	username, _ := input["userName"].(string)
	password, _ := input["password"].(string)

	var user bson.M
	err := collection.FindOne(ctx, bson.M{"userName": username}).Decode(&user)
	if err != nil {
		return nil, errors.New("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user["password"].(string)), []byte(password))
	if err != nil {
		return nil, errors.New("invalid password")
	}

	token, err := auth.GenerateToken(user["_id"].(primitive.ObjectID).Hex())
	if err != nil {
		return nil, err
	}

	return token, nil
}

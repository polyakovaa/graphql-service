package resolvers

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/graphql-go/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func BooksCollection() *mongo.Collection {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongo:27017"))

	if err != nil {
		log.Panic("Error when connecting to mongodb", err)
	}

	collection := client.Database("testing").Collection("books")

	return collection
}

func BookResolver(_ graphql.ResolveParams) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := BooksCollection()
	result, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Print("Error in finding book", err)
		return nil, err
	}
	defer result.Close(ctx)

	var r []bson.M
	err = result.All(ctx, &r)
	if err != nil {
		log.Print("Error in reading books from cursor", err)
	}
	return r, nil
}

func AddBookResolver(p graphql.ResolveParams) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := BooksCollection()
	id, err := collection.InsertOne(ctx, p.Args["input"])
	if err != nil {
		log.Print("Error in inserting book", err)
		return nil, err
	}

	var result bson.M
	err = collection.FindOne(ctx, bson.M{"_id": id.InsertedID}).Decode(&result)
	if err != nil {
		log.Print("Error in finding the inserted book by id", err)
		return nil, err
	}

	return result, nil
}

func UpdateBookResover(p graphql.ResolveParams) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := BooksCollection()
	id, ok := p.Args["_id"].(primitive.ObjectID)
	if !ok {
		return nil, errors.New("invalid book ID")
	}

	input, ok := p.Args["input"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalide input data")
	}
	update := bson.M{"$set": input}
	_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		log.Print("Error updating book:", err)
		return nil, err
	}
	var updatedBook bson.M
	err = collection.FindOne(ctx, bson.M{"_id": id}).Decode(&updatedBook)

	if err != nil {
		log.Print("Error reteieving updated book:", err)
		return nil, err
	}
	return updatedBook, nil
}

func DeleteBookResolver(p graphql.ResolveParams) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := BooksCollection()

	id, ok := p.Args["_id"].(primitive.ObjectID)
	if !ok {
		return nil, errors.New("invalid book ID")
	}

	res, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		log.Print("Error deleting book: ", err)
		return nil, err
	}

	if res.DeletedCount == 0 {
		return false, nil
	}

	return map[string]interface{}{
		"success": true,
		"message": "Book deleted successfully",
	}, nil
}

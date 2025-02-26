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

func ReviewCollection() *mongo.Collection {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongo:27017"))

	if err != nil {
		log.Panic("Error when connecting to mongodb", err)
	}

	collection := client.Database("testing").Collection("reviews")

	return collection
}

func ReviewResolver(_ graphql.ResolveParams) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := ReviewCollection()
	result, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Print("Error in finding review", err)
		return nil, err
	}
	defer result.Close(ctx)

	var r []bson.M
	err = result.All(ctx, &r)
	if err != nil {
		log.Print("Error in reading review from cursor", err)
	}
	return r, nil
}

func AddReviewResolver(p graphql.ResolveParams) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := ReviewCollection()
	id, err := collection.InsertOne(ctx, p.Args["input"])
	if err != nil {
		log.Print("Error in inserting review", err)
		return nil, err
	}

	var result bson.M
	err = collection.FindOne(ctx, bson.M{"_id": id.InsertedID}).Decode(&result)
	if err != nil {
		log.Print("Error in finding the inserted review by id", err)
		return nil, err
	}

	return result, nil
}

func DeleteReviewResolver(p graphql.ResolveParams) (interface{}, error) {
	collection := ReviewCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reviewID, ok := p.Args["reviewID"].(string)
	if !ok {
		return nil, errors.New("invalid reviewID")
	}
	objID, err := primitive.ObjectIDFromHex(reviewID)
	if err != nil {
		return nil, errors.New("invalid ObjectID format")
	}

	result, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		log.Print("Error deleting review:", err)
		return nil, err
	}

	if result.DeletedCount == 0 {
		return nil, errors.New("review not found")
	}
	return true, nil

}

func UpdateReviewResolver(p graphql.ResolveParams) (interface{}, error) {
	collection := ReviewCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reviewID, ok := p.Args["reviewID"].(string)
	if !ok {
		return nil, errors.New("invalid reviewID")
	}
	objID, err := primitive.ObjectIDFromHex(reviewID)
	if err != nil {
		return nil, errors.New("invalid ObjectID format")
	}

	input, ok := p.Args["input"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid input data")
	}
	update := bson.M{"$set": input}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		log.Print("Error updating review:", err)
		return nil, err
	}

	if result.MatchedCount == 0 {
		return nil, errors.New("review not found")
	}

	var updatedReview bson.M
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updatedReview)
	if err != nil {
		return nil, err
	}

	return updatedReview, nil

}

func FindReviewsResolver(p graphql.ResolveParams) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := ReviewCollection()
	filter := bson.M{}

	if title, ok := p.Args["title"].(string); ok && title != "" {
		filter["title"] = bson.M{"$regex": title, "$options": "i"}
	}

	if author, ok := p.Args["author"].(string); ok && author != "" {
		filter["author"] = bson.M{"$regex": author, "$options": "i"}
	}

	cursor, err := collection.Find(ctx, filter, options.Find())
	if err != nil {
		log.Println("Error finding reviews ", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var reviews []bson.M
	if err = cursor.All(ctx, &reviews); err != nil {
		log.Println("Error reading reviiew from cursor: ", err)
		return nil, err
	}

	if len(reviews) == 0 {
		return nil, errors.New("reviews not found")
	}

	return reviews, nil
}

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

	input, ok := p.Args["input"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid input data")
	}

	if _, exists := input["date"]; !exists {
		input["date"] = time.Now()
	}

	res, err := collection.InsertOne(ctx, input)
	if err != nil {
		log.Print("Error inserting review:", err)
		return nil, err
	}

	input["_id"] = res.InsertedID
	return input, nil
}

func DeleteReviewResolver(p graphql.ResolveParams) (interface{}, error) {
	collection := ReviewCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id, ok := p.Args["_id"].(primitive.ObjectID)
	if !ok {
		return nil, errors.New("missing review ID")
	}

	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
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

	id, ok := p.Args["_id"].(primitive.ObjectID)
	if !ok {
		return nil, errors.New("missing or invalid reviex ID")
	}

	input, ok := p.Args["input"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid input data")
	}

	update := bson.M{"$set": input}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		log.Print("Error updating review:", err)
		return nil, err
	}

	if result.MatchedCount == 0 {
		return nil, errors.New("review not found")
	}

	var updatedReview bson.M
	err = collection.FindOne(ctx, bson.M{"_id": id}).Decode(&updatedReview)
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

	if bookID, ok := p.Args["bookID"].(primitive.ObjectID); ok {
		filter["bookID"] = bookID
	}

	var bookFilter bson.M
	title, titleOk := p.Args["title"].(string)
	author, authorOk := p.Args["author"].(string)

	if titleOk && title != "" || authorOk && author != "" {
		bookFilter = bson.M{}

		if titleOk && title != "" {
			bookFilter["title"] = bson.M{"$regex": title, "$options": "i"}
		}
		if authorOk && author != "" {
			bookFilter["author"] = bson.M{"$regex": author, "$options": "i"}
		}

		booksCollection := BooksCollection()
		cursor, err := booksCollection.Find(ctx, bookFilter)
		if err != nil {
			log.Println("Error finding books:", err)
			return nil, err
		}
		defer cursor.Close(ctx)

		var books []bson.M
		if err = cursor.All(ctx, &books); err != nil {
			log.Println("Error reading books from cursor:", err)
			return nil, err
		}

		if len(books) == 0 {
			return nil, errors.New("no books found with the given title or author")
		}

		var bookIDs []primitive.ObjectID
		for _, book := range books {
			if id, ok := book["_id"].(primitive.ObjectID); ok {
				bookIDs = append(bookIDs, id)

			}
		}

		filter["bookID"] = bson.M{"$in": bookIDs}
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Println("Error finding reviews:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var reviews []bson.M
	if err = cursor.All(ctx, &reviews); err != nil {
		log.Println("Error reading reviews from cursor:", err)
		return nil, err
	}

	for i, review := range reviews {
		if date, ok := review["date"]; ok {
			if dt, isDate := date.(primitive.DateTime); isDate {
				reviews[i]["date"] = dt.Time()
			} else {
				log.Printf("Review %d has an invalid date format: %T\n", i, date)
			}
		}
	}

	if len(reviews) == 0 {
		return nil, errors.New("reviews not found")
	}

	return reviews, nil
}

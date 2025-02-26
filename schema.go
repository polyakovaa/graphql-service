package main

import (
	"grphqlserver/resolvers"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ObjectID = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "BSON",
	Description: "The `bson` scalar type represents a BSON Object.",
	// Serialize serializes `bson.ObjectId` to string.
	Serialize: func(value interface{}) interface{} {
		switch value := value.(type) {
		case primitive.ObjectID:
			return value.Hex()
		case *primitive.ObjectID:
			v := *value
			return v.Hex()
		default:
			return nil
		}
	},
	// ParseValue parses GraphQL variables from `string` to `bson.ObjectId`.
	ParseValue: func(value interface{}) interface{} {
		switch value := value.(type) {
		case string:
			id, _ := primitive.ObjectIDFromHex(value)
			return id
		case *string:
			id, _ := primitive.ObjectIDFromHex(*value)
			return id
		default:
			return nil
		}
	},
	// ParseLiteral parses GraphQL AST to `bson.ObjectId`.
	ParseLiteral: func(valueAST ast.Value) interface{} {
		switch valueAST := valueAST.(type) {
		case *ast.StringValue:
			id, _ := primitive.ObjectIDFromHex(valueAST.Value)
			return id
		}
		return nil
	},
})

var User = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"_id": &graphql.Field{
				Type: ObjectID,
			},
			"firstName": &graphql.Field{
				Type: graphql.String,
			},
			"lastName": &graphql.Field{
				Type: graphql.String,
			},
			"email": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

var Book = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Book",
		Fields: graphql.Fields{
			"_id": &graphql.Field{
				Type: ObjectID,
			},
			"author": &graphql.Field{
				Type: graphql.String,
			},
			"title": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

var UserInput = graphql.NewInputObject(
	graphql.InputObjectConfig{
		Name: "UserInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"firstName": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"lastName": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"email": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
		},
	},
)

var BookInput = graphql.NewInputObject(
	graphql.InputObjectConfig{
		Name: "BookInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"author": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"title": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
		},
	},
)

var Review = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Review",
		Fields: graphql.Fields{
			"_id": &graphql.Field{
				Type: ObjectID,
			},
			"bookID": &graphql.Field{
				Type: ObjectID,
			},
			"userID": &graphql.Field{
				Type: ObjectID,
			},
			"rating": &graphql.Field{
				Type: graphql.Int,
			},
			"comment": &graphql.Field{
				Type: graphql.String,
			},
			"date": &graphql.Field{
				Type: graphql.DateTime,
			},
		},
	},
)

var ReviewInput = graphql.NewInputObject(
	graphql.InputObjectConfig{
		Name: "ReviewInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"bookID":  &graphql.InputObjectFieldConfig{Type: ObjectID},
			"userID":  &graphql.InputObjectFieldConfig{Type: ObjectID},
			"rating":  &graphql.InputObjectFieldConfig{Type: graphql.String},
			"comment": &graphql.InputObjectFieldConfig{Type: graphql.String},
		},
	},
)

func defineSchema() graphql.SchemaConfig {
	return graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"users": &graphql.Field{
					Name:    "users",
					Type:    graphql.NewList(User),
					Resolve: resolvers.UserResolver,
				},
				"books": &graphql.Field{
					Name:    "books",
					Type:    graphql.NewList(Book),
					Resolve: resolvers.BookResolver,
				},
				"findBooks": &graphql.Field{
					Name: "findBooks",
					Type: graphql.NewList(Book),
					Args: graphql.FieldConfigArgument{
						"title": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
						"author": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
					},
					Resolve: resolvers.FindBooksResolver,
				},
				"findReviews": &graphql.Field{
					Name: "findReviews",
					Type: graphql.NewList(Review),
					Args: graphql.FieldConfigArgument{
						"bookID": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(ObjectID),
						},
					},
					Resolve: resolvers.FindReviewsResolver,
				},
			},
		}),
		Mutation: graphql.NewObject(graphql.ObjectConfig{
			Name: "Mutation",
			Fields: graphql.Fields{
				"addUser": &graphql.Field{
					Name:    "addUser",
					Type:    User,
					Resolve: resolvers.AddUserResolver,
					Args: graphql.FieldConfigArgument{
						"input": &graphql.ArgumentConfig{
							Type: UserInput,
						},
					},
				},
				"addBook": &graphql.Field{
					Name:    "addBook",
					Type:    Book,
					Resolve: resolvers.AddBookResolver,
					Args: graphql.FieldConfigArgument{
						"input": &graphql.ArgumentConfig{
							Type: BookInput,
						},
					},
				},
				"updateBook": &graphql.Field{
					Name: "updateBook",
					Type: Book,
					Args: graphql.FieldConfigArgument{
						"_id": &graphql.ArgumentConfig{
							Type: ObjectID,
						},
						"input": &graphql.ArgumentConfig{
							Type: BookInput,
						},
					},
					Resolve: resolvers.UpdateBookResolver,
				},
				"deleteBook": &graphql.Field{
					Name: "deleteBook",
					Type: graphql.Boolean,
					Args: graphql.FieldConfigArgument{
						"_id": &graphql.ArgumentConfig{
							Type: ObjectID,
						},
					},
					Resolve: resolvers.DeleteBookResolver,
				},
				"addReview": &graphql.Field{
					Name: "addReview",
					Type: Review,
					Args: graphql.FieldConfigArgument{
						"input": &graphql.ArgumentConfig{
							Type: ReviewInput,
						},
					},
					Resolve: resolvers.AddReviewResolver,
				},
				"updateReview": &graphql.Field{
					Name: "updateReview",
					Type: Review,
					Args: graphql.FieldConfigArgument{
						"reviewID": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(ObjectID),
						},
						"input": &graphql.ArgumentConfig{
							Type: ReviewInput,
						},
					},
					Resolve: resolvers.UpdateReviewResolver,
				},

				"deleteReview": &graphql.Field{
					Name: "deleteReview",
					Type: graphql.Boolean,
					Args: graphql.FieldConfigArgument{
						"reviewID": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(ObjectID),
						},
					},
					Resolve: resolvers.DeleteReviewResolver,
				},
			},
		}),
	}
}

package middleware

import (
	"context"
	"errors"
	"grphqlserver/auth"

	"github.com/graphql-go/graphql"
)

func AuthMiddleware(next graphql.FieldResolveFn) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		tokenString, ok := p.Context.Value("Authorization").(string)
		if !ok || tokenString == "" {
			return nil, errors.New("missing token")
		}

		userID, err := auth.ValidateToken(tokenString)
		if err != nil {
			return nil, err
		}

		p.Context = context.WithValue(p.Context, "userID", userID)

		return next(p)
	}
}

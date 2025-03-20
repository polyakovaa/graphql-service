package middleware

import (
	"context"
	"errors"
	"grphqlserver/auth"
	"net/http"
	"strings"

	"github.com/graphql-go/graphql"
)

func AuthMiddleware(next graphql.FieldResolveFn) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {

		authHeader, ok := p.Context.Value("Authorization").(string)
		if !ok || authHeader == "" {
			return nil, errors.New("missing token")
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return nil, errors.New("invalid token format")
		}

		userID, err := auth.ValidateToken(tokenString)
		if err != nil {
			return nil, err
		}

		p.Context = context.WithValue(p.Context, "userID", userID)

		return next(p)
	}
}

func InjectHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		ctx := context.WithValue(r.Context(), "Authorization", authHeader)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

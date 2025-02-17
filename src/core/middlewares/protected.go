package middlewares

import (
	"context"
	"errors"
	"net/http"

	"sme-backend/src/core/config"
	"sme-backend/src/core/database"
	"sme-backend/src/core/helpers"
	"sme-backend/src/enums/context_keys"
	"sme-backend/src/enums/user_types"

	"github.com/go-chi/jwtauth/v5"
)

func Protected(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		token := jwtauth.TokenFromHeader(request)
		jwtVerify(response, request, next, token, false)
	})
}

func AdminProtected(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		token := jwtauth.TokenFromHeader(request)
		jwtVerify(response, request, next, token, true)
	})
}

func PartiallyProtected(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		token := jwtauth.TokenFromHeader(request)
		if token != "" {
			jwtVerify(response, request, next, token, false)
		} else {
			db := database.CONTEXT_API_DB
			ctx := context.WithValue(request.Context(), context_keys.DB, db)
			next.ServeHTTP(response, request.WithContext(ctx))
		}
	})
}

func jwtVerify(response http.ResponseWriter, request *http.Request, next http.Handler, token string, is_admin bool) {
	tokenAuth := jwtauth.New("HS256", []byte(config.Config("JWT_SECRET")), nil)
	payload, err := jwtauth.VerifyToken(tokenAuth, token)
	if err != nil {
		helpers.HandleError(response, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	user_id := payload.Subject()
	user_type, ok := payload.Get(string(context_keys.USER_TYPE))
	if !ok {
		helpers.HandleError(response, http.StatusUnauthorized, "Invalid user type", errors.New("missing user type"))
		return
	}

	// CHECK if the user is ADMIN
	if is_admin && user_type != user_types.ADMIN {
		helpers.HandleError(response, http.StatusUnauthorized, "Unauthorized", errors.New("unauthorized user"))
		return
	}

	// Set Currennt context user's DB
	db := database.CONTEXT_API_DB
	database.SetRLS(db, user_id)
	ctx := context.WithValue(request.Context(), context_keys.USER_ID, user_id)
	ctx = context.WithValue(ctx, context_keys.USER_TYPE, user_type) // Add user_type to context
	ctx = context.WithValue(ctx, context_keys.DB, db)

	next.ServeHTTP(response, request.WithContext(ctx))
}

func VerifyJwtTokenExpiration(request *http.Request, token string) (string, error) {
	tokenAuth := jwtauth.New("HS256", []byte(config.Config("JWT_SECRET")), nil)
	payload, err := jwtauth.VerifyToken(tokenAuth, token)
	var user_id string
	if err != nil {
		return user_id, err
	}
	user_id = payload.Subject()
	return user_id, err
}

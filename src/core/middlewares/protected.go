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

// Middleware for protected routes
func Protected(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		token := jwtauth.TokenFromHeader(request)
		jwtVerify(response, request, next, token, false)
	})
}

// Middleware for admin-protected routes
func AdminProtected(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		token := jwtauth.TokenFromHeader(request)
		jwtVerify(response, request, next, token, true)
	})
}

// Middleware that allows both protected and unprotected access
func PartiallyProtected(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		token := jwtauth.TokenFromHeader(request)

		db := database.API_USER_DB
		ctx := context.WithValue(request.Context(), context_keys.DB, db)

		if token != "" {
			jwtVerify(response, request, next, token, false) // Handles RLS setup
		} else {
			next.ServeHTTP(response, request.WithContext(ctx)) // Skip RLS if no token
		}
	})
}

// Validates JWT and applies RLS
func jwtVerify(response http.ResponseWriter, request *http.Request, next http.Handler, token string, is_admin bool) {
	if token == "" {
		helpers.HandleError(response, http.StatusUnauthorized, "Missing token", errors.New("token required"))
		return
	}

	tokenAuth := jwtauth.New("HS256", []byte(config.Config("JWT_SECRET")), nil)

	// Verify JWT before extracting claims
	user_id, err := VerifyJwtTokenExpiration(request, token)
	if err != nil {
		helpers.HandleError(response, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	// Extract claims after verification
	payload, err := jwtauth.VerifyToken(tokenAuth, token)
	if err != nil {
		helpers.HandleError(response, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	user_type, ok := payload.Get(string(context_keys.USER_TYPE))
	if !ok {
		helpers.HandleError(response, http.StatusUnauthorized, "Invalid user type", errors.New("missing user type"))
		return
	}

	// Block access if user is not admin
	if is_admin && user_type != user_types.ADMIN {
		helpers.HandleError(response, http.StatusUnauthorized, "Unauthorized", errors.New("unauthorized user"))
		return
	}

	db := database.API_USER_DB

	// Set RLS for the current user
	err = database.SetRLS(db, user_id)
	if err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Internal server error", err)
		return
	}

	// Store user details in request context
	ctx := context.WithValue(request.Context(), context_keys.USER_ID, user_id)
	ctx = context.WithValue(ctx, context_keys.USER_TYPE, user_type)
	ctx = context.WithValue(ctx, context_keys.DB, db)

	next.ServeHTTP(response, request.WithContext(ctx))
}

// Verifies JWT token expiration
func VerifyJwtTokenExpiration(request *http.Request, token string) (string, error) {
	tokenAuth := jwtauth.New("HS256", []byte(config.Config("JWT_SECRET")), nil)
	payload, err := jwtauth.VerifyToken(tokenAuth, token)

	if err != nil {
		return "", err
	}

	return payload.Subject(), nil
}

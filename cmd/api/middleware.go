package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Pedro-Foramilio/social/internal/store"
	"github.com/golang-jwt/jwt/v5"
)

func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("missing authorization header"))
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Basic" {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("invalid authorization header"))
				return
			}

			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("invalid base64 encoding"))
				return
			}

			username := app.config.auth.basic.user
			pass := app.config.auth.basic.pass
			credentials := strings.SplitN(string(decoded), ":", 2)
			if len(credentials) != 2 || credentials[0] != username || credentials[1] != pass {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("invalid authorization credentials"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("missing authorization header"))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("invalid authorization header"))
			return
		}

		token := parts[1]

		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("invalid token: %w", err))
			return
		}

		claims, _ := jwtToken.Claims.(jwt.MapClaims)

		userID, err := strconv.ParseInt(fmt.Sprintf("%v", claims["sub"]), 10, 64)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("invalid user ID in token: %w", err))
			return
		}

		ctx := r.Context()
		user, err := app.store.Users.GetByID(ctx, userID)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("user not found: %w", err))
			return
		}

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) checkPostOwnership(role string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := getUserFromContext(r)
		post := getPostsFromCtx(r)

		if post.UserID == user.ID {
			next.ServeHTTP(w, r)
			return
		}

		allowed, err := app.checkRolePrecedence(r.Context(), user, role)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if !allowed {
			app.forbiddenErrorResponse(w, r, fmt.Errorf("insufficient permissions to modify this resource"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) checkRolePrecedence(context context.Context, user *store.User, roleName string) (bool, error) {
	role, err := app.store.Roles.GetByName(context, roleName)
	if err != nil {
		return false, err
	}

	return user.Role.Level >= role.Level, nil
}

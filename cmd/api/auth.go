package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/Pedro-Foramilio/social/internal/mailer"
	"github.com/Pedro-Foramilio/social/internal/store"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type UserWithToken struct {
	User  *store.User `json:"user"`
	Token string      `json:"token"`
}

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	token := uuid.New().String()
	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])
	err := app.store.Users.CreateAndInvite(r.Context(), user, hashToken, app.config.mail.exp)

	if err != nil {
		switch {
		case err == store.ErrDuplicateEmail:
			app.badRequestResponse(w, r, err)
		case err == store.ErrDuplicateUsername:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	userWithToken := UserWithToken{
		User:  user,
		Token: token,
	}

	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, token),
	}
	res, err := app.mailer.Send(
		mailer.UserWelcomeTemplate,
		user.Username,
		user.Email,
		vars,
		app.config.env != "production")

	if err != nil {

		app.logger.Errorw("Error sending welcome email", "error", err, "response", res)

		if err := app.store.Users.Delete(r.Context(), user.ID); err != nil {
			app.logger.Errorw("Error deleting user after failed email", "error", err)
		}

		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, userWithToken); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

type CreateUserTokenPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

func (app *application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateUserTokenPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)

	if err != nil {
		if err == store.ErrNotFound {
			app.badRequestResponse(w, r, fmt.Errorf("invalid credentials"))
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.config.auth.token.exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.config.auth.token.iss,
		"aud": app.config.auth.token.iss,
	}

	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, token); err != nil {
		app.internalServerError(w, r, err)
	}

}

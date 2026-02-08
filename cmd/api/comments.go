package main

import (
	"net/http"

	"github.com/Pedro-Foramilio/social/internal/store"
)

type CreateCommentPayload struct {
	PostID  int64  `json:"post_id" validate:"required"`
	Content string `json:"content" validate:"required,max=500"`
}

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateCommentPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if validationErr := Validate.Struct(payload); validationErr != nil {
		app.badRequestResponse(w, r, validationErr)
		return
	}

	user := getUserFromContext(r)
	comment := &store.Comment{
		PostID:  payload.PostID,
		Content: payload.Content,
		UserID:  user.ID,
	}

	if err := app.store.Comments.Create(r.Context(), comment); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, comment); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

package main

import "net/http"

func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	feed, err := app.store.Posts.GetUserFeed(ctx, int64(102))
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	err = app.jsonResponse(w, http.StatusOK, feed)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

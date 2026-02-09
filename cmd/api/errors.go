package main

import (
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {

	app.logger.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err)

	writeJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {

	app.logger.Warnw("bad request error", "method", r.Method, "path", r.URL.Path, "error", err)

	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {

	app.logger.Warnw("not found error", "method", r.Method, "path", r.URL.Path, "error", err)

	writeJSONError(w, http.StatusNotFound, "resource not found")
}

func (app *application) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {

	app.logger.Errorw("conflict error", "method", r.Method, "path", r.URL.Path, "error", err)

	writeJSONError(w, http.StatusConflict, err.Error())
}

func (app *application) unauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {

	app.logger.Warnw("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err)

	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

	writeJSONError(w, http.StatusUnauthorized, "unauthorized")
}

func (app *application) unauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {

	app.logger.Warnw("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err)

	writeJSONError(w, http.StatusUnauthorized, "unauthorized")
}

func (app *application) forbiddenErrorResponse(w http.ResponseWriter, r *http.Request, err error) {

	app.logger.Warnw("forbidden error", "method", r.Method, "path", r.URL.Path, "error", err)

	writeJSONError(w, http.StatusForbidden, "forbidden")
}

func (app *application) rateLimiterExceededResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {

	app.logger.Warnw("rate limit exceeded", "method", r.Method, "path", r.URL.Path, "retryAfter", retryAfter)
	w.Header().Set("Retry-After", retryAfter)
	writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
}

package main

import (
	"context"
	"net/http"

	"huytran2000-hcmus/greenlight/internal/data"
)

type contextKey string

const userCtxKey = contextKey("user")

func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, userCtxKey, user)

	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) *data.User {
	ctx := r.Context()
	user, ok := ctx.Value(userCtxKey).(*data.User)
	if !ok {
		panic("missing user in the request context")
	}

	return user
}

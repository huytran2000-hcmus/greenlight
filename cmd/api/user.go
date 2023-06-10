package main

import (
	"errors"
	"net/http"
	"time"

	"huytran2000-hcmus/greenlight/internal/data"
	"huytran2000-hcmus/greenlight/internal/validator"
)

const defaultActivationTimeout = 3 * 24 * time.Hour

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Email: input.Email,
		Name:  input.Name,
	}
	user.Password.PlainText = &input.Password

	v := validator.New()
	data.ValidateUser(v, user)
	if !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.User.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddFieldError("email", "a user with this email already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	token, err := app.models.Token.New(user.ID, defaultActivationTimeout, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	expiredIn := time.Until(token.Expiry).Round(time.Hour)

	app.background(func() {
		data := map[string]interface{}{
			"userID":          user.ID,
			"activationToken": token.Plaintext,
			"expiredIn":       fmtDuration(expiredIn),
		}

		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.Error(err, nil)
		}
	})

	err = app.writeJSON(w, http.StatusAccepted, nil, envelope{"user": user})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

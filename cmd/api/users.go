package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"huytran2000-hcmus/greenlight/internal/data"
	"huytran2000-hcmus/greenlight/internal/validator"
)

const (
	defaultActivationTimeout     = 3 * 24 * time.Hour
	defaultAuthenticationTimeout = 3 * time.Hour
)

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

	v := validator.New()
	err = user.Password.Set(input.Password)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrPasswordTooLong):
			v.AddFieldError("password", fmt.Sprintf("must not be more than %d character", data.MaxPasswordLen))
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

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

	err = app.models.Permission.AddForUser(user.ID, "movies:read")
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	token, err := app.models.Token.New(data.ScopeActivation, user.ID, defaultActivationTimeout)
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

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateTokenPlainText(v, input.TokenPlaintext)
	if !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.User.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddFieldError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	user.Activated = true
	err = app.models.User.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	err = app.models.Token.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, nil, envelope{"user": user})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlainText(v, input.Password)

	if !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.User.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	matched, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !matched {
		app.invalidCredentialsResponse(w, r)
		return
	}

	token, err := app.models.Token.New(data.ScopeAuthentication, user.ID, defaultAuthenticationTimeout)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, nil, envelope{"authentication": token})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

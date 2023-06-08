package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"

	"huytran2000-hcmus/greenlight/internal/validator"
)

type envelope map[string]any

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	rawID := params.ByName("id")
	id, err := strconv.ParseInt(rawID, 10, 64)

	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id %d", id)
	}

	return id, nil
}

func (app *application) readString(qs url.Values, key, defaultVal string) string {
	val := qs.Get(key)

	if val == "" {
		return defaultVal
	}

	return val
}

func (app *application) readCSV(qs url.Values, key string, defaultVals []string) []string {
	vals := qs.Get(key)

	if vals == "" {
		return defaultVals
	}

	return strings.Split(vals, ",")
}

func (app *application) readInt(qs url.Values, key string, defaultVal int, v *validator.Validator) int {
	val := qs.Get(key)

	if val == "" {
		return defaultVal
	}

	i, err := strconv.Atoi(val)
	if err != nil {
		v.AddFieldError(key, "must be an integer value")
		return defaultVal
	}

	return i
}

func (app *application) writeJSON(w http.ResponseWriter, status int, header http.Header, data envelope) error {
	resBody, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resBody = append(resBody, '\n')

	for key, val := range header {
		w.Header()[key] = val
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(resBody)
	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	var maxBytes int64 = 1_048_576
	body := http.MaxBytesReader(w, r.Body, maxBytes)
	dec := json.NewDecoder(body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		var invalidErr *json.InvalidUnmarshalError
		var maxBytesErr *http.MaxBytesError
		switch {
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("request body contains badly-formed JSON(at character %d)", syntaxErr.Offset)
		case errors.As(err, &typeErr):
			if typeErr.Struct != "" || typeErr.Field != "" {
				return fmt.Errorf("request body contains incorrect JSON type for field %q", typeErr.Field)
			}
			return fmt.Errorf("request body contains badly-formed JSON(at character %d)", typeErr.Offset)
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("request body contains unknown field %s", fieldName)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("request body contains badly-formed JSON")
		case errors.Is(err, io.EOF):
			return errors.New("request body is empty")
		case errors.As(err, &maxBytesErr):
			return fmt.Errorf("request body must not be larger than %d bytes", maxBytesErr.Limit)
		case errors.As(err, &invalidErr):
			panic(err)
		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("request body must only contain a single JSON value")
	}

	return nil
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) background(fn func()) {
	go func() {
		defer func() {
			err := recover()
			if err != nil {
				app.logger.Error(fmt.Errorf("recover: %s", err), nil)
			}
		}()

		fn()
	}()
}

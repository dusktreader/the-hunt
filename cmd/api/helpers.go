package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/dusktreader/the-hunt/internal/data"
	"github.com/dusktreader/the-hunt/internal/validator"
	"github.com/julienschmidt/httprouter"
)

func MaybeDie(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "There was an error:", err)
		os.Exit(1)
	}
}

func Die(msg string, flags ...interface{}) {
	msg = fmt.Sprintf(msg, flags...)
	fmt.Fprintln(os.Stderr, "Aborting:", msg)
	os.Exit(1)
}

func (app *application) logError(r *http.Request, er *data.ErrorPackage) {
	var logMessage string
	if er.LogMessage == "" {
		logMessage = fmt.Sprintf("Error encountered: %s", er.Message)
	} else {
		logMessage = er.LogMessage
	}
	app.logger.Error(
		logMessage,
		"error", er.Error,
		"method", r.Method,
		"uri", r.URL.RequestURI(),
	)

}

func (app *application) parseIdParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Int is required")
	} else if id < 0 {
		return 0, fmt.Errorf("Negative ids are not allowed")
	} else if id == 0 {
		return 0, fmt.Errorf("0 is not allowed")
	}
	return id, nil
}

func (app *application) readString(
	qs url.Values,
	key string,
	_ *validator.Validator,
) *string {
	var p *string
	s := qs.Get(key)
	if s != "" {
		app.logger.Debug("Read string", "key", key, "value", s)
		p = &s
	}
	return p
}

func (app *application) readCSV(
	qs url.Values,
	key string,
	_ *validator.Validator,
) []string {
	s := make([]string, 0)
	csv := qs.Get(key)
	if csv != "" {
		s = strings.Split(csv, ",")
	}
	return s
}

func (app *application) readInt(
	qs url.Values,
	key string,
	v *validator.Validator,
) *int {
	var p *int
	s := qs.Get(key)
	if s != "" {
		i, err := strconv.Atoi(s)
		if err != nil {
			v.AddError("query", "must be an integer")
		} else {
			p = &i
		}
	}
	return p
}

func (app *application) readBool(
	qs url.Values,
	key string,
	v *validator.Validator,
) *bool {
	var p *bool
	s := qs.Get(key)
	if s != "" {
		sl := strings.ToLower(s)
		if slices.Contains([]string{"t", "true", "y", "yes", "1"}, sl) {
			b := true
			p = &b
		} else if slices.Contains([]string{"f", "false", "n", "no", "0"}, sl) {
			b := false
			p = &b
		} else {
			v.AddError("query", fmt.Sprintf("could not map %q to a boolean value", s))
		}
	}
	return p
}

func (app *application) ParseFilters(
	qs url.Values,
	v *validator.Validator,
	c data.FilterConstraints,
) data.Filters {
	f := data.Filters{
		Search:		&data.SearchMap{},
		In:			&data.InMap{},
		Page:		app.readInt(qs, "page", v),
		PageSize:	app.readInt(qs, "page_size", v),
		Sort:		app.readString(qs, "sort", v),
		SortAsc: 	app.readBool(qs, "sort_asc", v),
	}

	rex := regexp.MustCompile(`^search_(\w+)$`)
	for key := range qs {
		m := rex.FindStringSubmatch(key)
		if len(m) == 2 {
			partKey := m[1]
			partVal := *app.readString(qs, key, v)
			if len(partVal) < 3 {
				v.AddError(key, "Search parameters must be at least 3 characters long")
			} else {
				(*f.Search)[partKey] = partVal
				app.logger.Debug("Added search parameter", "key", partKey, "value", partVal)
			}
		}
	}

	rex = regexp.MustCompile(`^in_(\w+)$`)
	for key := range qs {
		m := rex.FindStringSubmatch(key)
		if len(m) == 2 {
			partKey := m[1]
			partVal := *app.readString(qs, key, v)
			(*f.In)[partKey] = partVal
			app.logger.Debug("Added in parameter", "key", partKey, "value", partVal)
		}
	}

	if f.Search != nil && c.Search != nil {
		v.Check(c.Search(*f.Search), "search", "parameter is invalid")
	}
	if f.Sort != nil && c.Sort != nil {
		v.Check(c.Sort(*f.Sort), "sort", "parameter is invalid")
	}
	if f.In != nil && c.In != nil {
		v.Check(c.In(*f.In), "in", "parameter is invalid")
	}
	if f.Page != nil && c.Page != nil {
		v.Check(c.Page(*f.Page), "page", "parameter is invalid")
	}
	if f.PageSize != nil && c.PageSize != nil {
		v.Check(c.PageSize(*f.PageSize), "page_size", "parameter is invalid")
	}

	app.logger.Debug("Parsed filters", "filters", f)

	return f
}

func (app *application) writeJSON(w http.ResponseWriter, jr *data.JSONResponse) error {
	key := "data"
	if jr.EnvelopeKey != "" {
		key = jr.EnvelopeKey
	}
	env := make(data.Envelope)
	env[key] = jr.Data

	var serialized []byte
	var err error
	if app.config.APIEnv == "development" {
		serialized, err = json.MarshalIndent(env, "", "  ")
	} else {
		serialized, err = json.Marshal(env)
	}
	if err != nil {
		return fmt.Errorf("Failed to serialize response data: %w", err)
	}
	serialized = append(serialized, '\n')

	w.Header().Set("Content-Type", "application/json")
	for key, val := range jr.Headers {
		w.Header()[key] = val
	}

	w.WriteHeader(jr.StatusCode)
	w.Write(serialized)
	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {

		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
			case errors.As(err, &syntaxError):
				return fmt.Errorf("Body contains badly-formed JSON (at character %d)", syntaxError.Offset)

			case errors.Is(err, io.ErrUnexpectedEOF):
				return fmt.Errorf("Body contains badly-formed JSON")

			case errors.As(err, &unmarshalTypeError):
				if unmarshalTypeError.Field != "" {
					return fmt.Errorf("Body contains an incorrect JSON type for the %q field", unmarshalTypeError.Field)
				}
				return fmt.Errorf("Body contains an incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

			case errors.Is(err, io.EOF):
				return fmt.Errorf("Body must not be empty")

			case errors.As(err, &invalidUnmarshalError):
				panic(err)

			default:
				return err
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return fmt.Errorf("Body must only contain a single JSON value")
	}
	return nil
}

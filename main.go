package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-msvc/errors"
	"github.com/gorilla/mux"
	"github.com/jansemmelink/events/db"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/auth/exists", auth(authGetExists)).Methods(http.MethodGet)
	r.HandleFunc("/auth/register", auth(authPostRegister)).Methods(http.MethodPost)
	r.HandleFunc("/auth/activate", auth(authPostActivate)).Methods(http.MethodPost)
	r.HandleFunc("/auth/reset", auth(authPostReset)).Methods(http.MethodPost)
	r.HandleFunc("/auth/login", auth(authPostLogin)).Methods(http.MethodPost)
	r.HandleFunc("/validate/password", auth(authPostValidatePassword)).Methods(http.MethodPost)
	r.HandleFunc("/events", auth(getEventsList)).Methods(http.MethodGet)
	r.HandleFunc("/event/{id}", auth(getEventDetails)).Methods(http.MethodGet)
	http.Handle("/", CORS(r))
	http.ListenAndServe(":12345", nil)
}

type CtxParams struct{}

func CORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		fmt.Printf("HTTP %s %s (origin:%s)\n", r.Method, r.URL.Path, origin)
		w.Header().Set("Access-Control-Allow-Origin", origin)
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token, Authorization")
			return
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

func auth(fnc interface{}) func(http.ResponseWriter, *http.Request) {
	fncType := reflect.TypeOf(fnc)
	fncValue := reflect.ValueOf(fnc)
	var reqType reflect.Type
	if fncType.NumIn() > 1 {
		reqType = fncType.In(1)
	}

	type ErrorResponse struct {
		Error string `json:"error"`
	}

	return func(httpRes http.ResponseWriter, httpReq *http.Request) {
		ctx := context.Background()
		var err error
		var res interface{}
		defer func() {
			httpRes.Header().Set("Content-Type", "application/json")
			if err != nil {
				fmt.Printf("ERROR: %+v\n", err)

				//in response - only log the base error
				for {
					if baseErr, ok := err.(errors.IError); ok && baseErr.Parent() != nil {
						err = baseErr.Parent()
					} else {
						break
					}
				}
				res := ErrorResponse{Error: fmt.Sprintf("%+s", err)}
				jsonRes, _ := json.Marshal(res)
				httpRes.Write(jsonRes)
				return
			}
			if res != nil {
				if err = json.NewEncoder(httpRes).Encode(res); err != nil {
					http.Error(httpRes, fmt.Sprintf("failed to encode response: %+s", err), http.StatusInternalServerError)
					return
				}
			}
		}()

		params := map[string]string{}
		for n, v := range httpReq.URL.Query() {
			params[n] = strings.Join(v, ",")
		}
		vars := mux.Vars(httpReq)
		for n, v := range vars {
			params[n] = v
		}
		ctx = context.WithValue(ctx, CtxParams{}, params)

		//prepare fnc arguments
		args := []reflect.Value{reflect.ValueOf(ctx)}

		if fncType.NumIn() > 1 {
			ct := httpReq.Header.Get("Content-Type")
			if ct != "" && ct != "application/json" {
				err = errors.Errorc(http.StatusBadRequest, fmt.Sprintf("invalid Content-Type: %+s, expecting application/json", ct))
				return
			}

			reqValuePtr := reflect.New(reqType)
			if err = json.NewDecoder(httpReq.Body).Decode(reqValuePtr.Interface()); err != nil {
				err = errors.Errorc(http.StatusBadRequest, fmt.Sprintf("cannot parse JSON body: %+s", err))
				return
			}

			if validator, ok := reqValuePtr.Interface().(Validator); ok {
				if err = validator.Validate(); err != nil {
					err = errors.Errorc(http.StatusBadRequest, fmt.Sprintf("invalid request body: %+s", err))
					return
				}
			}
			args = append(args, reqValuePtr.Elem())
		}

		results := fncValue.Call(args)

		errValue := results[len(results)-1] //last result is error
		if !errValue.IsNil() {
			err = errors.Wrapf(errValue.Interface().(error), "handler failed")
			return
		}

		if fncType.NumOut() > 1 {
			if results[0].IsValid() {
				if results[0].Type().Kind() == reflect.Ptr && !results[0].IsNil() {
					res = results[0].Elem().Interface() //dereference the pointer
				} else {
					res = results[0].Interface()
				}
			}
		}
	}
}

//handler to check if nat_id/phone/email exists in DB to invalidate a registration form
//beforte it is submitted, without revealing any personal information, i.e. no auth required
//on this endpoint
func authGetExists(ctx context.Context) error {
	params := ctx.Value(CtxParams{}).(map[string]string)
	if len(params) == 0 {
		return errors.Errorc(http.StatusBadRequest, "missing URL param nat_id, phone or email")
	}
	if len(params) > 1 {
		return errors.Errorc(http.StatusBadRequest, "only one URL param allowed: nat_id, phone or email")
	}
	if _, err := db.GetPerson(params); err != nil {
		return errors.Errorc(http.StatusNotFound, "not found")
	}
	return nil
} //authGetExists

func authPostRegister(ctx context.Context, req db.RegisterRequest) (string, error) {
	personId, err := db.AddPerson(req)
	if err != nil {
		return "", errors.Wrapf(err, "failed to register")
	}
	return personId, nil
}

func authPostActivate(ctx context.Context, req db.AuthActivateRequest) error {
	err := db.AuthActivate(req)
	if err != nil {
		return errors.Wrapf(err, "failed to activate")
	}
	return nil
}

func authPostReset(ctx context.Context, req db.AuthResetRequest) error {
	err := db.AuthReset(req)
	if err != nil {
		return errors.Wrapf(err, "failed to reset")
	}
	return nil
}

func authPostLogin(ctx context.Context, req db.LoginRequest) (interface{}, error) {
	person, err := db.Login(req)
	if err != nil {
		return nil, errors.Wrapf(err, "login failed")
	}
	return person, nil
}

type ValidationRequest struct {
	Value string `json:"value"`
}

type ValidationResponse struct {
	Valid   bool   `json:"valid"`
	Details string `json:"details"`
}

func authPostValidatePassword(ctx context.Context, req ValidationRequest) (ValidationResponse, error) {
	err := db.ValidatePassword(req.Value)
	if err != nil {
		return ValidationResponse{Valid: false, Details: fmt.Sprintf("invalid password: %s", err)}, nil
	}
	return ValidationResponse{Valid: true}, nil
}

func getEventsList(ctx context.Context) (interface{}, error) {
	params := ctx.Value(CtxParams{}).(map[string]string)
	filter := params["filter"]
	events, err := db.ListEvents(filter)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func getEventDetails(ctx context.Context) (interface{}, error) {
	params := ctx.Value(CtxParams{}).(map[string]string)
	id := params["id"]
	events, err := db.GetEvent(id)
	if err != nil {
		return nil, err
	}
	return events, nil
}

type Validator interface {
	Validate() error
}

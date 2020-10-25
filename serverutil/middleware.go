package serverutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	cuserr "github.com/karta0807913/lab_server/error"
)

type MiddlewareInterface interface {
	Handle(req *http.Request, body interface{}) error
}

type Middleware struct {
	MiddlewareInterface
	err_msg string
}

func (self Middleware) Handle(req *http.Request, body interface{}) error {
	return errors.New("not a middleware")
}

type MethodCheck struct {
	Middleware
	Method string
}

type JsonBodyParser struct {
	Middleware
}

type BodyCheck struct {
	Middleware
}

func (self MethodCheck) Handle(req *http.Request, body interface{}) error {
	if req.Method != self.Method {
		return errors.New(self.err_msg)
	}
	return nil
}

func (self JsonBodyParser) Handle(req *http.Request, body interface{}) error {
	if body == nil {
		return nil
	}
	decoder := json.NewDecoder(req.Body)

	decoder.DisallowUnknownFields()

	if err := decoder.Decode(body); err != nil {
		return new(cuserr.IsNotJsonError)
	}
	return nil
}

func (self BodyCheck) Handle(req *http.Request, body interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(body))

	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).IsNil() {
			t := v.Type()
			return &cuserr.UserInputError{
				ErrMsg: fmt.Sprintf("key %s missing", t.Field(i).Name),
			}
		}
	}
	return nil
}

func MiddlewareCheckBuilder(middlewareList ...MiddlewareInterface) func(req *http.Request, body interface{}) error {
	return func(req *http.Request, body interface{}) error {
		for _, middleware := range middlewareList {
			if err := middleware.Handle(req, body); err != nil {
				return err
			}
		}
		return nil
	}
}

package server

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/armosec/ca-test/utils"
)

type RequestHandler func(w http.ResponseWriter, r *http.Request, reqBody string)

type RequestHandlerOption func(opts *requestHandlerOptions) error

//WithMethod option sets the method for the handler
var WithMethod = func(method string) RequestHandlerOption {
	return func(o *requestHandlerOptions) error {
		o.method = method
		return nil
	}
}

//WithPath option sets the path for the handler
var WithPath = func(path string) RequestHandlerOption {
	return func(o *requestHandlerOptions) error {
		if path != "" && (o.pathPrefix != "" || o.pathSuffix != "") {
			return fmt.Errorf("path can't be set with path prefix or suffix")
		}
		o.path = path
		return nil
	}
}

//WithPathPrefix option sets the path prefix for the handler
var WithPathPrefix = func(pathPrefix string) RequestHandlerOption {
	return func(o *requestHandlerOptions) error {
		if pathPrefix != "" && o.path != "" {
			return fmt.Errorf("path and path prefix can't be set together")
		}
		o.pathPrefix = pathPrefix
		return nil
	}
}

//WithPathSuffix option sets the path suffix for the handler
var WithPathSuffix = func(pathSuffix string) RequestHandlerOption {
	return func(o *requestHandlerOptions) error {
		if pathSuffix != "" && o.path != "" {
			return fmt.Errorf("path and path suffix can't be set together")
		}
		o.pathSuffix = pathSuffix
		return nil
	}
}

//WithResponse option sets the response for the handler
var WithResponse = func(response []byte) RequestHandlerOption {
	return func(o *requestHandlerOptions) error {
		if len(response) != 0 && o.handler != nil {
			return fmt.Errorf("response can't be set with handler")
		}
		o.response = response
		return nil
	}
}

//WithHandler option sets the handler for the handler
var WithHandler = func(handler RequestHandler) RequestHandlerOption {
	return func(o *requestHandlerOptions) error {
		if handler != nil && len(o.response) != 0 {
			return fmt.Errorf("handler can't be set with response")
		}
		o.handler = handler
		return nil
	}
}

//WithRequestNumber option sets the request number for the handler
var WithRequestNumber = func(reqNum int) RequestHandlerOption {
	return func(o *requestHandlerOptions) error {
		o.reqNum = reqNum
		return nil
	}
}

//WithUpdateExpected option sets the update expected flag for the handler
var WithTestRequest = func(t *testing.T, updateExpected bool, expectedRequest []byte, expectedRequestFile string) RequestHandlerOption {
	return func(o *requestHandlerOptions) error {
		if expectedRequest == nil || expectedRequestFile == "" || t == nil {
			return fmt.Errorf("test, expected request and expected request file must be provided")
		}
		o.t = t
		o.updateExpected = updateExpected
		o.expectedRequest = expectedRequest
		o.expectedRequestFile = expectedRequestFile
		return nil
	}
}

type requestHandlerOptions struct {
	method              string
	path                string
	response            []byte
	expectedRequest     []byte
	expectedRequestFile string
	updateExpected      bool
	reqNum              int
	pathPrefix          string
	pathSuffix          string
	handler             RequestHandler
	t                   *testing.T
}

func (o *requestHandlerOptions) validate() error {
	if o.t == nil {
		if o.updateExpected {
			return fmt.Errorf("test is required for update expected")
		}
	}
	return nil
}

func (o *requestHandlerOptions) getOrCreateHandler() RequestHandler {
	if o.handler != nil {
		return o.handler
	}
	return func(w http.ResponseWriter, r *http.Request, reqBody string) {
		if len(o.expectedRequest) != 0 {
			utils.DeepEqualOrUpdate(o.t, []byte(reqBody), o.expectedRequest, o.expectedRequestFile, o.updateExpected)
		}
		if len(o.response) != 0 {
			w.Write(o.response)
		}
	}
}

func makeRequestHandlerOptions(opts ...RequestHandlerOption) (*requestHandlerOptions, error) {
	o := &requestHandlerOptions{}
	for _, option := range opts {
		if err := option(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}

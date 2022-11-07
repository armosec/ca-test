package server

import (
	"fmt"
	"os"
)

type ServerOption func(opts *serverOptions, isUpdate bool) error

//Options

//WithBuiltInHandler option adds built in handler to the server.
//Built-in handlers are fixed and will not be removed when ResetHandlers is called
var WithBuiltInHandler = func(opts ...RequestHandlerOption) ServerOption {
	return func(o *serverOptions, isUpdate bool) error {
		if handler, err := newRequestHandler(opts...); err != nil {
			return err
		} else {
			o.defaultRequestHandlers = append(o.defaultRequestHandlers, *handler)
		}
		return nil
	}
}

//WithHeaders option adds headers to each response
var WithHeaders = func(headers map[string]string) ServerOption {
	return func(o *serverOptions, isUpdate bool) error {
		for k, v := range headers {
			o.headers[k] = v
		}
		return nil
	}
}

//WithPort option sets port for the server, if not specified the server will pick the next available port
//This option cannot be changed after the server is created
var WithPort = func(port int) ServerOption {
	return func(o *serverOptions, isUpdate bool) error {
		if isUpdate {
			return fmt.Errorf("port can't be updated")
		}
		o.port = port
		return nil
	}
}

//WithTLS option enables TLS for the server
//This option cannot be changed after the server is created
var WithTLS = func() ServerOption {
	return func(o *serverOptions, isUpdate bool) error {
		if isUpdate {
			return fmt.Errorf("tls option can't be updated")
		}
		o.tls = true
		return nil
	}
}

//WithRecord option enables recording of requests to the server to a specific folder and after a specific number of requests
var WithRequestsRecorder = func(record bool, recordsFolder string, recordAfterReqNum int, recordOnlyUnhandled bool) ServerOption {
	return func(o *serverOptions, isUpdate bool) error {
		if record {
			if recordsFolder == "" {
				return fmt.Errorf("recording folder must not be empty if recode is set to true")
			}
			if err := os.MkdirAll(recordsFolder, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create recordings directory %s error: %v", recordsFolder, err)
			}
		}
		o.recordFolder = recordsFolder
		o.record = record
		o.recordAfterReqNum = recordAfterReqNum
		o.recordOnlyUnhandled = recordOnlyUnhandled
		return nil
	}
}

//Options for test server
type serverOptions struct {
	port                   int
	tls                    bool //
	recordFolder           string
	recordAfterReqNum      int
	record                 bool
	recordOnlyUnhandled    bool
	headers                map[string]string
	defaultRequestHandlers []serverRequestHandler
}

func makeServerOptions(opts ...ServerOption) (*serverOptions, error) {
	o := &serverOptions{
		port:                   0,
		tls:                    false,
		defaultRequestHandlers: []serverRequestHandler{},
		headers:                map[string]string{},
		record:                 false,
		recordFolder:           "",
		recordAfterReqNum:      0,
	}
	return applyOptions(o, false, opts...)
}

func applyOptions(o *serverOptions, isUpdate bool, opts ...ServerOption) (*serverOptions, error) {
	for _, option := range opts {
		if err := option(o, isUpdate); err != nil {
			return nil, err
		}
	}
	return o, nil
}

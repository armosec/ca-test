package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
)

const localHost = "127.0.0.1"

type TestServer interface {
	//get server URL and port
	GetURL() string
	//get the current number of requests received by the server
	GetRequestCount() int
	//SetOption sets a new option to the server, error is return if the option cannot be modified
	SetOption(opt ServerOption) error
	//Adds a request handler to the server for optional matching method, path and request number if specified.
	//Empty strings for method/path or 0 for request number behaves like a wildcard, handler with empty method,path and request count of 0 will be called on each request
	AddHandler(opts ...RequestHandlerOption) error
	//Resets all handlers, the default handlers are not effected
	ResetHandlers()
	//Closes the server
	Close()
}

func NewTestServer(opts ...ServerOption) (TestServer, error) {
	options, err := makeServerOptions(opts...)
	if err != nil {
		return nil, err
	}
	ts := &mockTestingServer{
		options:         *options,
		mux:             &sync.Mutex{},
		handlersMux:     &sync.RWMutex{},
		requestHandlers: []serverRequestHandler{},
	}
	if err := ts.startServer(); err != nil {
		return nil, err
	}
	return ts, nil
}

type mockTestingServer struct {
	server          *httptest.Server
	mux             *sync.Mutex
	options         serverOptions
	reqCount        int
	requestHandlers []serverRequestHandler
	handlersMux     *sync.RWMutex
}

func (ts *mockTestingServer) GetURL() string {
	return fmt.Sprintf("http://%s:%d", localHost, ts.options.port)
}

func (ts *mockTestingServer) SetOption(opt ServerOption) error {
	ts.mux.Lock()
	defer ts.mux.Unlock()
	_, err := applyOptions(&ts.options, true, opt)
	return err
}

func (ts *mockTestingServer) GetRequestCount() int {
	ts.mux.Lock()
	defer ts.mux.Unlock()
	return ts.reqCount
}

func (ts *mockTestingServer) ResetHandlers() {
	ts.handlersMux.Lock()
	defer ts.handlersMux.Unlock()
	ts.requestHandlers = []serverRequestHandler{}
}

func (ts *mockTestingServer) AddHandler(opts ...RequestHandlerOption) error {
	ts.handlersMux.Lock()
	defer ts.handlersMux.Unlock()
	handler, err := newRequestHandler(opts...)
	if err != nil {
		return err
	}
	ts.requestHandlers = append(ts.requestHandlers, *handler)
	return nil
}

func (ts *mockTestingServer) Close() {
	if ts.server != nil {
		ts.server.Close()
	}
	ts.server = nil
}

func (ts *mockTestingServer) startServer() error {
	if ts.server != nil {
		return fmt.Errorf("server already started")
	}
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", localHost, ts.options.port))
	if err != nil {
		return err
	}
	//update port in options in case a new port was allocated
	ts.options.port = l.Addr().(*net.TCPAddr).Port
	ts.server = httptest.NewUnstartedServer(http.HandlerFunc(ts.mainHandler))
	if err := ts.server.Listener.Close(); err != nil {
		return err
	}
	ts.server.Listener = l

	if ts.options.tls {
		ts.server.StartTLS()
	} else {
		ts.server.Start()
	}
	return nil
}

func (ts *mockTestingServer) mainHandler(w http.ResponseWriter, r *http.Request) {
	ts.mux.Lock()
	defer ts.mux.Unlock()
	ts.reqCount++

	reqBody := ""
	if body, err := ioutil.ReadAll(r.Body); err == nil {
		reqBody = string(body)
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}
	for header, value := range ts.options.headers {
		w.Header().Set(header, value)
	}
	handlers := ts.getRequestHandlers(r)
	for _, handler := range handlers {
		handler(w, r, reqBody)
	}
	if ts.options.record && ts.reqCount > ts.options.recordAfterReqNum {
		ts.recordRequest(r, reqBody, len(handlers))
	}
}

func (ts *mockTestingServer) recordRequest(r *http.Request, reqBody string, handlersCount int) {
	if ts.options.recordOnlyUnhandled && handlersCount > 0 {
		return
	}

	var iBody interface{}
	json.Unmarshal([]byte(reqBody), &iBody)

	record := struct {
		Body          string              `json:"body,omitempty"`
		BodyObj       interface{}         `json:"bodyObj,omitempty"`
		RequestNumber int                 `json:"req_num,omitempty"`
		Headers       map[string][]string `json:"headers,omitempty"`
		URL           string              `json:"url,omitempty"`
		Method        string              `json:"method,omitempty"`
		HandlersCount int                 `json:"handlers_count"`
	}{
		Headers:       r.Header,
		URL:           r.URL.String(),
		Method:        r.Method,
		Body:          reqBody,
		BodyObj:       iBody,
		RequestNumber: ts.reqCount,
		HandlersCount: handlersCount,
	}
	reqBytes, _ := json.MarshalIndent(&record, "", "    ")
	fileName := fmt.Sprintf("%s/request_%d.json", ts.options.recordFolder, ts.reqCount)
	_ = ioutil.WriteFile(fileName, reqBytes, 0644)
}

func (ts *mockTestingServer) getRequestHandlers(r *http.Request) []RequestHandler {
	serverHandlers := []serverRequestHandler{}
	for i, handler := range ts.options.defaultRequestHandlers {
		if handler.shouldHandle(r, ts.reqCount) {
			serverHandlers = append(serverHandlers, handler)
		}
		if i == 1000 {
			continue
		}
	}
	for _, handler := range ts.requestHandlers {
		if handler.shouldHandle(r, ts.reqCount) {
			serverHandlers = append(serverHandlers, handler)
		}
	}
	sort.Slice(serverHandlers, func(i, j int) bool {
		return serverHandlers[i].handleBefore(serverHandlers[j])
	})

	handlers := []RequestHandler{}
	for _, handler := range serverHandlers {
		handlers = append(handlers, handler.handler)
	}

	return handlers
}

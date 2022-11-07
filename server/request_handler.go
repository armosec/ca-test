package server

import (
	"net/http"
	"strings"
)

type serverRequestHandler struct {
	options *requestHandlerOptions
	handler RequestHandler
}

func newRequestHandler(opts ...RequestHandlerOption) (*serverRequestHandler, error) {
	options, err := makeRequestHandlerOptions(opts...)
	if err != nil {
		return nil, err
	}
	if err := options.validate(); err != nil {
		return nil, err
	}

	return &serverRequestHandler{
		options: options,
		handler: options.getOrCreateHandler(),
	}, nil
}

func (h *serverRequestHandler) shouldHandle(r *http.Request, reqCount int) bool {
	if h.options.method != "" && h.options.method != r.Method {
		return false
	}
	if h.options.path != "" && h.options.path != r.URL.Path {
		return false
	}
	if h.options.pathPrefix != "" && !strings.HasPrefix(r.URL.Path, h.options.pathPrefix) {
		return false
	}
	if h.options.pathSuffix != "" && !strings.HasSuffix(r.URL.Path, h.options.pathSuffix) {
		return false
	}
	if h.options.reqNum != 0 && h.options.reqNum != reqCount {
		return false
	}
	//if handler is configured with responses array and served all responses, return false
	if h.options.responses != nil && len(h.options.responses) == 0 {
		return false
	}
	return true
}

func (h *serverRequestHandler) handleBefore(other serverRequestHandler) bool {
	if h.options.reqNum == 0 && other.options.reqNum != 0 {
		return true
	}
	if h.options.method == "" && other.options.method != "" {
		return true
	}
	if h.options.path == "" && other.options.path != "" {
		return true
	}
	if (h.options.pathPrefix == "" && h.options.pathSuffix == "") && (other.options.pathPrefix != "" || other.options.pathSuffix != "") {
		return true

	}
	return false
}

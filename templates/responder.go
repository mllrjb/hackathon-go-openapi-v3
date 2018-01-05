package operation

import (
	"encoding/json"
	"net/http"
)

type Responder interface {
	WriteResponse(writer http.ResponseWriter)
}

type statusCodeResponder struct {
	StatusCode int
}

func (r *statusCodeResponder) WriteResponse(writer http.ResponseWriter) {
	writer.WriteHeader(r.StatusCode)
}

func StatusCodeResponder(statusCode int) Responder {
	r := statusCodeResponder{
		StatusCode: statusCode,
	}
	return &r
}

type jsonResponder struct {
	StatusCode  int
	Body        interface{}
	ContentType string
}

func (r *jsonResponder) WriteResponse(writer http.ResponseWriter) {
	bytes, err := json.Marshal(r.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", r.ContentType)
	writer.WriteHeader(r.StatusCode)
	writer.Write(bytes)
}

func JsonResponder(statusCode int, contentType string, body interface{}) Responder {
	r := jsonResponder{
		StatusCode:  statusCode,
		ContentType: contentType,
		Body:        body,
	}
	return &r
}

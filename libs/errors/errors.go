package errors

import (
	"errors"
	"net/http"
	"strings"
)

const (
	StatusIncorrectXml      = "INCORRECT_XML"
	StatusIncorrectJson     = "INCORRECT_JSON"
	StatusIncorrectName     = "INCORRECT_NAME"
	StatusIncorrectEmail    = "INCORRECT_EMAIL"
	StatusIncorrectUsename  = "INCORRECT_USERNAME"
	StatusIncorrectPassword = "INCORRECT_PASSWORD"
	StatusIncorrectAuth     = "INCORRECT_AUTH"
	StatusIncorrectPayload  = "INCORRECT_PAYLOAD"

	StatusBadRequest   = "BAD_REQUEST"
	StatusBadGateway   = "BAD_GATEWAY"
	StatusBadParameter = "BAD_PARAMETER"

	StatusNotFound       = "NOT_FOUND"
	StatusNotUnique      = "NOT_UNIQUE"
	StatusNotSupported   = "NOT_SUPPORTED"
	StatusNotAcceptable  = "NOT_ACCEPTABLE"
	StatusNotImplemented = "NOT_IMPLEMENTED"

	StatusPaymentRequired = "PAYMENT_REQUIRED"
	StatusAccessDenied    = "ACCESS_DENIED"

	StatusForbidden        = "FORBIDDEN"
	StatusMethodNotAllowed = "METHOD_NOT_ALLOWED"

	StatusInternalServerError = "INTERNAL_SERVER_ERROR"

	StatusUnknown = "UNKNOWN"
)

type Err struct {
	Code   string
	Attr   string
	origin error
	http   *Http
}

type err struct {
	name string
}

func (e *Err) Err() error {
	return e.origin
}

func (e *Err) Http(w http.ResponseWriter) {
	e.http.send(w)
}

func New(name string) *err {
	return &err{strings.ToLower(name)}
}

func (self *err) NotFound(e ...error) *Err {
	return &Err{
		Code:   StatusNotFound,
		origin: getError(toUpperFirstChar(self.name)+": not found", e...),
		http:   HTTP.getNotFound(self.name),
	}
}

func (self *err) BadParameter(attr string, e ...error) *Err {
	return &Err{
		Code:   StatusBadParameter,
		Attr:   attr,
		origin: getError(toUpperFirstChar(self.name)+": bad parameter", e...),
		http:   HTTP.getBadParameter(attr),
	}
}

func (self *err) IncorrectJSON(e ...error) *Err {
	return &Err{
		Code:   StatusIncorrectJson,
		origin: getError(toUpperFirstChar(self.name)+": incorrect json", e...),
		http:   HTTP.getIncorrectJSON(),
	}
}

func (self *err) Unknown(e ...error) *Err {
	return &Err{
		Code:   StatusUnknown,
		origin: getError(toUpperFirstChar(self.name)+": unknow error", e...),
		http:   HTTP.getUnknown(),
	}
}

func getError(msg string, e ...error) error {
	if len(e) == 0 {
		return errors.New(msg)
	} else {
		return e[0]
	}
}

func toUpperFirstChar(srt string) string {
	return strings.ToUpper(srt[0:1]) + srt[1:]
}

package api

import (
	"encoding/json"
	"fmt"
)

//stringify print the contents of the obj
func stringify(data interface{}) string {
	if data == nil {
		return "null"
	}
	var p []byte
	type d struct {
		Arguments interface{}
	}
	p, err := json.MarshalIndent(d{Arguments: data}, "", "\t")
	if err != nil {
		return err.Error()
	}
	return string(p)
}

//ErrorStack bas class for providers error management
type ErrorStack struct {
	Cause   error
	Message string
}

//Error format error message
func (e *ErrorStack) Error() string {
	if e.Cause != nil {
		if e.Message != "" {
			return fmt.Sprintf("%s\nCaused by: %s", e.Message, e.Cause.Error())
		}
		return e.Cause.Error()
	}
	return e.Message
}

//NewErrorStack create a new provider error
func NewErrorStack(cause error, message string, args ...interface{}) *ErrorStack {
	msg := message
	if args != nil {
		msg = fmt.Sprintf("%s :\n%s", message, stringify(args))
	}
	return &ErrorStack{
		Cause:   cause,
		Message: msg,
	}
}

////NewErrorStackFromMessage create a new provider error
//func NewErrorStackFromMessage(cause error, message string) *ErrorStack {
//	return &ErrorStack{
//		Cause:   cause,
//		Message: message,
//	}
//}

//NewErrorStackFromError create a new provider error
func NewErrorStackFromError(cause error, err error) *ErrorStack {
	if err == nil && cause == nil {
		return nil
	}
	var msg string
	if err != nil {
		msg = err.Error()
	}
	return &ErrorStack{
		Cause:   cause,
		Message: msg,
	}
}

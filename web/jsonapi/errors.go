package jsonapi

import (
	"net/http"
	"os"

	"github.com/cozy/cozy-stack/couchdb"
	"github.com/cozy/cozy-stack/vfs"
)

// SourceError contains references to the source of the error
type SourceError struct {
	Pointer   string `json:"pointer,omitempty"`
	Parameter string `json:"parameter,omitempty"`
}

// Error objects provide additional information about problems encountered
// while performing an operation.
// See http://jsonapi.org/format/#error-objects
type Error struct {
	Status int         `json:"status,string"`
	Title  string      `json:"title"`
	Detail string      `json:"detail"`
	Source SourceError `json:"source,omitempty"`
}

// WrapCouchError returns a formatted error from a couchdb error
func WrapCouchError(err *couchdb.Error) *Error {
	return &Error{
		Status: err.StatusCode,
		Title:  err.Name,
		Detail: err.Reason,
	}
}

// WrapVfsError returns a formatted error from a golang error emitted by the vfs
func WrapVfsError(err error) *Error {
	if couchErr, isCouchErr := err.(*couchdb.Error); isCouchErr {
		return WrapCouchError(couchErr)
	}
	if os.IsExist(err) {
		return &Error{
			Status: http.StatusConflict,
			Title:  "Conflict",
			Detail: err.Error(),
		}
	}
	if os.IsNotExist(err) {
		return NotFound(err)
	}
	switch err {
	case vfs.ErrParentDoesNotExist:
		return NotFound(err)
	case vfs.ErrDocTypeInvalid:
		return InvalidAttribute("type", err)
	case vfs.ErrIllegalFilename:
		return InvalidParameter("folder-id", err)
	case vfs.ErrInvalidHash:
		return PreconditionFailed("Content-MD5", err)
	case vfs.ErrContentLengthMismatch:
		return PreconditionFailed("Content-Length", err)
	}
	return InternalServerError(err)
}

// NotFound returns a 404 formatted error
func NotFound(err error) *Error {
	return &Error{
		Status: http.StatusNotFound,
		Title:  "Not Found",
		Detail: err.Error(),
	}
}

// InternalServerError returns a 500 formatted error
func InternalServerError(err error) *Error {
	return &Error{
		Status: http.StatusInternalServerError,
		Title:  "Internal Server Error",
		Detail: err.Error(),
	}
}

// PreconditionFailed returns a 412 formatted error when an expectation from an
// HTTP header is not matched
func PreconditionFailed(parameter string, err error) *Error {
	return &Error{
		Status: http.StatusPreconditionFailed,
		Title:  "Precondition Failed",
		Detail: err.Error(),
		Source: SourceError{
			Parameter: parameter,
		},
	}
}

// InvalidParameter returns a 422 formatted error when an HTTP or Query-String
// parameter is invalid
func InvalidParameter(parameter string, err error) *Error {
	return &Error{
		Status: http.StatusUnprocessableEntity,
		Title:  "Invalid Parameter",
		Detail: err.Error(),
		Source: SourceError{
			Parameter: parameter,
		},
	}
}

// InvalidAttribute returns a 422 formatted error when an attribute is invalid
func InvalidAttribute(attribute string, err error) *Error {
	return &Error{
		Status: http.StatusUnprocessableEntity,
		Title:  "Invalid Attribute",
		Detail: err.Error(),
		Source: SourceError{
			Pointer: "/data/attributes/" + attribute,
		},
	}
}
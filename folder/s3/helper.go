package s3

import (
	"errors"
	"net/http"
)

// -----------------------------------------------------------------------
// helpers
// -----------------------------------------------------------------------

// isNotFound checks if an AWS error is a 404 / NoSuchKey.
func isNotFound(err error) bool {
	// aws-sdk-go-v2 wraps HTTP status in *smithy.OperationError → *http.Response.
	// The simplest portable check: look for the ResponseError interface.
	type httpResponseError interface {
		HTTPStatusCode() int
	}
	var re httpResponseError
	if errors.As(err, &re) {
		return re.HTTPStatusCode() == http.StatusNotFound
	}
	return false
}

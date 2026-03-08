package s3

import "net/http"

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
	if ok := asError(err, &re); ok {
		return re.HTTPStatusCode() == http.StatusNotFound
	}
	return false
}

// asError is a tiny generic helper to avoid importing errors in every call site.
func asError[T any](err error, target *T) bool {
	// This is equivalent to errors.As but generic.
	for err != nil {
		if t, ok := any(err).(T); ok { //nolint:errorlint
			*target = t
			return true
		}
		// unwrap
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}

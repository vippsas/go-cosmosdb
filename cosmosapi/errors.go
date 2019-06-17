package cosmosapi

import (
	"net/http"

	"github.com/pkg/errors"
)

// StatusRetryWith defines the 449 http error. Not present in go std lib
const (
	StatusRetryWith = 449
)

var (
	errRetry                   = errors.New("retry")
	ErrorNotImplemented        = errors.New("not implemented")
	ErrWrongQueryContentType   = errors.New("Wrong content type. Must be " + QUERY_CONTENT_TYPE)
	ErrMaxRetriesExceeded      = errors.New("Max retries exceeded")
	ErrInvalidPartitionKeyType = errors.New("Partition key type must be a simple type (nil, string, int, float, etc.)")

	// Map http codes to cosmos errors messages
	// Description taken directly from https://docs.microsoft.com/en-us/rest/api/cosmos-db/http-status-codes-for-cosmosdb
	ErrInvalidRequest     = errors.New("The JSON, SQL, or JavaScript in the request body is invalid")
	ErrUnautorized        = errors.New("The Authorization header is invalid for the requested resource")
	ErrForbidden          = errors.New("The authorization token expired, resource quota has been reached or high resource usage")
	ErrNotFound           = errors.New("Resource that no longer exists")
	ErrTimeout            = errors.New("The operation did not complete within the allotted amount of time")
	ErrConflict           = errors.New("The ID provided has been taken by an existing resource")
	ErrPreconditionFailed = errors.New("The operation specified an eTag that is different from the version available at the server")
	ErrTooLarge           = errors.New("The document size in the request exceeded the allowable document size for a request")
	ErrTooManyRequests    = errors.New("The collection has exceeded the provisioned throughput limit")
	ErrRetryWith          = errors.New("The operation encountered a transient error. It is safe to retry the operation")
	ErrInternalError      = errors.New("The operation failed due to an unexpected service error")
	ErrUnavailable        = errors.New("The operation could not be completed because the service was unavailable")
	// Undocumented code. A known scenario where it is used is when doing a ListDocuments request with ReadFeed
	// properties on a partition that was split by a repartition.
	ErrGone = errors.New("Resource is gone")

	CosmosHTTPErrors = map[int]error{
		http.StatusOK:                    nil,
		http.StatusCreated:               nil,
		http.StatusNoContent:             nil,
		http.StatusNotModified:           nil,
		http.StatusBadRequest:            ErrInvalidRequest,
		http.StatusUnauthorized:          ErrUnautorized,
		http.StatusForbidden:             ErrForbidden,
		http.StatusNotFound:              ErrNotFound,
		http.StatusRequestTimeout:        ErrTimeout,
		http.StatusConflict:              ErrConflict,
		http.StatusGone:                  ErrGone,
		http.StatusPreconditionFailed:    ErrPreconditionFailed,
		http.StatusRequestEntityTooLarge: ErrTooLarge,
		http.StatusTooManyRequests:       ErrTooManyRequests,
		StatusRetryWith:                  ErrRetryWith,
		http.StatusInternalServerError:   ErrInternalError,
		http.StatusServiceUnavailable:    ErrUnavailable,
	}
)

package cosmosdb

import (
	"github.com/pkg/errors"
)

var (
	errRetry = errors.New("retry")

	ErrPreconditionFailed = errors.New("precondition failed")
	ErrorNotImplemented   = errors.New("not implemented")

	ErrConflict              = errors.New("Resource with specified id or name already exists.")
	ErrWrongQueryContentType = errors.New("Wrong content type. Must be " + QUERY_CONTENT_TYPE)
)

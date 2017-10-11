package dom

import (
	"github.com/pkg/errors"
)

var (
	// ErrNoModificationAllowed indicates no modification was allowed.
	ErrNoModificationAllowed = errors.New("no modification allowed")
	// ErrIndexSize indicates an indexing argument (such as an "offset" or "count") was invalid
	// given the operation requested.
	ErrIndexSize = errors.New("index size error")
	// ErrChildNotFound indicates the child was not found
	ErrChildNotFound = errors.New("child not found")
	// ErrAttributeNotFound indicates the child attribute was not found
	ErrAttributeNotFound = errors.New("attribute not found")
	// ErrHierarchyRequest indicates a request element hierarchy error
	ErrHierarchyRequest = errors.New("hierarchy request error")

	errBadType = errors.New("unexpected type")
)

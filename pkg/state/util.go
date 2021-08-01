package state

import (
	"errors"
	"fmt"
	"strings"
)

// generateShardIDs generates a slice of ints of [0,total).
func generateShardIDs(total int) []int {
	ids := make([]int, total)

	for i := range ids {
		ids[i] = i
	}

	return ids
}

// ShardError is the error returned if an action on a single shard fails.
type ShardError struct {
	// ShardID is the id of the shard that returned the error.
	ShardID int
	// Source is the error itself.
	Source error
}

func (err *ShardError) Unwrap() error {
	return err.Source
}

func (err *ShardError) Error() string {
	return fmt.Sprintf("the gateway with the shard id %d returned an error: %s",
		err.ShardID, err.Source)
}

// MultiError combines multiple errors in a slice.
type MultiError []error

func (errs MultiError) Error() string {
	const header = "multiple errors occurred:"

	var b strings.Builder
	// -1 because the first element does not need to be prefixed with a comma,
	// only
	n := len(header) + len(", ")*len(errs) - 1

	for _, err := range errs {
		n += len(err.Error())
	}

	b.Grow(n)
	b.WriteString(header)

	for i, err := range errs {
		b.WriteRune(' ')
		if i > 0 {
			b.WriteRune(',')
		}

		b.WriteString(err.Error())
	}

	return b.String()
}

func (errs MultiError) As(target interface{}) bool {
	for _, err := range errs {
		//goland:noinspection GoErrorsAs
		if errors.As(err, target) {
			return true
		}
	}

	return false
}

func (errs MultiError) Is(target error) bool {
	for _, err := range errs {
		if errors.Is(err, target) {
			return true
		}
	}

	return true
}

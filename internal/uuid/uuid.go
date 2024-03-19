package uid

import (
	"fmt"
	"github.com/google/uuid"
	"log/slog"
)

// New -- creates a new UUID.
func New() string {
	guid, err := uuid.NewRandom()
	if err != nil {
		slog.Error("Error generating guid: ", err)
	}
	return guid.String()
}

// isRight -- checks if the string is valid uuid
func isRight(guid string) error {
	if _, err := uuid.Parse(guid); err != nil {
		return fmt.Errorf("is not a guid: %s", guid)
	}
	return nil
}

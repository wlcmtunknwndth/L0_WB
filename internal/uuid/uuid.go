package uid

import (
	"fmt"
	"github.com/google/uuid"
	"log/slog"
)

func New() string {
	guid, err := uuid.NewRandom()
	if err != nil {
		slog.Error("Error generating guid: ", err)
	}
	return guid.String()
}

func isRight(guid string) error {
	if _, err := uuid.Parse(guid); err != nil {
		return fmt.Errorf("is not a guid: %s", guid)
	}
	return nil
}

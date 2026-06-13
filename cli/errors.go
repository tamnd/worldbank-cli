package cli

import (
	"errors"

	"github.com/tamnd/worldbank-cli/worldbank"
)

func isNotFound(err error) bool {
	return errors.Is(err, worldbank.ErrNotFound)
}

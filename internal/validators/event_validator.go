package validators

import (
	"fmt"

	"fraud-platform/internal/domain"
)

func ValidateBusinessRules(event domain.Event) error {
	if event.Country == "BR" && event.Device == "unknown" {
		return fmt.Errorf("won't accept BR signup/login with unknown device here")
	}

	return nil
}

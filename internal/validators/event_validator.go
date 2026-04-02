package validators

import (
	"fmt"

	"fraud-platform/internal/domain"
)

func ValidateBusinessRules(event domain.Event) error {
	if event.Country == "BR" && event.Device == "unknown" {
		return fmt.Errorf("suspicious combination: BR + unknown device")
	}

	return nil
}

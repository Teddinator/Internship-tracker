package contacts

import "strings"

func stringPtrOrNil(value string) *string {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil
	}

	return &value

}

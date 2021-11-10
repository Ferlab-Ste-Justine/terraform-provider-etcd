package provider

import (
	"errors"
	"fmt"
)

func isStringInSlice(val string, slice []string) bool {
	for _, elem := range slice {
		if elem == val {
			return true
		}
	}
	return false
}

func getPrefixRangeEnd(key string) (string, error) {
	end := []byte(key)
	for idx := len(end) - 1; idx >= 0; idx-- {
		if end[idx] < 0xff {
			end[idx] = end[idx] + 1
			return string(end), nil
		}
	}

	return "", errors.New(fmt.Sprintf("String '%s' cannot be a prefix as it cannot be incremented", key))
}
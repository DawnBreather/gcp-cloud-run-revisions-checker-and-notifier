package utils

import (
	"strings"
)

func ExtractShortServiceName(longServiceName string) string {
	parts := strings.Split(longServiceName, "/")
	return parts[len(parts)-1] // Get the last element of the slice
}

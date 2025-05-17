package helpers

import "os"

// GetEnv wrapper allowing for the setting of a default string value
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

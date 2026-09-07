// Package config reads process environment variables the same way in every service.
package config

import (
	"cmp"
	"fmt"
	"os"
	"strconv"
)

func Require(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}

func Or(name, fallback string) string {
	return cmp.Or(os.Getenv(name), fallback)
}

func Bool(name string, fallback bool) (bool, error) {
	value, ok := os.LookupEnv(name)
	if !ok || value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean", name)
	}
	return parsed, nil
}

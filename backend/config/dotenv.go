// Package config reads settings from environment variables.
package config

import (
	"bufio"
	"os"
	"strings"
)

// LoadDotEnv reads KEY=VALUE lines from a file and sets them as environment
// variables. It is only for local development: on Vercel and Render the
// variables come from the host, and the file does not exist.
//
// Variables that are already set are left alone, so a real environment
// variable always wins over the file.
func LoadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return // no file is the normal case in production
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if _, alreadySet := os.LookupEnv(key); !alreadySet {
			os.Setenv(key, value)
		}
	}
}

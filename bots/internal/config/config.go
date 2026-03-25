// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

// Package config provides .env file parsing utilities.
package config

import (
	"bytes"
	"fmt"
	"os"
)

type Config struct {
	AccountID string
	TokenVar  string
	EnvPath   string
}

func (c Config) Token() (string, error) {
	token, err := lookup(c.EnvPath, c.TokenVar)
	if err != nil {
		return "", fmt.Errorf("config: token: %w", err)
	}

	return token, nil
}

func lookup(path, key string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("lookup: %w", err)
	}
	bkey := []byte(key + "=")
	for line := range bytes.SplitSeq(data, []byte("\n")) {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		if bytes.HasPrefix(line, bkey) {
			value := bytes.TrimRight(line[len(bkey):], "\r")
			return string(value), nil
		}
	}
	return "", fmt.Errorf("lookup: key %q not found", key)
}

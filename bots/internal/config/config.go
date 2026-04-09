// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

// Package config provides .env file parsing utilities.
package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

func Lookup(path, key string) (string, error) {
	cpath := filepath.Clean(path)
	data, err := os.ReadFile(cpath)
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

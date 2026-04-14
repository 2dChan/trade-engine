// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package trade

import (
	"fmt"
	"strings"
)

const (
	sep = "_"
)

type InstrumentID string

func NewInstrumentID(ticker string, classCode string) (InstrumentID, error) {
	if err := validateInstrumentIDPart(ticker); err != nil {
		return "", fmt.Errorf("instrument id: ticker: %w", err)
	}
	if err := validateInstrumentIDPart(classCode); err != nil {
		return "", fmt.Errorf("instrument id: class code: %w", err)
	}
	return InstrumentID(ticker + sep + classCode), nil
}

func ParseInstrumentID(raw string) (InstrumentID, error) {
	if raw == "" {
		return "", fmt.Errorf("instrument id: empty value")
	}
	ticker, classCode, ok := strings.Cut(raw, sep)
	if !ok {
		return "", fmt.Errorf("instrument id: invalid format %q, expected TICKER%sCLASSCODE", raw, sep)
	}
	if err := validateInstrumentIDPart(ticker); err != nil {
		return "", fmt.Errorf("instrument id: ticker: %w", err)
	}
	if err := validateInstrumentIDPart(classCode); err != nil {
		return "", fmt.Errorf("instrument id: class code: %w", err)
	}
	return InstrumentID(raw), nil
}

func (id InstrumentID) String() string {
	return string(id)
}

func (id InstrumentID) Ticker() string {
	ticker, _, ok := id.Split()
	if !ok {
		return ""
	}
	return ticker
}

func (id InstrumentID) ClassCode() string {
	_, classCode, ok := id.Split()
	if !ok {
		return ""
	}
	return classCode
}

func (id InstrumentID) Split() (ticker string, classCode string, ok bool) {
	raw := string(id)
	if raw == "" {
		return "", "", false
	}
	ticker, classCode, ok = strings.Cut(raw, sep)
	if !ok {
		return "", "", false
	}
	if err := validateInstrumentIDPart(ticker); err != nil {
		return "", "", false
	}
	if err := validateInstrumentIDPart(classCode); err != nil {
		return "", "", false
	}
	return ticker, classCode, true
}

func validateInstrumentIDPart(part string) error {
	if part == "" {
		return fmt.Errorf("empty value")
	}

	for i, b := range part {
		if (b < 'A' || b > 'Z') && (b < '0' || b > '9') {
			return fmt.Errorf("must contain only [A-Z0-9] at byte %d", i)
		}
	}
	return nil
}

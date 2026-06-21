// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package cron

import (
	"fmt"
	"math/bits"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ParseOption is a bit flag of options to control how the parser interprets
// cron expression fields.
type ParseOption int

const (
	Second         ParseOption = 1 << iota // Seconds field, default 0
	SecondOptional                         // Optional seconds field, default 0
	Minute                                 // Minutes field, default 0
	Hour                                   // Hours field, default 0
	Dom                                    // Day of month field, default *
	Month                                  // Month field, default *
	Dow                                    // Day of week field, default *
	DowOptional                            // Optional day of week field, default *
	Descriptor                             // Allow @every, @hourly etc descriptors
)

var monthNames = map[string]int{
	"JAN": 1, "FEB": 2, "MAR": 3, "APR": 4,
	"MAY": 5, "JUN": 6, "JUL": 7, "AUG": 8,
	"SEP": 9, "OCT": 10, "NOV": 11, "DEC": 12,
}

var weekdayNames = map[string]int{
	"SUN": 0, "MON": 1, "TUE": 2, "WED": 3,
	"THU": 4, "FRI": 5, "SAT": 6,
}

// fieldBounds defines supported cron field ranges.
var fieldBounds = []struct {
	min int
	max int
}{
	{0, 59}, // seconds
	{0, 59}, // minutes
	{0, 23}, // hours
	{1, 31}, // day of month
	{1, 12}, // month
	{0, 7},  // day of week (0 and 7 both represent Sunday)
}

// Parser is a configurable cron expression parser.
type Parser struct {
	options ParseOption
}

// NewParser creates a Parser with the given options.
// Returns an error if more than one optional flag is configured.
func NewParser(options ParseOption) (Parser, error) {
	optionals := 0
	if options&DowOptional > 0 {
		optionals++
	}
	if options&SecondOptional > 0 {
		optionals++
	}
	if optionals > 1 {
		return Parser{}, fmt.Errorf("cron: multiple optional flags may not be configured")
	}
	return Parser{options: options}, nil
}

// defaultParser fields are lazily initialized via [sync.Once] to avoid a
// package-level init panic. The default config is a compile-time constant
// known to be valid, so the error path is unreachable in practice; the
// error is propagated through [Parse] rather than panicking to honor the
// "no panic in business logic" rule.
var (
	defaultParserOnce sync.Once
	defaultParser     Parser
	defaultParserErr  error
)

// initDefaultParser lazily initializes and returns the default parser.
// The error is unreachable in practice (the default config is valid),
// but is returned rather than panicking to honor the "no panic in
// business logic" rule.
func initDefaultParser() (Parser, error) {
	defaultParserOnce.Do(func() {
		defaultParser, defaultParserErr = NewParser(
			SecondOptional | Minute | Hour | Dom | Month | Dow | Descriptor,
		)
	})
	return defaultParser, defaultParserErr
}

// Parse parses a cron expression using the default parser.
func Parse(expr string) (Schedule, error) {
	p, err := initDefaultParser()
	if err != nil {
		return nil, fmt.Errorf("cron: init default parser: %w", err)
	}
	return p.Parse(expr)
}

// Parse parses a cron expression according to the parser's options.
func (p Parser) Parse(spec string) (Schedule, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, fmt.Errorf("cron: empty spec string")
	}

	// Extract timezone prefix
	loc := time.Local
	if strings.HasPrefix(spec, "TZ=") || strings.HasPrefix(spec, "CRON_TZ=") {
		parts := strings.SplitN(spec, " ", 2)
		var tzStr string
		if strings.HasPrefix(parts[0], "TZ=") {
			tzStr = strings.TrimPrefix(parts[0], "TZ=")
		} else {
			tzStr = strings.TrimPrefix(parts[0], "CRON_TZ=")
		}
		var err error
		loc, err = time.LoadLocation(tzStr)
		if err != nil {
			return nil, fmt.Errorf("cron: provided bad location %s: %w", tzStr, err)
		}
		if len(parts) > 1 {
			spec = parts[1]
		} else {
			spec = ""
		}
	}

	// Handle @ descriptors
	if strings.HasPrefix(spec, "@") {
		if p.options&Descriptor == 0 {
			return nil, fmt.Errorf("cron: descriptors are not supported")
		}
		return parseDescriptor(spec, loc)
	}

	// Split spec into fields
	fields := strings.Fields(spec)

	// Calculate expected field count
	places := []ParseOption{Second, Minute, Hour, Dom, Month, Dow}
	defaults := []string{"0", "0", "0", "*", "*", "*"}

	minFields := 0
	maxFields := 0
	for _, place := range places {
		if p.options&place > 0 {
			minFields++
			maxFields++
		}
	}
	if p.options&SecondOptional > 0 {
		maxFields++
	}
	if p.options&DowOptional > 0 {
		maxFields++
	}

	if len(fields) < minFields || len(fields) > maxFields {
		return nil, fmt.Errorf("cron: expected %d to %d fields, got %d", minFields, maxFields, len(fields))
	}

	// Parse each field
	sched := &specSchedule{loc: loc}
	extraFields := len(fields) - minFields
	domIsQuestion := false
	dowIsQuestion := false

	fieldIndex := 0
	for i, place := range places {
		var token string
		if p.options&place > 0 {
			token = fields[fieldIndex]
			fieldIndex++
		} else if (place == Second && p.options&SecondOptional > 0) ||
			(place == Dow && p.options&DowOptional > 0) {
			if extraFields > 0 {
				token = fields[fieldIndex]
				fieldIndex++
				extraFields--
			} else {
				token = defaults[i]
			}
		} else {
			token = defaults[i]
		}

		// Reject L, W, # characters (case insensitive)
		upper := strings.ToUpper(token)
		if strings.ContainsAny(upper, "LW#") {
			return nil, fmt.Errorf("cron: unsupported syntax: %s contains L, W, or #", token)
		}

		// ? only allowed in DOM (index 3) and DOW (index 5)
		if token == "?" && i != 3 && i != 5 {
			return nil, fmt.Errorf("cron: ? is only allowed in day-of-month and day-of-week fields")
		}

		// Track ? for mutual exclusion check
		if token == "?" {
			if i == 3 {
				domIsQuestion = true
			}
			if i == 5 {
				dowIsQuestion = true
			}
		}

		mask, err := parseField(token, fieldBounds[i].min, fieldBounds[i].max, i)
		if err != nil {
			return nil, err
		}

		switch i {
		case 0:
			sched.sec = mask
		case 1:
			sched.min = mask
		case 2:
			sched.hour = uint32(mask)
		case 3:
			sched.dom = uint32(mask)
			if token == "*" || token == "?" {
				sched.domStar = true
			}
		case 4:
			sched.month = uint16(mask)
		case 5:
			sched.dow = uint8(mask)
			if token == "*" || token == "?" {
				sched.dowStar = true
			}
		}
	}

	// DOM and DOW cannot both be ?
	if domIsQuestion && dowIsQuestion {
		return nil, fmt.Errorf("cron: ? cannot be used for both day-of-month and day-of-week")
	}

	// Validate non-zero fields
	if sched.sec == 0 {
		return nil, fmt.Errorf("cron: seconds field cannot be empty")
	}
	if sched.min == 0 {
		return nil, fmt.Errorf("cron: minutes field cannot be empty")
	}
	if sched.hour == 0 {
		return nil, fmt.Errorf("cron: hours field cannot be empty")
	}
	if sched.dom == 0 {
		return nil, fmt.Errorf("cron: day-of-month field cannot be empty")
	}
	if sched.month == 0 {
		return nil, fmt.Errorf("cron: month field cannot be empty")
	}
	if sched.dow == 0 {
		return nil, fmt.Errorf("cron: day-of-week field cannot be empty")
	}

	return sched, nil
}

func parseDescriptor(spec string, loc *time.Location) (Schedule, error) {
	lower := strings.ToLower(spec)
	switch {
	case strings.HasPrefix(lower, "@every "):
		duration, err := time.ParseDuration(spec[7:])
		if err != nil {
			return nil, fmt.Errorf("cron: failed to parse duration %s: %w", spec[7:], err)
		}
		return Every(duration), nil

	case lower == "@yearly" || lower == "@annually":
		return &specSchedule{
			sec:     1 << 0,
			min:     1 << 0,
			hour:    1 << 0,
			dom:     1 << 1,
			month:   1 << 1,
			dow:     uint8(allBits(0, 7)),
			dowStar: true,
			loc:     loc,
		}, nil

	case lower == "@monthly":
		return &specSchedule{
			sec:     1 << 0,
			min:     1 << 0,
			hour:    1 << 0,
			dom:     1 << 1,
			month:   uint16(allBits(1, 12)),
			dow:     uint8(allBits(0, 7)),
			dowStar: true,
			loc:     loc,
		}, nil

	case lower == "@weekly":
		return &specSchedule{
			sec:     1 << 0,
			min:     1 << 0,
			hour:    1 << 0,
			dom:     uint32(allBits(1, 31)),
			month:   uint16(allBits(1, 12)),
			dow:     1 << 0,
			domStar: true,
			loc:     loc,
		}, nil

	case lower == "@daily" || lower == "@midnight":
		return &specSchedule{
			sec:     1 << 0,
			min:     1 << 0,
			hour:    1 << 0,
			dom:     uint32(allBits(1, 31)),
			month:   uint16(allBits(1, 12)),
			dow:     uint8(allBits(0, 7)),
			domStar: true,
			dowStar: true,
			loc:     loc,
		}, nil

	case lower == "@hourly":
		return &specSchedule{
			sec:     1 << 0,
			min:     1 << 0,
			hour:    uint32(allBits(0, 23)),
			dom:     uint32(allBits(1, 31)),
			month:   uint16(allBits(1, 12)),
			dow:     uint8(allBits(0, 7)),
			domStar: true,
			dowStar: true,
			loc:     loc,
		}, nil

	case lower == "@reboot":
		return &immediateSchedule{}, nil

	default:
		return nil, fmt.Errorf("cron: unrecognized descriptor: %s", spec)
	}
}

func parseField(token string, min, max, index int) (uint64, error) {
	token = strings.TrimSpace(token)
	if token == "*" || token == "?" {
		return allBits(min, max), nil
	}

	mask := uint64(0)
	for _, part := range strings.Split(token, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			return 0, fmt.Errorf("cron: invalid empty token")
		}

		step := 1
		hasStep := false
		if strings.Contains(part, "/") {
			pieces := strings.Split(part, "/")
			if len(pieces) != 2 {
				return 0, fmt.Errorf("cron: invalid step syntax: %s", part)
			}
			part = strings.TrimSpace(pieces[0])
			if part == "" {
				part = "*"
			}
			stepValue, err := strconv.Atoi(pieces[1])
			if err != nil || stepValue <= 0 {
				return 0, fmt.Errorf("cron: invalid step value: %s", pieces[1])
			}
			step = stepValue
			hasStep = true
		}

		var start, end int
		if part == "*" || part == "?" {
			start, end = min, max
		} else if strings.Contains(part, "-") {
			rangePieces := strings.Split(part, "-")
			if len(rangePieces) != 2 {
				return 0, fmt.Errorf("cron: invalid range syntax: %s", part)
			}
			var err error
			start, err = parseValue(rangePieces[0], min, max, index)
			if err != nil {
				return 0, err
			}
			end, err = parseValue(rangePieces[1], min, max, index)
			if err != nil {
				return 0, err
			}
			if start > end {
				if index == 5 {
					for value := start; value <= max; value += step {
						mask |= bitFor(value)
					}
					for value := min; value <= end; value += step {
						mask |= bitFor(value)
					}
					continue
				}
				return 0, fmt.Errorf("cron: range start cannot be greater than end: %s", part)
			}
		} else {
			value, err := parseValue(part, min, max, index)
			if err != nil {
				return 0, err
			}
			if hasStep {
				start, end = value, max
			} else {
				start, end = value, value
			}
		}

		for value := start; value <= end; value += step {
			mask |= bitFor(value)
		}
	}

	return mask, nil
}

func parseValue(token string, min, max, index int) (int, error) {
	token = strings.TrimSpace(strings.ToUpper(token))
	if token == "*" || token == "?" {
		return min, nil
	}

	if index == 4 {
		if v, ok := monthNames[token]; ok {
			return v, nil
		}
	}
	if index == 5 {
		if v, ok := weekdayNames[token]; ok {
			return v, nil
		}
		if token == "7" {
			return 0, nil
		}
	}

	value, err := strconv.Atoi(token)
	if err != nil {
		return 0, fmt.Errorf("cron: invalid numeric value %q", token)
	}

	if index == 5 && value == 7 {
		return 0, nil
	}

	if value < min || value > max {
		return 0, fmt.Errorf("cron: value %d out of range for field %d", value, index)
	}
	return value, nil
}

func allBits(min, max int) uint64 {
	if min > max || max-min+1 >= 64 {
		return ^uint64(0)
	}
	return ((uint64(1) << uint(max-min+1)) - 1) << uint(min)
}

func bitFor(value int) uint64 {
	return uint64(1) << uint(value)
}

func bitMatch(mask uint64, value int) bool {
	return mask&(uint64(1)<<uint(value)) != 0
}

func nextSetBit(mask uint64, from, max int) (int, bool) {
	if from > max {
		return 0, false
	}
	shifted := mask >> uint(from)
	if shifted == 0 {
		return 0, false
	}
	pos := bits.TrailingZeros64(shifted)
	value := from + pos
	if value > max {
		return 0, false
	}
	return value, true
}

package gbcarkhos

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	coreenv "goark.dev/goark/core/env"
	arkerrors "goark.dev/goark/errors"
)

var byteSizeUnits = map[string]int64{
	"":    1,
	"B":   1,
	"KB":  1000,
	"MB":  1000 * 1000,
	"GB":  1000 * 1000 * 1000,
	"KIB": 1 << 10,
	"MIB": 1 << 20,
	"GIB": 1 << 30,
}

func readByteSize(environment coreenv.Environment, key string, target *int) error {
	value, ok := environment.GetProperty(key)
	if !ok {
		return nil
	}
	size, err := parseByteSize(value)
	if err != nil {
		return arkerrors.Wrapf(arkerrors.CodeConversion, err, "failed to read byte size property %q", key)
	}
	if size > int64(math.MaxInt) {
		return arkerrors.Newf(arkerrors.CodeConversion, "byte size property %q overflows int", key)
	}
	*target = int(size)
	return nil
}

func readByteSize64(environment coreenv.Environment, key string, target *int64) error {
	value, ok := environment.GetProperty(key)
	if !ok {
		return nil
	}
	size, err := parseByteSize(value)
	if err != nil {
		return arkerrors.Wrapf(arkerrors.CodeConversion, err, "failed to read byte size property %q", key)
	}
	*target = size
	return nil
}

func parseByteSize(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("byte size is empty")
	}
	index := 0
	for index < len(value) && value[index] >= '0' && value[index] <= '9' {
		index++
	}
	if index == 0 {
		return 0, fmt.Errorf("byte size %q has no numeric value", value)
	}
	number, err := strconv.ParseInt(value[:index], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse byte size %q: %w", value, err)
	}
	unit := strings.ToUpper(strings.TrimSpace(value[index:]))
	multiplier, ok := byteSizeUnits[unit]
	if !ok {
		return 0, fmt.Errorf("byte size %q uses unsupported unit %q", value, unit)
	}
	if number <= 0 {
		return 0, fmt.Errorf("byte size %q must be positive", value)
	}
	if number > math.MaxInt64/multiplier {
		return 0, fmt.Errorf("byte size %q overflows int64", value)
	}
	return number * multiplier, nil
}

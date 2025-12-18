package common

import (
	"strconv"
	"strings"
)

func ParseDiskSize(sizeStr string) (float64, error) {
	sizeStr = strings.TrimSpace(sizeStr)
	if len(sizeStr) == 0 {
		return 0, nil
	}

	multiplier := float64(1)
	unit := sizeStr[len(sizeStr)-1:]

	switch strings.ToUpper(unit) {
	case "K":
		multiplier = 1024
	case "M":
		multiplier = 1024 * 1024
	case "G":
		multiplier = 1024 * 1024 * 1024
	case "T":
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		unit = ""
	}

	numStr := sizeStr
	if unit != "" {
		numStr = sizeStr[:len(sizeStr)-1]
	}

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, err
	}

	return num * multiplier, nil
}

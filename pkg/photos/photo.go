package photos

import (
	"strings"
)

type Size int

const (
	SizeXS     Size = 10
	SizeSmall  Size = 200
	SizeMedium Size = 800
	SizeLarge  Size = 1200
	SizeXL     Size = 2000
	SizeFull   Size = 999999
)

func (s Size) String() string {
	switch s {
	case SizeXS:
		return "x-small"
	case SizeSmall:
		return "small"
	case SizeMedium:
		return "medium"
	case SizeLarge:
		return "large"
	case SizeXL:
		return "x-large"
	case SizeFull:
		return "full"
	}
	return "medium"
}

func ParseSize(s string) Size {
	switch strings.ToLower(s) {
	case "x-small", "xsmall", "xs":
		return SizeXS
	case "small", "sm", "s":
		return SizeSmall
	case "medium", "med", "m":
		return SizeMedium
	case "large", "lg", "l":
		return SizeLarge
	case "x-large", "xlarge", "xl":
		return SizeXL
	case "full", "f":
		return SizeFull
	}
	return SizeMedium
}

func (s Size) Int() int {
	if s != SizeFull {
		return int(s)
	}
	return 0
}

// https://stackoverflow.com/questions/52161555/how-to-custom-marshal-map-keys-in-json
func (s Size) MarshalText() ([]byte, error) { return []byte(s.String()), nil }

package network

import (
	"errors"
	"strings"
)

var ErrInvalidIPFamily = errors.New("ip family invalid")

type IPFamily string

const (
	IPFamilyInet  IPFamily = "inet"
	IPFamilyInet6 IPFamily = "inet6"
)

func ParseIPFamily(value string) (IPFamily, error) {
	switch IPFamily(strings.TrimSpace(value)) {
	case IPFamilyInet:
		return IPFamilyInet, nil
	case IPFamilyInet6:
		return IPFamilyInet6, nil
	default:
		return "", ErrInvalidIPFamily
	}
}

func ParseOptionalIPFamily(value *string) (*IPFamily, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // Nil means the caller did not provide an IP family preference.
	}

	ipFamily, err := ParseIPFamily(*value)
	if err != nil {
		return nil, err
	}

	return &ipFamily, nil
}

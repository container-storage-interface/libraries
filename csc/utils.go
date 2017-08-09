package main

import "regexp"

var addrRX = regexp.MustCompile(
	`(?i)^((?:(?:tcp|udp|ip)[46]?)|(?:unix(?:gram|packet)?))://(.+)$`)

// ParseProtoAddr parses a Golang network address.
func ParseProtoAddr(protoAddr string) (proto string, addr string, err error) {
	m := addrRX.FindStringSubmatch(protoAddr)
	if m == nil {
		return "", "", ErrInvalidCSIEndpoint
	}
	return m[1], m[2], nil
}

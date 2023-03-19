package utils

import "net"

type IP struct {
	V4             net.IP
	V4PrefixLength int    // valid if > 0
	V6             net.IP // valid if not empty
	V6PrefixLength int    // valid if > 0
}

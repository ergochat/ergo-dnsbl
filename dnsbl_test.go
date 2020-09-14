// Copyright (c) 2020 Shivaram Lingamneni
// released under the MIT license

package main

import (
	"net"
	"testing"
)

func assertCorrect(ip, expected string, t *testing.T) {
	ipaddr := net.ParseIP(ip)
	if ipaddr == nil {
		t.Errorf("invalid ip address string %s", ip)
	}

	reversed, _ := ReverseIP(ipaddr)
	if reversed != expected {
		t.Errorf("expected %s to reverse to %s, got %s", ipaddr.String(), expected, reversed)
	}

	if ipaddr.To4() == nil && len(reversed) != 64 {
		t.Errorf("%s is invalid, must have 64 characters", expected)
	}
}

func TestReverseIP(t *testing.T) {
	assertCorrect("1.2.3.4", "4.3.2.1.", t)
	assertCorrect("8.8.8.8", "8.8.8.8.", t)
	assertCorrect("255.254.253.252", "252.253.254.255.", t)

	assertCorrect("2001::0db8", "8.b.d.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.1.0.0.2.", t)
	assertCorrect("2620:1ec:c11::200", "0.0.2.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.1.1.c.0.c.e.1.0.0.2.6.2.", t)
}

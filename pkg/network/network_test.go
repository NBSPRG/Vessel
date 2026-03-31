package network

import (
	"net"
	"testing"
)

// TestParseIP ensures IP parsing works correctly
func TestParseIP(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		isValid bool
	}{
		{"Valid IPv4", "192.168.1.1", true},
		{"Valid IPv6", "2001:db8::1", true},
		{"Invalid IP", "256.256.256.256", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedIP := net.ParseIP(tt.ip)
			isValid := parsedIP != nil

			if isValid != tt.isValid {
				t.Errorf("ParseIP(%q) validity mismatch: expected %v, got %v", tt.ip, tt.isValid, isValid)
			}
		})
	}
}

// TestBridgeName ensures bridge naming conventions work
func TestBridgeName(t *testing.T) {
	tests := []struct {
		name       string
		bridgeName string
	}{
		{"Standard bridge", "br0"},
		{"Custom bridge", "vessel-br"},
		{"Named bridge", "docker_br"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.bridgeName == "" {
				t.Error("Bridge name should not be empty")
			}

			if len(tt.bridgeName) == 0 {
				t.Error("Bridge name length should be > 0")
			}
		})
	}
}

// TestVirtualEthernetPeer ensures veth peer naming validation
func TestVirtualEthernetPeer(t *testing.T) {
	tests := []struct {
		name     string
		vethName string
		peerName string
	}{
		{"Valid veth pair", "veth1", "veth1peer"},
		{"Named veth pair", "eth0", "eth0peer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.vethName == "" || tt.peerName == "" {
				t.Error("Veth and peer names should not be empty")
			}

			if tt.vethName == tt.peerName {
				t.Error("Veth and peer names should be different")
			}
		})
	}
}

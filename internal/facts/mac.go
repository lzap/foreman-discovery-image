package facts

import (
	"bytes"
	"encoding/json"
	"net"
	"strings"
)

// MAC is a hardware address (Ethernet 6-byte or InfiniBand 20-byte).
// It uses net.HardwareAddr as the underlying type. Use EmptyMAC for nil/zero/none.
type MAC net.HardwareAddr

// EmptyMAC is the zero value (nil), representing no interface.
var EmptyMAC MAC

// String returns the MAC in colon-separated lowercase form
// (e.g. "52:54:00:94:9e:52" for Ethernet, or 20 bytes for InfiniBand).
func (m MAC) String() string {
	if len(m) == 0 {
		return "00:00:00:00:00:00"
	}
	const hexDigit = "0123456789abcdef"
	var b strings.Builder
	b.Grow(len(m)*3 - 1)
	for i := 0; i < len(m); i++ {
		if i > 0 {
			b.WriteByte(':')
		}
		b.WriteByte(hexDigit[m[i]>>4])
		b.WriteByte(hexDigit[m[i]&0xf])
	}
	return b.String()
}

// IsEmpty reports whether m is nil or empty.
func (m MAC) IsEmpty() bool {
	return len(m) == 0
}

// Equal reports whether m and o have the same length and bytes.
func (m MAC) Equal(o MAC) bool {
	return bytes.Equal(m, o)
}

// HardwareAddr returns m as net.HardwareAddr (copy).
func (m MAC) HardwareAddr() net.HardwareAddr {
	return net.HardwareAddr(append([]byte(nil), m...))
}

// MACFromHardwareAddr creates a MAC from net.HardwareAddr (copy).
// Supports Ethernet (6-byte) and InfiniBand (20-byte). Returns EmptyMAC if hw is nil or empty.
func MACFromHardwareAddr(hw net.HardwareAddr) MAC {
	if len(hw) == 0 {
		return EmptyMAC
	}
	return MAC(append([]byte(nil), hw...))
}

// ParseMAC parses s as a MAC in colon or BOOTIF format. Returns error on failure.
// Accepts Ethernet (6-byte) and InfiniBand (20-byte) colon-separated strings.
func ParseMAC(s string) (MAC, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return EmptyMAC, nil
	}
	// BOOTIF / PXELinux: "01-52-54-00-94-9e-52" — drop first octet, then hex with - or :
	if strings.Contains(s, "-") {
		_, after, ok := strings.Cut(s, "-")
		if !ok {
			return EmptyMAC, &net.AddrError{Err: "invalid BOOTIF MAC", Addr: s}
		}
		s = strings.ReplaceAll(after, "-", ":")
	}
	hw, err := net.ParseMAC(s)
	if err != nil {
		return EmptyMAC, err
	}
	return MAC(append([]byte(nil), hw...)), nil
}

// MustParseMAC parses s as a MAC like ParseMAC but panics on error.
func MustParseMAC(s string) MAC {
	m, err := ParseMAC(s)
	if err != nil {
		panic(err)
	}
	return m
}

// ParseBOOTIF parses a PXELinux BOOTIF value (e.g. "01-52-54-00-94-9e-52").
// Returns EmptyMAC on parse error.
func ParseBOOTIF(s string) MAC {
	m, _ := ParseMAC(s)
	return m
}

// MarshalJSON encodes the MAC as a colon-separated string.
func (m MAC) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

// UnmarshalJSON decodes a MAC from a JSON string (colon or dash format).
func (m *MAC) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseMAC(s)
	if err != nil {
		return err
	}
	*m = parsed
	return nil
}

// MACSliceFromHardwareAddrs converts a slice of net.HardwareAddr to []MAC.
func MACSliceFromHardwareAddrs(hws []net.HardwareAddr) []MAC {
	out := make([]MAC, len(hws))
	for i, hw := range hws {
		out[i] = MACFromHardwareAddr(hw)
	}
	return out
}

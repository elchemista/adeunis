package main

import (
	"encoding/base64"
	"math"
	"testing"
)

// encode helper
func encode(raw []byte) string {
	return base64.StdEncoding.EncodeToString(raw)
}

// approxEqual checks float equality within epsilon.
func approxEqual(a, b, eps float64) bool {
	return math.Abs(a-b) <= eps
}

func TestDecodeInvalidBase64(t *testing.T) {
	d := NewDecoder()
	_, err := d.Decode("not_base64!!")
	if err == nil {
		t.Fatal("expected error for invalid base64, got nil")
	}
}

func TestDecodeTooShort(t *testing.T) {
	d := NewDecoder()
	short := make([]byte, 5)
	b64 := encode(short)
	_, err := d.Decode(b64)
	if err == nil || err.Error() != "payload too short: need at least 7 bytes" {
		t.Fatalf("expected payload too short error, got %v", err)
	}
}

func TestDecodeUnsupportedChannelType(t *testing.T) {
	// raw with type nibbles 3 (unsupported) at offsets
	raw := make([]byte, 10)
	// set type nibble 3 at offset 2 and 6
	raw[2] = 0x03
	raw[6] = 0x03
	b64 := encode(raw)
	d := NewDecoder()
	_, err := d.Decode(b64)
	if err == nil {
		t.Fatal("expected unsupported channel type error, got nil")
	}
}

func TestDecodeValidPayload(t *testing.T) {
	// Construct raw payload of length 10
	raw := make([]byte, 10)
	// Channel A: offset 2, type=1, value=0x0000F424 (62500)
	raw[2] = 0x01
	raw[3] = 0x00
	raw[4] = 0xF4
	raw[5] = 0x24
	// Channel B: offset 6, type=2, value=0x000003E8 (1000)
	raw[6] = 0x02
	raw[7] = 0x00
	raw[8] = 0x03
	raw[9] = 0xE8
	b64 := encode(raw)

	d := NewDecoder()
	metrics, err := d.Decode(b64)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Expect metrics: V, I, W, R, mA
	expectedNames := []string{"V", "I", "W", "R", "mA"}
	if len(metrics) != len(expectedNames) {
		t.Fatalf("expected %d metrics, got %d", len(expectedNames), len(metrics))
	}
	// Precomputed values
	v := math.Round((62500.0/1e6)*1000) / 1000 // 0.063
	i := v * 5                                 // 0.315
	w := i * 220                               // 69.3
	rv := 220 / i                              // ~698.412698
	ma := math.Round((1000.0/1e5)*1000) / 1000 // 0.01

	expectedValues := []float64{v, i, w, rv, ma}
	for idx, m := range metrics {
		if m.Name != expectedNames[idx] {
			t.Errorf("metric %d name = %q; want %q", idx, m.Name, expectedNames[idx])
		}
		if !approxEqual(m.Value, expectedValues[idx], 1e-6) {
			t.Errorf("metric %q value = %v; want %v", m.Name, m.Value, expectedValues[idx])
		}
	}
}

package main

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

// Measurement holds a named metric value.
type Measurement struct {
	Name  string
	Value float64
}

// Decoder decodes two-channel Adeunis analog payloads.
type Decoder struct{}

// NewDecoder returns a fresh Decoder.
func NewDecoder() *Decoder {
	return &Decoder{}
}

// Decode takes a base64-encoded payload and returns measurements for channel A and B.
func (d *Decoder) Decode(b64 string) ([]Measurement, error) {
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("base64 decode error: %w", err)
	}
	if len(raw) < 7 {
		return nil, errors.New("payload too short: need at least 7 bytes")
	}

	// parse channel A at offset 2
	m := make([]Measurement, 0)
	if err := decodeChannel(raw, 2, &m); err != nil {
		return nil, fmt.Errorf("channel A decode error: %w", err)
	}

	// parse channel B at offset 6
	if err := decodeChannel(raw, 6, &m); err != nil {
		return nil, fmt.Errorf("channel B decode error: %w", err)
	}
	return m, nil
}

// decodeChannel reads type and value at given offset and appends to metrics.
func decodeChannel(raw []byte, off int, metrics *[]Measurement) error {
	if off+4 > len(raw) {
		return errors.New("not enough bytes for uint32")
	}
	// type is low nibble of first byte
	typ := raw[off] & 0x0F
	// read uint32 BE from offset
	val := binary.BigEndian.Uint32(raw[off:]) & 0x00FFFFFF
	value := float64(val)

	switch typ {
	case 1:
		// voltage measurement (value in microvolts)
		v := math.Round((value/1e6)*1000) / 1000
		*metrics = append(*metrics, Measurement{"V", v})
		i := v * 5
		*metrics = append(*metrics, Measurement{"I", i})
		w := i * 220
		*metrics = append(*metrics, Measurement{"W", w})
		r := 220 / i
		*metrics = append(*metrics, Measurement{"R", r})
	case 2:
		// current in mA (value in units of 0.1µA)
		mA := math.Round((value/1e5)*1000) / 1000
		*metrics = append(*metrics, Measurement{"mA", mA})
	default:
		// unsupported type
		return fmt.Errorf("unsupported channel type %d", typ)
	}
	return nil
}

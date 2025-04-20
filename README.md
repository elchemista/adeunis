# Adeunis decoder

- **Decoder** with `Decode(b64 string) ([]Measurement, error)`
- **Measurement** struct for results
- Parses two channels at offsets 2 and 6
- Handles voltage (type 1) and current (type 2) with clear unit conversions

## Features

- **Zero external deps** – only Go’s standard library.
- **Two channels** – parses both channel A (offset 2) and channel B (offset 6).
- **Voltage & Current** – outputs V, I, W, R for voltage mode; mA for current mode.
- **Extensible** – add custom channel types easily.

## Installation

```bash
go get github.com/elchemista/adeunis
```

Or copy the `adeunis` folder into your project.

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/elchemista/adeunis"
)

func main() {
    payload := \"YOUR_BASE64_PAYLOAD_HERE\"

    decoder := adeunis.NewDecoder()
    metrics, err := decoder.Decode(payload)
    if err != nil {
        log.Fatalf(\"Decode error: %v\", err)
    }

    fmt.Println(\"Decoded metrics:\")
    for _, m := range metrics {
        fmt.Printf(\" • %s = %.3f\\n\", m.Name, m.Value)
    }
}
```

## How It Works

1. **Base64 decode** the payload.
2. **Channel A**: read type (`payload[2] & 0x0F`) and 24‑bit value.
3. **Channel B**: same at `payload[6]`.
4. For **type 1**: convert microvolts → volts, calculate current (×5), power (×220), resistance.
5. For **type 2**: convert to milliamps.

## Extending

To support more channel types, update `decodeChannel()`:

```go
switch typ {
case 1:
    // existing
case 2:
    // existing
case 3:
    // new type logic
default:
    return fmt.Errorf(\"unsupported channel type %d\", typ)
}
```

## License

MIT License – see [LICENSE](LICENSE) for details.

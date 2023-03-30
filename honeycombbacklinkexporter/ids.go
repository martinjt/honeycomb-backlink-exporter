package honeycombbacklinkexporter

import (
	"encoding/binary"
	"fmt"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

const (
	traceIDShortLength = 8
)

// getHoneycombID returns an ID suitable for use for spans and traces. Before
// encoding the bytes as a hex string, we want to handle cases where we are
// given 128-bit IDs with zero padding, e.g. 0000000000000000f798a1e7f33c8af6.
// To do this, we borrow a strategy from Jaeger [1] wherein we split the byte
// sequence into two parts. The leftmost part could contain all zeros. We use
// that to determine whether to return a 64-bit hex encoded string or a 128-bit
// one.
//
// [1]: https://github.com/jaegertracing/jaeger/blob/cd19b64413eca0f06b61d92fe29bebce1321d0b0/model/ids.go#L81
func getHoneycombTraceID(traceID pcommon.TraceID) string {
	// binary.BigEndian.Uint64() does a bounds check on traceID which will
	// cause a panic if traceID is fewer than 8 bytes. In this case, we don't
	// need to check for zero padding on the high part anyway, so just return a
	// hex string.

	var low uint64
	tID := traceID

	low = binary.BigEndian.Uint64(tID[traceIDShortLength:])
	if high := binary.BigEndian.Uint64(tID[:traceIDShortLength]); high != 0 {
		return fmt.Sprintf("%016x%016x", high, low)
	}

	return fmt.Sprintf("%016x", low)
}

// getHoneycombSpanID just takes a byte array and hex encodes it.
func getHoneycombSpanID(id pcommon.SpanID) string {
	if !id.IsEmpty() {
		return fmt.Sprintf("%x", id)
	}
	return ""
}

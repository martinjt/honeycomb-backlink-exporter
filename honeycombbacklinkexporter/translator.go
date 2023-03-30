package honeycombbacklinkexporter

import (
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// spanAttributesToMap converts an opencensus proto Span_Attributes object into a map
// of strings to generic types usable for sending events to honeycomb.
func spanAttributesToMap(spanAttrs pcommon.Map) map[string]interface{} {
	var attrs = make(map[string]interface{}, spanAttrs.Len())

	spanAttrs.Range(func(key string, value pcommon.Value) bool {
		switch value.Type() {
		case pcommon.ValueTypeStr:
			attrs[key] = value.AsString()
		case pcommon.ValueTypeBool:
			attrs[key] = value.Bool()
		case pcommon.ValueTypeInt:
			attrs[key] = value.Int()
		case pcommon.ValueTypeDouble:
			attrs[key] = value.Double()
		}
		return true
	})

	return attrs
}

// timestampToTime converts a protobuf timestamp into a time.Time.
func timestampToTime(ts pcommon.Timestamp) (t time.Time) {
	if ts == 0 {
		return
	}
	return time.Unix(0, int64(ts)).UTC()
}

// getStatusCode returns the status code
func getStatusCode(status ptrace.Status) int32 {
	return int32(status.Code())
}

// getStatusMessage returns the status message as a string
func getStatusMessage(status ptrace.Status) string {
	if len(status.Message()) > 0 {
		return status.Message()
	}

	return status.Code().String()
}

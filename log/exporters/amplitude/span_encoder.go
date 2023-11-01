package amplitude

import (
	"strconv"

	"github.com/renatoaf/amplitude-go/amplitude/data"
	"go.opencensus.io/trace"
)

// SpanEncoder is a function that knows how to transform a trace span into some
// amplitude data event.
type SpanEncoder func(s *trace.SpanData) *data.Event

// DefaultSpanEncoder is the default span encoder provided by this package.
func DefaultSpanEncoder(opts ...EncoderOption) SpanEncoder {
	o := defaultEncoderOptions(opts...)

	return func(s *trace.SpanData) *data.Event {
		if s.Attributes == nil {
			return nil
		}

		userID := getUserID(s) // the user's signature for messages with some "user_id" attribute

		var message string
		if len(s.Annotations) > 0 {
			message = s.Annotations[0].Message // the message logged
		}

		properties := make(map[string]interface{}, len(s.Attributes)+4)
		for k, v := range s.Attributes {
			properties[k] = v // all zap attributes added to the log entry
		}

		properties["service"] = o.app           // configured app name (e.g. "resilience-api", "domino")
		properties["eventName"] = s.Name        // span name (e.g. the name of the calling function)
		properties["message"] = message         // logged message
		properties["status"] = s.Status.Message // span status (e.g. "OK")

		id32, _ := strconv.ParseInt(s.TraceID.String(), 10, 32)
		id := int32(id32)

		return &data.Event{
			EventID:         id,
			Time:            s.StartTime.Unix(),
			EventType:       o.eventType,
			AppVersion:      o.version,
			UserID:          userID,
			EventProperties: properties,
		}
	}
}

func getUserID(s *trace.SpanData) string {
	val, ok := s.Attributes["user_id"]
	if ok {
		if userID, isString := val.(string); isString && userID != "" {
			return userID
		}
	}

	val, ok = s.Attributes["userID"]
	if ok {
		if userID, isString := val.(string); isString && userID != "" {
			return userID
		}
	}

	return ""
}

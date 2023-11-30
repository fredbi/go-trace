package amplitude

import (
	"strings"

	"go.opencensus.io/trace"
)

type (
	// Filter knows how to filter a trace span.
	Filter interface {
		// IsFiltered determines whether a trace should be skipped.
		IsFiltered(*trace.SpanData) bool
	}

	// Filters is a collection of Filter's that knows how to apply the IsFiltered operator.
	Filters []Filter

	// UserFilter filters out messages that are not signed with some "user_id" or "userID" attribute.
	UserFilter struct{}

	// InfoLevelFilter filters out debug messages.
	InfoLevelFilter struct{}

	// ContainsMessageFilter filters out spans that don't contain some message string.
	ContainsMessageFilter struct {
		message string
	}
)

// DefaultFilters is the default span filtering capability provided by this package.
func DefaultFilters() Filters {
	return Filters{
		NewUserFilter(),
		NewInfoLevelFilter(),
	}
}

var (
	_ Filter = &UserFilter{}
	_ Filter = &InfoLevelFilter{}
)

// NewUserFilter builds a UserFilter
func NewUserFilter() *UserFilter {
	return &UserFilter{}
}

// NewInfoLevelFilter builds an InfoLevelFilter
func NewInfoLevelFilter() *InfoLevelFilter {
	return &InfoLevelFilter{}
}

func (filters Filters) IsFiltered(s *trace.SpanData) bool {
	for _, filter := range filters {
		if filter.IsFiltered(s) {
			return true
		}
	}

	return false
}

func (f *UserFilter) IsFiltered(s *trace.SpanData) bool {
	if s.Attributes == nil {
		return true
	}

	if userID := getUserID(s); userID == "" {
		return true // non-user specific trace are filtered out
	}

	return false
}

func (f *InfoLevelFilter) IsFiltered(s *trace.SpanData) bool {
	if s.Attributes == nil {
		return true
	}

	if level := s.Attributes["level"]; level == "debug" {
		return true // debug messages are filtered out
	}

	return false
}

func (f *ContainsMessageFilter) IsFiltered(s *trace.SpanData) bool {
	if len(s.Annotations) < 1 {
		return true
	}

	if message := s.Annotations[0].Message; !strings.Contains(message, f.message) {
		return true
	}

	return false
}

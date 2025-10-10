package responsex

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

// Pagination contains pagination metadata and cursors.
type Pagination struct {
	Total  *int64  `json:"total,omitempty"`
	Limit  *int    `json:"limit,omitempty"`
	Offset *int    `json:"offset,omitempty"`
	Cursor *string `json:"cursor,omitempty"`
	Next   *string `json:"next,omitempty"`
	Prev   *string `json:"prev,omitempty"`
}

// cursorPayload is used for encode/decode of cursor values.
type cursorPayload struct {
	Offset int `json:"offset,omitempty"`
	Limit  int `json:"limit,omitempty"`
}

// EncodeCursor encodes a cursor payload to base64 JSON.
func EncodeCursor(offset, limit int) (string, error) {
	p := cursorPayload{Offset: offset, Limit: limit}
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// DecodeCursor decodes a base64 JSON cursor into offset/limit.
func DecodeCursor(s string) (int, int, error) {
	if s == "" {
		return 0, 0, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return 0, 0, err
	}
	var p cursorPayload
	if err := json.Unmarshal(b, &p); err != nil {
		return 0, 0, err
	}
	return p.Offset, p.Limit, nil
}

// ErrCursorExpired is returned when a cursor is no longer valid (optional helper)
var ErrCursorExpired = errors.New("cursor expired")

// CursorResetTime returns a time in future for rate limit reset header convenience.
func CursorResetTime(d time.Duration) time.Time { return time.Now().Add(d) }

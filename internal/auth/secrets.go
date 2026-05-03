package auth

import (
	"encoding/base64"
	"time"
)

const apiKey = "XoeWxarxOs1PKmZ2UkAnm8LfSjo29sei4P01NEbo"

var aest = mustLoadLocation("Australia/Brisbane")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic("auth: failed to load timezone " + name + ": " + err.Error())
	}
	return loc
}

// APIKey returns the static API key header value.
func APIKey() string {
	return apiKey
}

// BasicToken generates the Basic auth token using the current AEST time.
// Algorithm: AEST Unix ms → reverse digit string → base64-encode.
func BasicToken() string {
	ms := time.Now().In(aest).UnixMilli()
	return base64.StdEncoding.EncodeToString([]byte(reverseDigits(ms)))
}

func reverseDigits(n int64) string {
	s := []byte(itoa(n))
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return string(s)
}

// itoa converts an int64 to its decimal string without importing strconv
// to keep the dependency surface minimal. For negative values (should never
// happen with UnixMilli in practice) we still handle them correctly.
func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 20)
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}
	if neg {
		buf = append(buf, '-')
	}
	// buf is reversed relative to normal — flip it
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}

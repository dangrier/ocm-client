package auth

import (
	"encoding/base64"
	"testing"
)

func TestReverseDigits(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{12345, "54321"},
		{1000, "0001"},
		{1746230400000, "0000040326471"}, // a plausible AEST ms timestamp
	}
	for _, c := range cases {
		got := reverseDigits(c.in)
		if got != c.want {
			t.Errorf("reverseDigits(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBasicTokenIsBase64(t *testing.T) {
	token := BasicToken()
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		t.Fatalf("BasicToken() is not valid base64: %v", err)
	}
	if len(decoded) == 0 {
		t.Fatal("BasicToken() decoded to empty string")
	}
	// All characters in the decoded payload must be digits (the reversed ms timestamp).
	for _, b := range decoded {
		if b < '0' || b > '9' {
			t.Errorf("unexpected non-digit byte %q in decoded token", b)
		}
	}
}

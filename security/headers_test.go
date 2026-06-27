package security

import (
	"net/http"
	"testing"
)

func TestApplyFrameOptions(t *testing.T) {
	header := http.Header{}
	ApplyFrameOptions(header, FrameDeny)
	if got := header.Get(HeaderXFrameOptions); got != FrameDeny {
		t.Fatalf("X-Frame-Options = %q, want DENY", got)
	}
}

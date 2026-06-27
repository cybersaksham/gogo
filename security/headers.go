package security

import "net/http"

const (
	HeaderXFrameOptions = "X-Frame-Options"
	FrameDeny           = "DENY"
	FrameSameOrigin     = "SAMEORIGIN"
)

// ApplyFrameOptions sets X-Frame-Options when the response did not already choose one.
func ApplyFrameOptions(header http.Header, option string) {
	if option == "" || header.Get(HeaderXFrameOptions) != "" {
		return
	}
	header.Set(HeaderXFrameOptions, option)
}

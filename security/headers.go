package security

import (
	"fmt"
	"net/http"
)

const (
	HeaderXFrameOptions           = "X-Frame-Options"
	HeaderStrictTransportSecurity = "Strict-Transport-Security"
	HeaderContentTypeOptions      = "X-Content-Type-Options"
	HeaderReferrerPolicy          = "Referrer-Policy"
	HeaderCrossOriginOpenerPolicy = "Cross-Origin-Opener-Policy"
	HeaderSecureCookieDiagnostics = "X-Gogo-Secure-Cookie-Diagnostics"
	FrameDeny                     = "DENY"
	FrameSameOrigin               = "SAMEORIGIN"
)

// ApplyFrameOptions sets X-Frame-Options when the response did not already choose one.
func ApplyFrameOptions(header http.Header, option string) {
	if option == "" || header.Get(HeaderXFrameOptions) != "" {
		return
	}
	header.Set(HeaderXFrameOptions, option)
}

func ApplySecurityHeaders(header http.Header, options SecurityMiddlewareOptions, secure bool) {
	if secure && options.HSTSSeconds > 0 {
		value := "max-age=" + intString(options.HSTSSeconds)
		if options.HSTSIncludeSubdomains {
			value += "; includeSubDomains"
		}
		if options.HSTSPreload {
			value += "; preload"
		}
		header.Set(HeaderStrictTransportSecurity, value)
	}
	if options.ContentTypeNoSniff {
		header.Set(HeaderContentTypeOptions, "nosniff")
	}
	if options.ReferrerPolicy != "" {
		header.Set(HeaderReferrerPolicy, options.ReferrerPolicy)
	}
	if options.CrossOriginOpenerPolicy != "" {
		header.Set(HeaderCrossOriginOpenerPolicy, options.CrossOriginOpenerPolicy)
	}
	ApplyFrameOptions(header, options.FrameOptions)
}

func intString(value int) string {
	return fmt.Sprintf("%d", value)
}

package middleware

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// SecurityHeaders sets CSP, frame deny, nosniff, and referrer policy.
func SecurityHeaders(frontendURL, publicBaseURL string, secure bool) func(http.Handler) http.Handler {
	csp := contentSecurityPolicy(frontendURL, publicBaseURL)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("Content-Security-Policy", csp)
			h.Set("X-Frame-Options", "DENY")
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("Referrer-Policy", "no-referrer")
			h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			h.Set("X-DNS-Prefetch-Control", "off")
			if secure {
				h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	}
}

func contentSecurityPolicy(frontendURL, publicBaseURL string) string {
	connect := []string{"'self'"}
	img := []string{"'self'", "data:", "blob:", "https://www.gravatar.com", "https://secure.gravatar.com"}
	if origin := originOf(publicBaseURL); origin != "" {
		connect = append(connect, origin)
		img = append(img, origin)
	}
	if origin := originOf(frontendURL); origin != "" {
		connect = append(connect, origin)
	}
	return strings.Join([]string{
		"default-src 'self'",
		"base-uri 'self'",
		"form-action 'self'",
		"frame-ancestors 'none'",
		"object-src 'none'",
		"script-src 'self' 'unsafe-eval'",
		"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
		"font-src 'self' https://fonts.gstatic.com",
		fmt.Sprintf("img-src %s", strings.Join(img, " ")),
		fmt.Sprintf("connect-src %s", strings.Join(unique(connect), " ")),
	}, "; ")
}

func originOf(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return ""
	}
	scheme := u.Scheme
	if scheme == "" {
		scheme = "https"
	}
	return scheme + "://" + u.Host
}

func unique(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

package wsman

import (
	"net/http"
	"strings"
)

func CanHandle(r *http.Request) bool {
	if r.Method != "POST" {
		return false
	}

	if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/soap+xml") {
		return false
	}

	return true
}

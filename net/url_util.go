package net

import (
	urlpkg "net/url"
	"strings"
)

// EncodeURL 编码 url
func EncodeURL(url string) string {
	url = urlpkg.QueryEscape(url)
	return strings.ReplaceAll(url, "+", "%20")
}

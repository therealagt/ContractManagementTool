package webui

import (
	"embed"
	"net/http"
)

//go:embed review.html
var content embed.FS

func ReviewPage(w http.ResponseWriter, _ *http.Request) {
	data, _ := content.ReadFile("review.html")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}

package snapshot

import (
	"log/slog"
	"net/http"
	"strings"
)

func urlGen(baseUrl, cam string, secure bool) string {
	var snapShotURL strings.Builder
	var prefix string

	if secure {
		prefix = "https://"
	} else {
		prefix = "http://"
	}

	for _, block := range []string{prefix, baseUrl, "/api/", cam, "/latest.jpg?h=", "800"} {
		_, err := snapShotURL.WriteString(block)
		if err != nil {
			slog.Error("error creating snapshot url", slog.String("error", err.Error()))
			return ""
		}
	}
	slog.Info("snapshot url created", slog.String("URL", snapShotURL.String()))
	return snapShotURL.String()
}

func GetSnapshot(baseUrl, cam string, secure bool) *http.Response {
	url := urlGen(baseUrl, cam, secure)
	response, err := http.Get(url)

	slog.Info("fetching snapshot using", slog.String("URL", url))
	if err != nil {
		slog.Error("error fetching snapshot", slog.String("error", err.Error()))
		return nil
	}
	return response
}

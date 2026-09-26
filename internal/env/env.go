package env

import (
	"log/slog"
	"net/http"
	"os"
)

var Devmode bool

func init() {
	val, ok := os.LookupEnv("DEVMODE")
	Devmode = ok && val == "1"
	if Devmode {
		slog.Warn("Devmode enabled")
	}
}

var ConnectHttp *http.Client

func init() {
	if Devmode {
		var protocols http.Protocols
		protocols.SetUnencryptedHTTP2(true)
		httpClient := http.Client{
			Transport: &http.Transport{
				Protocols: &protocols,
			},
		}
		ConnectHttp = &httpClient
	} else {
		ConnectHttp = http.DefaultClient
	}
}

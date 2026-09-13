package env

import (
	"log/slog"
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

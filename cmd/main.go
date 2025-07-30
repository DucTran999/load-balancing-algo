package main

import (
	"flag"
	"os"

	"github.com/DucTran999/load-balancing-algo/internal/app"
	"github.com/rs/zerolog"
)

func main() {
	// Initialize zerolog with ConsoleWriter for pretty terminal output
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	appName := flag.String("app-name", "rr", "Load balance app to run")
	flag.Parse()

	switch *appName {
	case "rr":
		app.RunRoundRobinApp(logger)
	default:
		logger.Fatal().Msg("[ERROR] app not available")
	}
}

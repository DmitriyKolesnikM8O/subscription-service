package logger

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func SetLogrus(level string) {
	logrusLevel, err := log.ParseLevel(level)
	if err != nil {
		log.Warnf("Invalid log level %s, defaulting to Debug: %v", level, err)
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(logrusLevel)
	}

	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		PadLevelText:    true,
	})

	log.SetOutput(os.Stdout)
}

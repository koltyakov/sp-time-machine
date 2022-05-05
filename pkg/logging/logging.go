package logging

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
)

// InitLogger initialer logging configuration
func InitLogger(folderPath string, logFilename string, logLevel log.Level) {
	logFilePath := filepath.Join(folderPath, logFilename)
	_ = os.MkdirAll(folderPath, os.ModePerm)

	rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
		Filename:   logFilePath,
		MaxSize:    50, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
		Level:      logLevel,
		Formatter:  &log.JSONFormatter{},
	})
	if err != nil {
		log.Fatalf("failed to initialize file rotate hook: %v", err)
	}

	log.SetOutput(colorable.NewColorableStdout())

	log.SetFormatter(&reFormatter{log.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC822,
	}})

	log.SetLevel(logLevel)
	log.AddHook(rotateFileHook)
}

// Custom formatter for os.Stdout redundent data trim
type reFormatter struct {
	log.TextFormatter
}

func (formatter *reFormatter) Format(entry *log.Entry) ([]byte, error) {
	// Trim metadata object
	delete(entry.Data, "metadata")
	// Cut long messages
	maxLen := 80
	if len(strings.TrimSpace(entry.Message)) > maxLen {
		message := ""
		for _, w := range strings.Split(strings.TrimSpace(entry.Message), " ") {
			if len(message) <= maxLen {
				message += " " + w
			}
		}
		// entry.Message = strings.TrimSpace(string([]rune(entry.Message)[0:maxLen])) + " [check log]"
		entry.Message = message + "... [check logs]"
	}
	return formatter.TextFormatter.Format(entry)
}

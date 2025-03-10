package logger
import (
	"os"
	"time"
	"github.com/rs/zerolog"
)
var log zerolog.Logger
func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	log = zerolog.New(output).With().Timestamp().Caller().Logger()
}
func Info(message string, fields ...map[string]interface{}) {
	event := log.Info().Str("level", "info")
	if len(fields) > 0 {
		for k, v := range fields[0] {
			event = event.Interface(k, v)
		}
	}
	event.Msg(message)
}
func Error(message string, err error, fields ...map[string]interface{}) {
	event := log.Error().Str("level", "error").Err(err)
	if len(fields) > 0 {
		for k, v := range fields[0] {
			event = event.Interface(k, v)
		}
	}
	event.Msg(message)
}

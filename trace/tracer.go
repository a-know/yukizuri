package trace

import (
	"fmt"

	"github.com/hnakamur/ltsvlog"
)

type Tracer interface {
	TraceInfo(map[string]interface{})
	TraceError(map[string]interface{}, error)
	LogContent(string, string, string, string) map[string]interface{}
}

type tracer struct{}

func New() Tracer {
	return &tracer{}
}

func (t *tracer) LogContent(kind string, username string, remoteAddr string, message string) map[string]interface{} {
	content := map[string]interface{}{
		"kind":        kind,
		"username":    username,
		"remote_addr": remoteAddr,
		"message":     message,
	}
	return content
}

func logContent2Info(logContent map[string]interface{}) map[string]interface{} {
	logContent["err"] = "-"
	logContent["stack"] = "-"
	return logContent
}

func (t *tracer) TraceInfo(content map[string]interface{}) {
	log := ltsvlog.Logger.Info()
	content = logContent2Info(content)

	for k, v := range content {
		switch vi := v.(type) {
		case int:
			log.Int(k, vi)
		case string:
			log.String(k, vi)
		default:
			log.String(k, v.(string))
		}
	}
	log.Log()
}

func (t *tracer) TraceError(content map[string]interface{}, err error) {
	loggerError := ltsvlog.Err(err)

	for k, v := range content {
		switch vi := v.(type) {
		case int:
			loggerError.Int(k, vi)
		case string:
			loggerError.String(k, vi)
		default:
			loggerError.String(k, v.(string))
		}
	}

	loggerError.Stack("")
	ltsvlog.Logger.Err(
		ltsvlog.WrapErr(loggerError, func(err error) error {
			return fmt.Errorf("occurs error is: %v", err)
		}))
}

type nilTracer struct{}

func (t *nilTracer) LogContent(kind string, username string, remoteAddr string, message string) map[string]interface{} {
	return nil
}

func (t *nilTracer) TraceInfo(content map[string]interface{}) {}

func (t *nilTracer) TraceError(content map[string]interface{}, err error) {}

func Off() Tracer {
	return &nilTracer{}
}

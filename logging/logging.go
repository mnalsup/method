package logging

import (
	"encoding/json"
	"fmt"
	"os"

	"go.uber.org/zap"
)

var log *zap.SugaredLogger

var logLevels = []string{
	zap.DebugLevel.String(),
	zap.InfoLevel.String(),
	zap.WarnLevel.String(),
	zap.ErrorLevel.String(),
	zap.DPanicLevel.String(),
	zap.PanicLevel.String(),
	zap.FatalLevel.String(),
}

func contains(arr []string, val string) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

func initLogger() *zap.SugaredLogger {
	level := os.Getenv("METHOD_LOGGING_LEVEL")
	if !contains(logLevels, level) {
		level = zap.InfoLevel.String()
	}
	rawJSON := []byte(fmt.Sprintf(`{
	  "level": "%s",
	  "encoding": "console",
	  "outputPaths": ["stdout"],
	  "errorOutputPaths": ["stderr"],
	  "initialFields": {},
	  "encoderConfig": {
	    "messageKey": "message",
	    "levelKey": "level",
	    "levelEncoder": "lowercase"
	  }
	}`, level))

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	Logger, err := cfg.Build()
	if err != nil {
		panic(fmt.Sprintf("Unable to initialize logger: %s", err.Error()))
	}
	log = Logger.Sugar()
	log.Debugf("Logging level set to %s", level)
	return log
}

func GetLogger() *zap.SugaredLogger {
	if log != nil {
		return log
	} else {
		return initLogger()
	}
}

package logger

import (
	"fmt"
	"os"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gopkg.in/olivere/elastic.v5"
	"gopkg.in/sohlich/elogrus.v2"
)

//CLogger defined
type CLogger struct {
	Log *logrus.Logger
}

type LogInfo struct {
	method     string
	path       string
	clientIP   string
	userAgent  string
	dataLength int
	function   string
	line       int
}

//Logger interface
type Logger interface {
	Debug(direction string, i *LogInfo, msg string)
	Info(direction string, i *LogInfo, msg string)
	Error(direction string, i *LogInfo, msg string)
	Fatal(direction string, i *LogInfo, msg string)
	GetLogger() *logrus.Logger
}

//InitLogger used to get Logger componnet
func InitLogger(fileName, elasticDBEndpoint string) *CLogger {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	log := logrus.New()
	log.Level = logrus.DebugLevel

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err == nil {
		log.Out = file
	} else {
		log.Info("Failed to log to file, using default stderr")
	}

	client, err := elastic.NewClient(elastic.SetURL(elasticDBEndpoint))
	if err != nil {
		log.Info("Failed to connect elasticsearch server: " + elasticDBEndpoint)
	} else {
		hostName, _ := os.Hostname()
		hook, _ := elogrus.NewElasticHook(client, hostName, logrus.InfoLevel, "ionian_api")
		log.Hooks.Add(hook)
	}

	return &CLogger{
		log}
}

func (logger *CLogger) GetLogger() *logrus.Logger {
	return logger.Log
}

func (logger *CLogger) Debug(direction string, i *LogInfo, msg string) {
	logger.Log.WithFields(logrus.Fields{
		"method":     i.method,
		"path":       i.path,
		"direction":  direction,
		"clientIP":   i.clientIP,
		"userAgent":  i.userAgent,
		"dataLength": i.dataLength,
		"func":       i.function,
		"line":       i.line,
	}).Debug(msg)
}

func (logger *CLogger) Info(direction string, i *LogInfo, msg string) {
	logger.Log.WithFields(logrus.Fields{
		"method":     i.method,
		"path":       i.path,
		"direction":  direction,
		"clientIP":   i.clientIP,
		"userAgent":  i.userAgent,
		"dataLength": i.dataLength,
		"func":       i.function,
		"line":       i.line,
	}).Info(msg)
}

func (logger *CLogger) Error(direction string, i *LogInfo, msg string) {
	logger.Log.WithFields(logrus.Fields{
		"method":     i.method,
		"path":       i.path,
		"direction":  direction,
		"clientIP":   i.clientIP,
		"userAgent":  i.userAgent,
		"dataLength": i.dataLength,
		"func":       i.function,
		"line":       i.line,
	}).Error(msg)
}

func (logger *CLogger) Fatal(direction string, i *LogInfo, msg string) {
	logger.Log.WithFields(logrus.Fields{
		"method":     i.method,
		"path":       i.path,
		"direction":  direction,
		"clientIP":   i.clientIP,
		"userAgent":  i.userAgent,
		"dataLength": i.dataLength,
		"func":       i.function,
		"line":       i.line,
	}).Fatal(msg)
}

func Log(logger Logger, level string, direction string, i *LogInfo, msg string) {
	switch level {
	case "Debug":
		logger.Debug(direction, i, msg)
	case "Info":
		logger.Info(direction, i, msg)
	case "Error":
		logger.Error(direction, i, msg)
	case "Fatal":
		logger.Fatal(direction, i, msg)
	default:
		fmt.Printf("no such log level %s.", level)
	}
}

func Trace() *LogInfo {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	_, line := f.FileLine(pc[0])
	return &LogInfo{
		function: f.Name(),
		line:     line,
	}
}

func BuildLogInfo(c *gin.Context) *LogInfo {

	dataLength := c.Writer.Size()
	if dataLength < 0 {
		dataLength = 0
	}
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	_, line := f.FileLine(pc[0])

	return &LogInfo{
		method:     c.Request.Method,
		path:       c.Request.URL.Path,
		clientIP:   c.ClientIP(),
		userAgent:  c.Request.UserAgent(),
		dataLength: dataLength,
		function:   f.Name(),
		line:       line,
	}
}

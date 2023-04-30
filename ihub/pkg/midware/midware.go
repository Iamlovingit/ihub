package midware

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"ihub/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func InitMidwares(r *gin.Engine) error {
	midwareMap := map[string]gin.HandlerFunc{
		"log":     LoggerToFile(),
		"trace":   Trace(),
		"inout":   InOut(),
		"auth":    Auth(),
		"approve": Approve(),
	}
	for _, mw := range config.GetConfig().Midwares {
		if f, ok := midwareMap[mw.Midware]; ok {
			r.Use(f)
		} else {
			return fmt.Errorf("midware %s is not exist.\n", mw.Midware)
		}
	}
	return nil
}

type LogFormatter struct {
	TimestampFormat string
	LevelDesc       []string
}

func (f *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.RFC3339
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	levelText := "INFO"
	if entry.Level != logrus.InfoLevel {
		levelText = f.LevelDesc[entry.Level]
	}

	b.WriteString(entry.Time.Format(timestampFormat))
	b.WriteByte(' ')
	b.WriteString(levelText)
	b.WriteByte(' ')
	b.WriteString(entry.Message)
	b.WriteByte('\n')

	return b.Bytes(), nil
}

// 日志中间件 将信息输出到标准输出和日志中
func LoggerToFile() gin.HandlerFunc {
	logger := logrus.New()
	logger.SetFormatter(&LogFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LevelDesc:       []string{"DEBUG", "INFO", "WARNING", "ERROR", "FATAL"},
	})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.DebugLevel)

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		c.Next()
		end := time.Now()
		latency := end.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		if raw != "" {
			path = path + "?" + raw
		}
		logger.WithFields(logrus.Fields{
			"status":     statusCode,
			"method":     method,
			"ip":         clientIP,
			"path":       path,
			"latency":    latency,
			"user-agent": c.Request.UserAgent(),
		}).Info()
	}
}

// Auth
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

// Approve
func Approve() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func InOut() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

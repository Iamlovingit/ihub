package logger

import (
	"bytes"
	"fmt"
	"hfproxy/pkg/config"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// 声明一个全局的logger
// Logger用于记录日志，可以记录到文件，也可以记录到控制台
// 提供了多种日志级别，可以根据需要设置日志级别
// 使用方法：
// logger.Getlog().Info("info")
// logger.Getlog().Debug("debug")
// logger.Getlog().Warn("warn")
// logger.Getlog().Error("error")
// logger.Getlog().Fatal("fatal")
// logger.Getlog().Panic("panic")

var Glog *logrus.Logger

// LogFormatter 自定义日志格式
// 2020-01-01 00:00:00 [INFO] [main.go:12 main.main] info
type LogFormatter struct{}

// Format 实现logrus.Formatter接口
// 将 logrus.Entry 类型的日志条目格式化为字节数组并返回。
func (f *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// - 参数：
	// - entry：logrus.Entry 类型的日志条目。
	// - 返回值：
	// - []byte：格式化后的日志条目的字节数组。
	// - error：如果格式化过程中出现错误，则返回一个非 nil 的错误。
	// - 作用：将 logrus.Entry 类型的日志条目格式化为字节数组并返回。
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	// - func (entry *Entry) Time.Format(layout string) string
	// - 参数：
	// - layout：时间格式字符串。
	// - 返回值：
	// - string：格式化后的时间字符串。
	// - 作用：将时间格式化为指定格式的字符串。
	timestampFormat := entry.Time.Format("2006-01-02 15:04:05")
	var newLog string

	// func (entry *Entry) HasCaller() bool
	// - 返回值：
	// - bool：如果日志条目中包含调用者信息，则返回 true；否则返回 false。
	// - 作用：检查日志条目中是否包含调用者信息。
	if entry.HasCaller() {
		// - func Base(path string) string
		// - 参数：
		// - path：文件路径。
		// - 返回值：
		// - string：文件名。
		// - 作用：返回文件路径中的文件名部分。
		fName := filepath.Base(entry.Caller.File)
		newLog = fmt.Sprintf("[%s] [%s] [%s:%d %s] %s\n",
			timestampFormat, entry.Level.String(), fName, entry.Caller.Line, entry.Caller.Function, entry.Message)
	} else {
		newLog = fmt.Sprintf("[%s] [%s] %s\n",
			timestampFormat, entry.Level.String(), entry.Message)
	}

	b.WriteString(newLog)
	return b.Bytes(), nil
}

func Getlog() *logrus.Logger {
	level, err := logrus.ParseLevel(config.GetConfig().LOG.Level)
	if err != nil {
		level = logrus.DebugLevel
	}
	logrus.SetLevel(level)
	logrus.SetReportCaller(true)
	f, _ := os.OpenFile(config.GetConfig().LOG.DefaultConfigName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)
	logrus.SetOutput(io.MultiWriter(os.Stdout, f))
	logrus.SetFormatter(&LogFormatter{})
	logrus.StandardLogger().SetNoLock()
	return logrus.StandardLogger()
}

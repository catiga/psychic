package ml

import (
	"eli/config"
	"io"
	"os"
	"path"

	"time"

	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func init() {
	os.MkdirAll(config.Get().Env.Log, 0755)
	Log.SetReportCaller(false) // 这个打印调用的文件路径 先false
	// 设置日志输出控制台样式
	Log.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})

	// Log.SetLevel(logrus.DebugLevel)

	// 按天分割
	logFileName := path.Join(config.Get().Env.Log, "eli") + ".%Y%m%d.log"
	// 配置日志分割
	logFileCut := LogFileCut(logFileName)
	writers := []io.Writer{logFileCut, os.Stdout}

	// 输出到控制台，方便定位到那个文件
	fileAndStdoutWriter := io.MultiWriter(writers...)
	gin.DefaultWriter = fileAndStdoutWriter
	Log.SetOutput(fileAndStdoutWriter)

	// // 为当前logrus实例设置消息输出格式为json格式。
	// // 同样地，也可以单独为某个logrus实例设置日志级别和hook，这里不详细叙述。
	// Log.Formatter = &logrus.TextFormatter{}

	// file, err := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	// if err != nil {
	// 	Log.Fatal("err")
	// }

	// //文件使用完之后关闭
	// // defer file.Close()

	// multiWriter := io.MultiWriter(os.Stdout, file)
	// Log.SetOutput(multiWriter)

	// Log.Info("hahhahahaha")

}

// 配置日志切割
// LogFileCut 日志文件切割
func LogFileCut(fileName string) *rotatelogs.RotateLogs {
	logier, err := rotatelogs.New(
		// 切割后日志文件名称
		fileName,
		//rotatelogs.WithLinkName(Current.LogDir),   // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(30*24*time.Hour),    // 文件最大保存时间
		rotatelogs.WithRotationTime(24*time.Hour), // 日志切割时间间隔
		//rotatelogs.WithRotationCount(3),
		//rotatelogs.WithRotationTime(time.Minute), // 日志切割时间间隔
	)
	if err != nil {
		panic(err)
	}
	lfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.InfoLevel:  logier,
		logrus.FatalLevel: logier,
		logrus.DebugLevel: logier,
		logrus.WarnLevel:  logier,
		logrus.ErrorLevel: logier,
		logrus.PanicLevel: logier,
	},
		// 设置分割日志样式
		&logrus.TextFormatter{})
	logrus.AddHook(lfHook)
	return logier
}

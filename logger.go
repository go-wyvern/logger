package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	config *LogConfig
	info   *LogInfo
	Mu     sync.Mutex
	Cache  *BackEndCache
}

var output = make(chan string, 1024)
var stop_log = make(chan bool)

func New(c *LogConfig) *Logger {
	logger := &Logger{config: c, info: NewLogInfo(), Cache: NewCache(c.CacheSize)}
	logger.Start()
	return logger
}

func (l *Logger) GetConfig() *LogConfig {
	return l.config
}

func (l *Logger) Start() {
	go func() {
		for {
			select {
			case out_byte := <-output:
				for _, line := range strings.Split(out_byte, "\n") {
					if line != "" {
						if l.config.LogPlace&ToFile != 0 && l.config.LogPlace&ToConsole != 0 {
							l.ToFileAndStdout([]byte(line))
						} else if l.config.LogPlace&ToConsole != 0 {
							l.ToConsole([]byte(line))
						} else if l.config.LogPlace&ToFile != 0 {
							l.ToFile([]byte(line))
						}
					}
				}
			case <-stop_log:
				break
			}

		}
	}()
}

func (l *Logger) Stop() {
	l.Cache.Sync() //先同步日志在关闭
	stop_log <- true
	close(output)
	close(stop_log)
	l.Cache.Stop()
}

func (l Logger) GetLogInfo() *LogInfo {
	return l.info
}

func (l Logger) WithField(key string, value interface{}) *Logger {
	info := NewLogInfo()
	info.Module = l.info.Module
	l.info = info.WithField(key, value)
	return &l
}

func (l Logger) WithFields(fields Fields) *Logger {
	info := NewLogInfo()
	info.Module = l.info.Module
	l.info = info.WithFields(fields)
	return &l
}

func (l *Logger) SetModule(module string) *Logger {
	l.info.Module = module
	return l
}

func (l *Logger) SetLogInfo(info *LogInfo) *Logger {
	module := l.info.Module
	l.info = info
	l.info.Module = module
	return l
}

func (l *Logger) Output(level string, s string) {
	now := time.Now()
	var file string
	var line int
	var ok bool
	_, file, line, ok = runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	_, filename := path.Split(file)
	l.Mu.Lock()
	defer l.Mu.Unlock()
	l.info.Filename = filename
	l.info.Line = line
	l.info.LogLevel = level
	l.info.LogTime = now.UTC().Format("2016-01-02T15:04:05")
	l.info.Message = s
	json_format, _ := json.Marshal(&l.info)
	if json_format[len(json_format)-1] != '\n' {
		json_format = append(json_format, []byte("\n")...)
	}
	//l.Cache.PushToCache(json_format)
	if l.Cache.Switch {
		l.Cache.PushToCache(json_format)
	} else {
		output <- string(json_format)
	}
}

func (l *Logger) Debug(format string, a ...interface{}) {
	if l.GetConfig().level >= Debug {
		l.Output("Debug", fmt.Sprintf(format, a...))
	}
}

func (l *Logger) Warn(format string, a ...interface{}) {
	if l.GetConfig().level >= Warn {
		l.Output("Warn", fmt.Sprintf(format, a...))
	}
}

func (l *Logger) Alert(format string, a ...interface{}) {
	if l.GetConfig().level >= Alert {
		l.Output("Alert", fmt.Sprintf(format, a...))
	}
}

func (l *Logger) Info(format string, a ...interface{}) {
	if l.GetConfig().level >= Info {
		l.Output("Info", fmt.Sprintf(format, a...))
	}
}

func (l *Logger) Error(format string, a ...interface{}) {
	if l.GetConfig().level >= Error {
		l.Output("Error", fmt.Sprintf(format, a...))
	}
}

func (l *Logger) ToConsole(jsondata []byte) error {
	if v, ok := l.GetConfig().PlaceContentType[ToConsole]; ok {
		return l.Writeplace(l.GetConfig().PlaceIoWriter[ToConsole], v, jsondata)
	}
	return fmt.Errorf("please set content type")
}

func (l *Logger) ToFile(jsondata []byte) error {
	if v, ok := l.GetConfig().PlaceContentType[ToFile]; ok {
		if fd, ok := l.GetConfig().PlaceIoWriter[ToFile]; ok {
			return l.Writeplace(fd, v, jsondata)
		} else {
			fmt.Errorf("please set IoWriter")
		}
	}
	return fmt.Errorf("please set content type")
}

func (l *Logger) ToFileAndStdout(jsondata []byte) error {
	err := l.ToConsole(jsondata)
	if err != nil {
		return err
	}
	return l.ToFile(jsondata)
}

func (l *Logger) Writeplace(iw io.Writer, cType ContentType, jsondata []byte) error {
	var strByte []byte
	var err error
	if cType == FormatJson {
		log_info := NewLogInfo()
		json.Unmarshal(jsondata, log_info)
		strByte, _ = log_info.FormatJson()
	} else if cType == FormatText {
		log_info := NewLogInfo()
		json.Unmarshal(jsondata, log_info)
		strByte = []byte(log_info.FormatText())
	} else {
		return fmt.Errorf("Unknow ContentType")
	}
	if strByte[len(strByte)-1] != '\n' {
		strByte = append(strByte, []byte("\n")...)
	}
	_, err = iw.Write(strByte)
	if err != nil {
		return err
	}
	return nil
}

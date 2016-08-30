package logger

import (
	"encoding/json"
	"fmt"
)

type LogInfo struct {
	LogLevel string    `json:"log_level"`
	Module   string  `json:"module"`
	LogTime  string `json:"time"`
	Filename string  `json:"filename"`
	Line     int   `json:"line"`
	Message  string   `json:"message"`

	Data     Fields `json:"data,omitempty"`
}

type Fields map[string]interface{}

func NewLogInfo() *LogInfo {
	info := new(LogInfo)
	info.Data = make(map[string]interface{})
	return info
}

func (info LogInfo) FormatJson() ([]byte, error) {
	info.Data["log_level"] = info.LogLevel
	info.Data["module"] = info.Module
	info.Data["time"] = info.LogTime
	info.Data["filename"] = info.Filename
	info.Data["line"] = info.Line
	info.Data["message"] = info.Message
	return json.Marshal(&info.Data)
}

func (info LogInfo) FormatText() string {
	fileandline := fmt.Sprintf("%s:%d", info.Filename, info.Line)
	model := fmt.Sprintf("<%s>", info.Module)
	text := fmt.Sprintf("%s [%s] %s  %s: %-25s", info.LogTime, fileandline, model, info.LogLevel, info.Message)
	for k, v := range info.Data {
		text = fmt.Sprintf("%s %s = %v", text, k, v)
	}
	return text
}

func (info *LogInfo) WithField(key string, value interface{}) *LogInfo {
	info.WithFields(Fields{key: value})
	return info
}

func (info *LogInfo) WithFields(fields Fields) *LogInfo {
	data := make(Fields, len(info.Data) + len(fields))
	for k, v := range info.Data {
		data[k] = v
	}
	for k, v := range fields {
		data[k] = v
	}
	info.Data = data
	return info
}

package logger

import (
	"io"
	"os"
)

type Place int
type ContentType string
type Level int

const (
	ToConsole Place = 1 << iota
	ToFile
)

const (
	FormatJson ContentType = "Json"
	FormatText ContentType = "Text"
)

const (
	Alert Level = iota
	Error
	Warn
	Info
	Debug
)

var LogLevels = map[string]Level{
	"Alert":   Alert,
	"Error":   Error,
	"Warn": Warn,
	"Info":    Info,
	"Debug":   Debug,
}

var defaultCacheSize = 1024

type LogConfig struct {
	LogPlace         Place
	level            Level
	PlaceContentType map[Place]ContentType
	PlaceIoWriter    map[Place]io.Writer
	CacheSize        int
}

func NewConfig(p Place) *LogConfig {
	c := new(LogConfig)
	c.LogPlace = p
	c.PlaceContentType = make(map[Place]ContentType)
	c.PlaceIoWriter = make(map[Place]io.Writer)
	c.PlaceIoWriter[ToConsole] = os.Stdout
	c.CacheSize = defaultCacheSize
	return c
}

func (c *LogConfig) SetCententType(place Place, contenttype ContentType) *LogConfig {
	c.PlaceContentType[place] = contenttype
	return c
}

func (c *LogConfig) SetIoWriter(place Place, io_writer io.Writer) *LogConfig {
	c.PlaceIoWriter[place] = io_writer
	return c
}

func (c *LogConfig) SetLevel(level Level) *LogConfig {
	c.level = level
	return c
}

func (c *LogConfig) GetLevel() Level {
	return c.level
}

func (c *LogConfig) SetCacheSize(size int) *LogConfig {
	c.CacheSize = size
	return c
}
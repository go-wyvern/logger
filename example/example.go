package main

import (
	"time"

	"github.com/go-wyvern/logger"
)

var DefaultLog *logger.Logger

func main() {
	var i int
	for {
		i++
		DefaultLog.WithFields(logger.Fields{
			"foo":"test1",
			"bar":"test2",
		}).Debug("test for log %d", i)
		time.Sleep(200 * time.Millisecond)
	}
}

func init() {
	cfg := logger.NewConfig(logger.ToConsole)
	cfg.SetCacheSize(2048) //可设置,也可以不设置,默认1024
	cfg.SetCententType(logger.ToConsole, logger.FormatJson)
	cfg.SetLevel(logger.Debug)
	DefaultLog = logger.New(cfg)
	DefaultLog.SetModule("example")
	DefaultLog.Cache.CacheMonitor()//监控缓存 可开可不开 不开注释即可
}
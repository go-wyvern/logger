package logger

import (
	"bytes"
	"bufio"
	"os"
	"fmt"
	"time"

	"github.com/gosuri/uiprogress"
)

type BackEndCache struct {
	Switch    bool
	CacheSize int
	Container *Container
	In        *bufio.Writer
}

type Container struct {
	NextContainer *Container
	Data          *bytes.Buffer
}

func NewContainer(size int) *Container {
	c := &Container{}
	cNext := &Container{}
	c.Data = bytes.NewBuffer(make([]byte, size))
	cNext.Data = bytes.NewBuffer(make([]byte, size))
	c.NextContainer = cNext
	return c
}

func NewCache(size int) *BackEndCache {
	container := NewContainer(size)
	w := bufio.NewWriterSize(container.Data, size)
	container.Data.Truncate(0)
	return &BackEndCache{
		CacheSize:size,
		Container:container,
		In: w,
		Switch:true,
	}
}

func (c *Container) Next() *Container {
	c.NextContainer.NextContainer = c
	return c.NextContainer
}

func (b *BackEndCache) Stop() {
	b.Container.Data.Reset()
	b.Container.NextContainer.Data.Reset()
	b.Switch = false
}

func (b *BackEndCache) CanWriter(s []byte) bool {
	return b.In.Buffered() + len(s) < b.CacheSize
}

func (b *BackEndCache) PushToCache(s []byte) (int, error) {
	if b.Switch {
		if b.CanWriter(s) {
			return b.In.Write(s)
		} else {
			b.Sync()
			return b.In.Write(s)
		}
	} else {
		return 0, fmt.Errorf("The cache was closed!")
	}
}

func (b *BackEndCache) CacheMonitor() {
	go func() {
		uiprogress.Start()
		container_bar := uiprogress.AddBar(b.CacheSize)
		container_bar.PrependFunc(func(b *uiprogress.Bar) string {
			return fmt.Sprintf("app: Container %d/%d", b.Current(), b.Total)
		})
		container_bar.AppendCompleted()
		container_bar.PrependElapsed()

		for {
			container_bar.Set(b.In.Buffered())
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func (b *BackEndCache) TimeoutFlush(timeout time.Duration) {
	done := make(chan bool, 1)
	go func() {
		b.In.Flush()
		done <- true
	}()
	select {
	case <-done:
		output <- string(b.Container.Data.Bytes())
	case <-time.After(timeout):
		fmt.Fprintln(os.Stderr, "Flush took longer than", timeout)
	}
}

func (b *BackEndCache) Convert() {
	b.Container = b.Container.Next()
	b.Container.Data.Truncate(0)
}

func (b *BackEndCache) Sync() {
	b.TimeoutFlush(10 * time.Second)
	b.In.Reset(b.Container.NextContainer.Data)
	b.Convert()
}


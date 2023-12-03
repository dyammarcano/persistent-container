package cache2you

import (
	"embed"
	"fmt"
	"github.com/caarlos0/log"
	"github.com/patrickmn/go-cache"
	"net/http"
	"strings"
	"time"
)

const (
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = 0
)

type (
	FS struct {
		assets map[string]Item
		fs     embed.FS
		ttl    time.Duration
	}

	Data struct {
		cc *cache.Cache
	}

	Item struct {
		name string
		data []byte
		ttl  int64
		mime string
	}
)

func NewCacheData(defaultExpiration, cleanupInterval time.Duration) *Data {
	return &Data{
		cc: cache.New(defaultExpiration, cleanupInterval),
	}
}

func (c *Data) Get(k string) (any, bool) {
	return c.cc.Get(k)
}

func (c *Data) Set(k string, x any, d time.Duration) {
	c.cc.Set(k, x, d)
}

func NewCacheFS(assetsFiles embed.FS, ttl time.Duration) *FS {
	c := &FS{
		ttl:    ttl,
		assets: make(map[string]Item),
		fs:     assetsFiles,
	}

	go func() {
		for {
			for k, v := range c.assets {
				if v.ttl < time.Now().Unix() {
					delete(c.assets, k)
					log.Infof("cache expired for %s", k)
				}
			}

			time.Sleep(ttl)
		}
	}()

	return c
}

func (c *FS) AssetFile(name string) ([]byte, string, error) {
	return c.RootFile(fmt.Sprintf("assets%s", name))
}

func (c *FS) RootFile(name string) ([]byte, string, error) {
	item, ok := c.assets[name]
	if !ok {
		data, err := c.fs.ReadFile(fmt.Sprintf("dist/%s", name))
		if err != nil {
			return nil, "", err
		}

		item = Item{
			name: name,
			data: data,
			mime: c.geMimeType(data, name),
			ttl:  time.Now().Add(c.ttl).Unix(),
		}

		c.assets[name] = item
		log.Infof("cache miss for %s", name)
	}

	return item.data, item.mime, nil
}

// workarounds for mime types http.DetectContentType() doesn't detect
func (c *FS) geMimeType(data []byte, name string) string {
	switch {
	case strings.HasSuffix(name, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(name, ".html"):
		return "text/html; charset=utf-8"
	case strings.HasSuffix(name, ".js"):
		return "application/javascript; charset=utf-8"
	case strings.HasSuffix(name, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(name, ".ico"):
		return "image/x-icon"
	default:
		return http.DetectContentType(data)
	}
}

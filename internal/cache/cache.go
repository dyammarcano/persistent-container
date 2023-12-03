package cache

import (
	"embed"
	"fmt"
	"github.com/caarlos0/log"
	"net/http"
	"strings"
	"time"
)

type (
	Cache struct {
		assets map[string]Item
		fs     embed.FS
		ttl    time.Duration
	}

	Item struct {
		name string
		data []byte
		ttl  int64
		mime string
	}
)

func NewCacheFS(assetsFiles embed.FS, ttl time.Duration) *Cache {
	c := &Cache{
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

func (c *Cache) AssetFile(name string) ([]byte, string, error) {
	return c.RootFile(fmt.Sprintf("assets%s", name))
}

func (c *Cache) RootFile(name string) ([]byte, string, error) {
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
func (c *Cache) geMimeType(data []byte, name string) string {
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

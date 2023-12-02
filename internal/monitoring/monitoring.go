package monitoring

import (
	"bytes"
	"context"
	"dataStore/internal/store"
	"embed"
	"fmt"
	"github.com/caarlos0/log"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"sync"
	"time"
)

//go:embed all:dist/*
var assetsFiles embed.FS

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

func NewCacheFS(ttl time.Duration) *Cache {
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
		return "image/svg+xml; charset=utf-8"
	case strings.HasSuffix(name, ".ico"):
		return "image/x-icon"
	default:
		return http.DetectContentType(data)
	}
}

//	type UI struct {
//		r *mux.Router
//	}
//
//	func NewUI(router *mux.Router) *UI {
//		return &UI{
//			r: router,
//		}
//	}
//
//	func (u *UI) AddRoute() {
//		u.r.PathPrefix("/assets/").Handler(http.StripPrefix(fmt.Sprintf("%s/assets/", baseDir), http.FileServer(http.FS(assetsFiles))))
//		u.r.Handle("/", http.StripPrefix(baseDir, http.FileServer(http.FS(assetsFiles))))
//	}
var mon *Monitoring

func init() {
	mon = newMonitoring()
}

type (
	Monitoring struct {
		wg     sync.WaitGroup
		err    chan error
		port   string
		router *gin.Engine
		ctx    context.Context
		db     *store.Store
		cache  *Cache
	}
)

func newMonitoring() *Monitoring {
	m := &Monitoring{
		wg:     sync.WaitGroup{},
		err:    make(chan error),
		port:   ":8080",
		ctx:    context.Background(),
		router: gin.New(),
		cache:  NewCacheFS(24 * time.Hour),
	}

	var err error
	m.db, err = store.NewStore(m.ctx, "dataStore.db")
	if err != nil {
		log.Fatal(err.Error())
	}

	m.router.Use(hacks(m.ctx)) // TODO: remove this
	m.router.Use(gin.Recovery())

	m.router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	return m
}

func hackAction(ctx context.Context, ticker *time.Ticker, persistLogsCh chan *http.Request) {
	defer func() {
		ticker.Stop()
		close(persistLogsCh)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if len(persistLogsCh) > 0 {
				for len(persistLogsCh) > 0 {
					data := bytes.Buffer{}

					request := <-persistLogsCh

					data.WriteString(fmt.Sprintf("Method: %s\n", request.Method))
					data.WriteString(fmt.Sprintf("URL: %s\n", request.URL))
					data.WriteString(fmt.Sprintf("Proto: %s\n", request.Proto))
					data.WriteString(fmt.Sprintf("Host: %s\n", request.Host))
					data.WriteString(fmt.Sprintf("RemoteAddr: %s\n", request.RemoteAddr))
					data.WriteString(fmt.Sprintf("RequestURI: %s\n", request.RequestURI))
					data.WriteString(fmt.Sprintf("Header: %s\n", request.Header))
					data.WriteString(fmt.Sprintf("Body: %s\n", request.Body))
					data.WriteString(fmt.Sprintf("ContentLength: %d\n", request.ContentLength))
					data.WriteString(fmt.Sprintf("TransferEncoding: %s\n", request.TransferEncoding))
					data.WriteString(fmt.Sprintf("Close: %t\n", request.Close))
					data.WriteString(fmt.Sprintf("Form: %s\n", request.Form))
					data.WriteString(fmt.Sprintf("PostForm: %s\n", request.PostForm))
					data.WriteString(fmt.Sprintf("MultipartForm: %v\n", request.MultipartForm))

					//data, err := json.Marshal(<-persistLogsCh)
					//if err != nil {
					//	log.Errorf("error marshalling request: %s", err.Error())
					//	return
					//}

					if err := mon.db.Put("requests", uuid.NewString(), data.Bytes()); err != nil {
						log.Errorf("error putting key/value pair: %s", err.Error())
					}
				}
			}
		}
	}
}

func hacks(ctx context.Context) gin.HandlerFunc {
	ticker := time.NewTicker(15 * time.Second)
	persistLogsCh := make(chan *http.Request, 10)

	go hackAction(ctx, ticker, persistLogsCh)

	return func(c *gin.Context) {
		persistLogsCh <- c.Request
		c.Next()
	}
}

func basicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		//user, pass, ok := c.Request.BasicAuth()
		//if !ok {
		//	c.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}
		//
		//if user != "admin" || pass != "admin" {
		//	c.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}

		log.Infof("restricted access to %s", c.Request.URL.Path)

		c.Next()
	}
}

func (m *Monitoring) routes() {
	v1 := m.router.Group("/api/v1", basicAuth())

	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "GET API endpoint hit [health]"})
		})

		v1.GET("/data", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "GET API endpoint hit [data]"})
		})

		v1.GET("/data/:id", func(c *gin.Context) {
			id := c.Param("id")
			c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("GET API endpoint hit [data/%s]", id)})
		})

		v1.POST("/data", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "POST API endpoint hit [data]"})
		})
	}

	{
		m.router.GET("/assets/*filepath", func(c *gin.Context) {
			path := c.Param("filepath")
			data, mime, err := m.cache.AssetFile(path)
			if err != nil {
				c.String(http.StatusNotFound, "Resource file not found")
				return
			}

			c.Data(http.StatusOK, mime, data)
		})

		m.router.GET("/", func(c *gin.Context) {
			data, mime, err := m.cache.RootFile("index.html")
			if err != nil {
				c.String(http.StatusNotFound, "File not found")
				return
			}

			c.Data(http.StatusOK, mime, data)
		})

		m.router.GET("/favicon.ico", func(c *gin.Context) {
			data, mime, err := m.cache.RootFile("favicon.ico")
			if err != nil {
				c.String(http.StatusNotFound, "favicon file not found")
				return
			}

			c.Data(http.StatusOK, mime, data)
		})
	}
}

func (m *Monitoring) StartServer() {
	m.routes()

	server := &http.Server{
		Addr:    m.port,
		Handler: m.router,
	}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.err <- server.ListenAndServe()
	}()
}

func (m *Monitoring) Error() <-chan error {
	return m.err
}

func (m *Monitoring) StopServer() error {
	return nil
}

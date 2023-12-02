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
	"sync"
	"time"
)

//go:embed all:dist/*
var assetsFiles embed.FS

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
	}
)

func newMonitoring() *Monitoring {
	m := &Monitoring{
		wg:     sync.WaitGroup{},
		err:    make(chan error),
		port:   ":8080",
		ctx:    context.Background(),
		router: gin.New(),
	}

	var err error
	m.db, err = store.NewStore(m.ctx, "dataStore.db")
	if err != nil {
		log.Fatal(err.Error())
	}

	m.router.Use(hacks())
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

func hacks() gin.HandlerFunc {
	ticker := time.NewTicker(5 * time.Second)
	persistLogsCh := make(chan []byte, 10)

	go func() {
		defer func() {
			ticker.Stop()
			close(persistLogsCh)
		}()

		for {
			select {
			case <-mon.ctx.Done():
				return
			case <-ticker.C:
				if len(persistLogsCh) > 0 {
					logs := make([][]byte, 0, len(persistLogsCh))
					for len(persistLogsCh) > 0 {
						logs = append(logs, <-persistLogsCh)
					}

					if err := mon.db.PutBatch("logs", uuid.NewString(), logs); err != nil {
						log.Errorf("error putting batch key/value pair: %s", err.Error())
					}
				}
			}
		}
	}()

	return func(c *gin.Context) {
		//get request and save it in a buffer
		buf := bytes.Buffer{}

		//process request
		c.Next()

		//get response and save it in a buffer
		buf.WriteString("\n")
	}
}

func (m *Monitoring) routes() {
	v1 := m.router.Group("/api/v1")

	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "GET API endpoint hit [health]"})
		})

		v1.GET("/data", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "GET API endpoint hit [data]"})
		})

		v1.GET("/data/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "GET API endpoint hit [data/:id]"})
		})

		v1.POST("/data", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "POST API endpoint hit [data]"})
		})
	}

	{
		m.router.GET("/assets/*filepath", func(c *gin.Context) {
			path := fmt.Sprintf("dist/assets%s", c.Param("filepath"))
			data, err := assetsFiles.ReadFile(path)
			if err != nil {
				c.String(http.StatusNotFound, "Resource file not found")
				return
			}

			c.Data(http.StatusOK, http.DetectContentType(data), data)
		})

		m.router.GET("/", func(c *gin.Context) {
			data, err := assetsFiles.ReadFile("dist/index.html")
			if err != nil {
				c.String(http.StatusNotFound, "File not found")
				return
			}

			c.Data(http.StatusOK, http.DetectContentType(data), data)
		})

		m.router.GET("/favicon.ico", func(c *gin.Context) {
			data, err := assetsFiles.ReadFile("dist/favicon.ico")
			if err != nil {
				c.String(http.StatusNotFound, "favicon file not found")
				return
			}

			c.Data(http.StatusOK, http.DetectContentType(data), data)
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

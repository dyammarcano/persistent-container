package monitoring

import (
	"bytes"
	"context"
	"dataStore/internal/cache"
	"dataStore/internal/monitoring/vue-project"
	"dataStore/internal/store"
	"fmt"
	"github.com/caarlos0/log"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"sync"
	"time"
)

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
		cache  *cache.Cache
	}
)

func newMonitoring() *Monitoring {
	m := &Monitoring{
		wg:     sync.WaitGroup{},
		err:    make(chan error),
		port:   ":8080",
		ctx:    context.Background(),
		router: gin.New(),
		cache:  cache.NewCacheFS(vue.AssetsFiles, 24*time.Hour),
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

	v1.GET("/health", m.healthHandler)
	v1.GET("/data", m.dataHandler)
	v1.GET("/data/:id", m.dataIDHandler)
	v1.POST("/data", m.postDataHandler)

	m.router.GET("/", m.rootHandler)
	m.router.GET("/favicon.ico", m.faviconHandler)
	m.router.GET("/vite.svg", m.svgHandler)
	m.router.GET("/assets/*filepath", m.assetsHandler)
}

func (m *Monitoring) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GET API endpoint hit [health]"})
}

func (m *Monitoring) dataHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GET API endpoint hit [data]"})
}

func (m *Monitoring) dataIDHandler(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("GET API endpoint hit [data/%s]", id)})
}

func (m *Monitoring) postDataHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "POST API endpoint hit [data]"})
}

func (m *Monitoring) assetsHandler(c *gin.Context) {
	path := c.Param("filepath")
	data, mime, err := m.cache.AssetFile(path)
	if err != nil {
		c.String(http.StatusNotFound, "Resource file not found")
		return
	}

	c.Data(http.StatusOK, mime, data)
}

func (m *Monitoring) rootHandler(c *gin.Context) {
	data, mime, err := m.cache.RootFile("index.html")
	if err != nil {
		c.String(http.StatusNotFound, "File not found")
		return
	}

	c.Data(http.StatusOK, mime, data)
}

func (m *Monitoring) faviconHandler(c *gin.Context) {
	data, mime, err := m.cache.RootFile("favicon.ico")
	if err != nil {
		c.String(http.StatusNotFound, "favicon file not found")
		return
	}

	c.Data(http.StatusOK, mime, data)
}

func (m *Monitoring) svgHandler(c *gin.Context) {
	data, mime, err := m.cache.RootFile("vite.svg")
	if err != nil {
		c.String(http.StatusNotFound, "favicon file not found")
		return
	}

	c.Data(http.StatusOK, mime, data)
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

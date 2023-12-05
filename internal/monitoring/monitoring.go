package monitoring

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/caarlos0/log"
	"github.com/dyammarcano/persistent-container/internal/algorithm/compression"
	"github.com/dyammarcano/persistent-container/internal/cache2you"
	vue "github.com/dyammarcano/persistent-container/internal/monitoring/ui-store"
	"github.com/dyammarcano/persistent-container/internal/owner"
	"github.com/dyammarcano/persistent-container/internal/store"
	"github.com/dyammarcano/persistent-container/internal/version"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	bucketNotFound = "bucket not found"
)

type (
	Monitoring struct {
		wg        sync.WaitGroup
		err       chan error
		port      string
		router    *gin.Engine
		ctx       context.Context
		db        *store.Store
		cacheFs   *cache2you.FS
		cacheData *cache2you.Data
	}
)

func NewMonitoring(ctx context.Context, db *store.Store, port int) *Monitoring {
	m := &Monitoring{
		wg:        sync.WaitGroup{},
		err:       make(chan error),
		port:      fmt.Sprintf(":%d", port),
		ctx:       ctx,
		router:    gin.New(),
		cacheFs:   cache2you.NewCacheFS(vue.AssetsFiles, 24*time.Hour),
		cacheData: cache2you.NewCacheData(5*time.Minute, 10*time.Minute),
		db:        db,
	}

	m.router.Use(m.hacks(m.ctx)) // for demo purposes, please don't do this in production
	m.router.Use(gin.Recovery())

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000/"}

	m.router.Use(cors.New(config))

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

func (m *Monitoring) hackAction(ctx context.Context, ticker *time.Ticker, persistLogsCh chan *http.Request) {
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

					if err := m.db.Put("requests", store.GenerateKey(), data.Bytes()); err != nil {
						log.Errorf("error putting key/value pair: %s", err.Error())
					}
				}
			}
		}
	}
}

func (m *Monitoring) hacks(ctx context.Context) gin.HandlerFunc {
	ticker := time.NewTicker(15 * time.Second)
	persistLogsCh := make(chan *http.Request, 10)

	go m.hackAction(ctx, ticker, persistLogsCh)

	return func(c *gin.Context) {
		persistLogsCh <- c.Request
		c.Next()
	}
}

func validateEmailFormat(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// authHandler this handler is used to generate a key for the user and password (for demo purposes, please don't do this in production)
func (m *Monitoring) authHandler(c *gin.Context) {
	// get user and password from request
	user, pass, ok := c.Request.BasicAuth()
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "user and password not found"})
		return
	}

	if user == "" || pass == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "user and password not found"})
		return
	}

	email := c.Request.Header.Get("email")
	if !validateEmailFormat(email) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "email header not found"})
		return
	}

	h := sha256.New()
	h.Write([]byte(user + pass))
	h.Write([]byte(email))

	key := fmt.Sprintf("%x", h.Sum(nil))

	if data, ok := m.cacheData.Get(key); ok {
		c.JSON(http.StatusCreated, gin.H{"token": data})
		return
	}

	token, err := owner.NewToken(user, pass, email, time.Now().Add(48*time.Hour).Unix())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data, err := token.Encode()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	m.cacheData.Set(key, data, cache2you.DefaultExpiration)

	c.JSON(http.StatusCreated, gin.H{"token": data})
}

// apiAuth is a middleware that checks if the request has a valid authorization token (for demo purposes, please don't do this in production)
func (m *Monitoring) apiAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		encToken := strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer ")
		if encToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			return
		}

		var token owner.Token
		obj, ok := m.cacheData.Get(encToken)
		if ok {
			token = obj.(owner.Token)

			if token.IsValid() {
				c.Set("bucket", token.Bucket)
				c.Next()
				return
			}
		}

		if err := token.Decode(encToken); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		if !token.IsValid() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("bucket", token.Bucket)
		m.cacheData.Set(encToken, token, cache2you.NoExpiration)
		c.Next()
	}
}

func (m *Monitoring) routes() {
	v1 := m.router.Group("/api/v1", m.apiAuth())

	v1.GET("/data", m.dataHandler)
	v1.GET("/data/:id", m.dataIDHandler)
	v1.POST("/data", m.postDataHandler)
	v1.DELETE("/data/:id", m.deleteIDHandler)

	m.router.GET("/", m.rootHandler)
	m.router.GET("/metrics", m.metricsHandler)
	m.router.GET("/health", m.healthHandler)
	m.router.GET("/version", m.versionHandler)
	m.router.GET("/authorization", m.authHandler)
	m.router.GET("/assets/*filepath", m.assetsHandler)

	m.router.GET("/favicon.ico", m.faviconHandler)
	m.router.GET("/vite.svg", m.svgHandler)
}

func (m *Monitoring) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"health": true})
}

func (m *Monitoring) versionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, version.GetVersion())
}

func (m *Monitoring) dataHandler(c *gin.Context) {
	bucket := c.GetString("bucket")
	if bucket == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": bucketNotFound})
		return
	}

	data, err := m.db.GetBucketKeys(bucket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (m *Monitoring) dataIDHandler(c *gin.Context) {
	id := c.Param("id")

	if len(id) != 36 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id length"})
		return
	}

	bucket := c.GetString("bucket")
	if bucket == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": bucketNotFound})
		return
	}

	data, err := m.db.Get(bucket, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dec, err := compression.DecompressData(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if dec == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bucket empty"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": string(dec)})
}

func (m *Monitoring) deleteIDHandler(c *gin.Context) {
	id := c.Param("id")

	if len(id) != 36 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id length"})
		return
	}

	bucket := c.GetString("bucket")
	if bucket == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": bucketNotFound})
		return
	}

	if err := m.db.DeleteKey(bucket, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (m *Monitoring) postDataHandler(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comp, err := compression.CompressData(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	bucket := c.GetString("bucket")
	if bucket == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": bucketNotFound})
		return
	}

	id := store.GenerateKey()
	if err = m.db.Put(bucket, id, comp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (m *Monitoring) assetsHandler(c *gin.Context) {
	path := c.Param("filepath")
	data, mime, err := m.cacheFs.AssetFile(path)
	if err != nil {
		c.String(http.StatusNotFound, "Resource file not found")
		return
	}

	c.Data(http.StatusOK, mime, data)
}

func (m *Monitoring) rootHandler(c *gin.Context) {
	data, mime, err := m.cacheFs.RootFile("index.html")
	if err != nil {
		c.String(http.StatusNotFound, "File not found")
		return
	}

	c.Data(http.StatusOK, mime, data)
}

func (m *Monitoring) metricsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, m.db.GetMetrics())
}

func (m *Monitoring) faviconHandler(c *gin.Context) {
	data, mime, err := m.cacheFs.RootFile("favicon.ico")
	if err != nil {
		c.String(http.StatusNotFound, "favicon file not found")
		return
	}

	c.Data(http.StatusOK, mime, data)
}

func (m *Monitoring) svgHandler(c *gin.Context) {
	data, mime, err := m.cacheFs.RootFile("vite.svg")
	if err != nil {
		c.String(http.StatusNotFound, "favicon file not found")
		return
	}

	c.Data(http.StatusOK, mime, data)
}

func (m *Monitoring) StartServer(fn func(err error)) {
	m.routes()

	server := m.createServer()

	go m.handleServerErrors(fn, server)

	m.err <- server.ListenAndServe()
}

func (m *Monitoring) createServer() *http.Server {
	return &http.Server{
		Addr:    m.port,
		Handler: m.router,
	}
}

func (m *Monitoring) handleServerErrors(fn func(err error), server *http.Server) {
	m.wg.Add(1)
	defer m.wg.Done()

	for {
		select {
		case <-m.ctx.Done():
			if err := server.Shutdown(m.ctx); err != nil {
				fn(err)
			}
		case err := <-m.err:
			fn(err)
		}
	}
}

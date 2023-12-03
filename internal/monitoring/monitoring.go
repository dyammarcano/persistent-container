package monitoring

import (
	"bytes"
	"context"
	"crypto/sha256"
	"dataStore/internal/cache2you"
	"dataStore/internal/monitoring/vue-project"
	"dataStore/internal/owner"
	"dataStore/internal/store"
	version "dataStore/internal/version/gen"
	"fmt"
	"github.com/caarlos0/log"
	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	bucketNotFound = "bucket not found"
)

var mon *Monitoring

func init() {
	ctx := context.TODO()
	databasePath, err := filepath.Abs("../../dataStore.db")
	if err != nil {
		log.Fatal(err.Error())
	}

	db, err := store.NewStore(ctx, databasePath)
	if err != nil {
		log.Fatal(err.Error())
	}

	mon = NewMonitoring(ctx, db)
}

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
		//tokens    map[string]owner.Token
	}
)

func NewMonitoring(ctx context.Context, db *store.Store) *Monitoring {
	m := &Monitoring{
		wg:        sync.WaitGroup{},
		err:       make(chan error),
		port:      ":8080",
		ctx:       ctx,
		router:    gin.New(),
		cacheFs:   cache2you.NewCacheFS(vue.AssetsFiles, 24*time.Hour),
		cacheData: cache2you.NewCacheData(5*time.Minute, 10*time.Minute),
		//tokens:    make(map[string]owner.Token),
		db: db,
	}

	m.router.Use(hacks(m.ctx)) // for demo purposes, please don't do this in production
	m.router.Use(gin.Recovery())

	//c := rsCors.Options{
	//	AllowedOrigins:   []string{"http://localhost:5173"}, // replace specific origin with your desired origin
	//	AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
	//	AllowedHeaders:   []string{"Origin", "Content-Type", "Accept"},
	//	AllowCredentials: true,
	//}

	m.router.Use(cors.Default())

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

					if err := mon.db.Put("requests", store.GenerateKey(), data.Bytes()); err != nil {
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

func validateEmailFormat(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// authHandler this handler is used to generate a key for the user and password (for demo purposes, please don't do this in production)
func (m *Monitoring) authHandler(c *gin.Context) {
	// get user and password from request
	user, pass, ok := c.Request.BasicAuth()
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "user and password not found"})
		return
	}

	if user == "" || pass == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "user and password not found"})
		return
	}

	email := c.Request.Header.Get("email")
	if !validateEmailFormat(email) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "email header not found"})
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := token.Encode()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	m.cacheData.Set(key, data, cache2you.DefaultExpiration)

	c.JSON(http.StatusCreated, gin.H{"token": data})
}

// apiAuth is a middleware that checks if the request has a valid authorization token (for demo purposes, please don't do this in production)
func (m *Monitoring) apiAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		encToken := strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer ")

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
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			return
		}

		if !token.IsValid() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
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

	m.router.GET("/", m.rootHandler)

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
		c.JSON(http.StatusInternalServerError, gin.H{"message": bucketNotFound})
		return
	}

	data, err := m.db.GetBucketKeys(bucket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (m *Monitoring) dataIDHandler(c *gin.Context) {
	id := c.Param("id")

	if len(id) != 36 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}

	bucket := c.GetString("bucket")
	if bucket == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"message": bucketNotFound})
		return
	}

	data, err := m.db.GetObject(bucket, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if data.Object == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "bucket empty"})
		return
	}

	c.JSON(http.StatusOK, data.Object)
}

func (m *Monitoring) postDataHandler(c *gin.Context) {
	if c.Request.Header.Get("Content-Type") != "application/json" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid content type"})
		return
	}

	var data any
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	bucket := c.GetString("bucket")
	if bucket == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"message": bucketNotFound})
		return
	}

	id := store.GenerateKey()
	if err := m.db.PutObject(bucket, id, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
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

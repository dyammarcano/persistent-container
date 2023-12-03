package monitoring

import (
	"bytes"
	"context"
	"crypto/sha256"
	"dataStore/internal/algorithm/encoding"
	"dataStore/internal/cache"
	"dataStore/internal/monitoring/vue-project"
	"dataStore/internal/store"
	version "dataStore/internal/version/gen"
	"fmt"
	"github.com/caarlos0/log"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	bucketNotFound = "bucket not found"
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
	Token struct {
		Bucket string `json:"bucket"`
		Expire int64  `json:"expire"`
	}

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

	databasePath, err := filepath.Abs("../../dataStore.db")
	if err != nil {
		log.Fatal(err.Error())
	}

	m.db, err = store.NewStore(m.ctx, databasePath)
	if err != nil {
		log.Fatal(err.Error())
	}

	m.router.Use(hacks(m.ctx)) // for demo purposes, please don't do this in production
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

func verifyToken(c *gin.Context) bool {
	encToken := strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer ")

	token := Token{}
	if err := encoding.DeserializeStruct(encToken, &token); err != nil {
		return false
	}

	if token.Expire < time.Now().Unix() {
		return false
	}

	c.Set("bucket", token.Bucket)

	return true
}

func formatBearerToken(c *gin.Context, data []byte) {
	tokenBearer := struct {
		Authorization string `json:"authorization"`
	}{
		Authorization: fmt.Sprintf("%s", data),
	}

	c.JSON(http.StatusOK, tokenBearer)
}

func generateKey(user, pass string) string {
	h := sha256.New()
	h.Write([]byte(user + pass))

	// output same format of uuid
	return fmt.Sprintf("%x-%x-%x-%x-%x", h.Sum(nil)[:4], h.Sum(nil)[4:6], h.Sum(nil)[6:8], h.Sum(nil)[8:10], h.Sum(nil)[10:16])
}

// apiAuth is a middleware that checks if the request has a valid authorization token (for demo purposes, please don't do this in production)
func apiAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if ok := verifyToken(c); !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		log.Infof("restricted access to %s", c.Request.URL.Path)

		c.Next()
	}
}

func (m *Monitoring) routes() {
	v1 := m.router.Group("/api/v1", apiAuth())

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

// authHandler this handler is used to generate a key for the user and password (for demo purposes, please don't do this in production)
func (m *Monitoring) authHandler(c *gin.Context) {
	// get user and password from request
	user, pass, ok := c.Request.BasicAuth()
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// generate key from user and password
	key := generateKey(user, pass)

	// check if key already exists in database
	data, err := m.db.Get("authorization", key)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if len(data) > 0 {
		formatBearerToken(c, data)
		return
	}

	token := Token{
		Bucket: key,
		Expire: time.Now().Add(24 * time.Hour).Unix(),
	}

	data, err = encoding.SerializeStruct(token)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// store key in database
	if err = m.db.Put("authorization", key, data); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	formatBearerToken(c, data)
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

	var data any
	if err := m.db.GetObject(bucket, id, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if data == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "bucket empty"})
		return
	}

	c.JSON(http.StatusOK, data)
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

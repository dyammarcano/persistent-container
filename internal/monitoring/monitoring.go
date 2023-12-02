package monitoring

import (
	"embed"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

const baseDir = "/dist"

//go:embed all:dist/*
var assetsFiles embed.FS

type UI struct {
	r *mux.Router
}

func NewUI(router *mux.Router) *UI {
	return &UI{
		r: router,
	}
}

func (u *UI) AddRoute() {
	u.r.PathPrefix("/assets/").Handler(http.StripPrefix(fmt.Sprintf("%s/assets/", baseDir), http.FileServer(http.FS(assetsFiles))))
	u.r.Handle("/", http.StripPrefix(baseDir, http.FileServer(http.FS(assetsFiles))))
}

//var mon *Monitoring
//
//func init() {
//	mon = newMonitoring()
//}
//
//type Monitoring struct {
//	wg     sync.WaitGroup
//	err    chan error
//	port   string
//	server *http.Server
//	router *mux.Router
//	ctx    context.Context
//}
//
//func newMonitoring() *Monitoring {
//	return &Monitoring{
//		wg:     sync.WaitGroup{},
//		err:    make(chan error),
//		port:   ":8080",
//		ctx:    context.Background(),
//		router: mux.NewRouter(),
//	}
//}
//
//func (m *Monitoring) Routes() {
//	m.router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusOK)
//		if _, err := w.Write([]byte("OK")); err != nil {
//			m.err <- err
//		}
//	})
//
//	m.router.PathPrefix("/assets/").Handler(http.StripPrefix(fmt.Sprintf("%s/assets/", baseDir), http.FileServer(http.FS(assetsFiles))))
//	m.router.Handle("/", http.StripPrefix(baseDir, http.FileServer(http.FS(assetsFiles))))
//
//	m.server = &http.Server{
//		Addr:    m.port,
//		Handler: m.router,
//	}
//}
//
//func (m *Monitoring) StartServer() {
//	m.Routes()
//
//	m.wg.Add(1)
//	go func() {
//		defer m.wg.Done()
//		m.err <- m.server.ListenAndServe()
//	}()
//}
//
//func (m *Monitoring) Error() <-chan error {
//	return m.err
//}
//
//func (m *Monitoring) StopServer() error {
//	return nil
//}

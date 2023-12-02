package monitoring

import (
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

//func TestNewUI(t *testing.T) {
//	router := mux.NewRouter()

//ui := NewUI(router)
//ui.AddRoute()

//router.PathPrefix("/assets/").Handler(http.StripPrefix(fmt.Sprintf("%s/assets/", baseDir), http.FileServer(http.FS(assetsFiles))))
//router.Handle("/", http.StripPrefix(baseDir, http.FileServer(http.FS(assetsFiles))))
//
//server := &http.Server{
//	Addr:    ":8080",
//	Handler: router,
//}
//
//if err := server.ListenAndServe(); err != nil {
//	t.Error(err)
//}
//}

func TestMonitoring_StartServer(t *testing.T) {
	mon.StartServer()

	<-mon.err
}

func TestMonitoring_API(t *testing.T) {
	mon.StartServer()

	// Mock the HTTP request
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Serve the request to our router
	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	assert.Equal(t, http.StatusOK, rr.Code, "status code differs")

	// Check the response body is what we expect.
	expected := `OK`
	assert.Equal(t, expected, rr.Body.String(), "response body differs")
}

// To test StopServer method, since it's not yet implemented,
// we just check that it doesn't return any error.
func TestMonitoring_StopServer(t *testing.T) {
	// criate a new route home that read from embedded file and load vue from dist/index.html and assets in dist/assets
	m := newMonitoring()

	err := m.StopServer()
	assert.NoError(t, err, "StopServer should not return an error")
}

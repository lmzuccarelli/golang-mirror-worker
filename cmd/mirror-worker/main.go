package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/lmzuccarelli/golang-mirror-worker/pkg/connectors"
	"github.com/lmzuccarelli/golang-mirror-worker/pkg/handlers"
	clog "github.com/lmzuccarelli/golang-mirror-worker/pkg/log"
	"github.com/lmzuccarelli/golang-mirror-worker/pkg/validator"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	CONTENTTYPE     string = "Content-Type"
	APPLICATIONJSON string = "application/json"
)

var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "golang_mirror_worker_http_duration_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"path"})
)

// prometheusMiddleware implements mux.MiddlewareFunc.
func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(CONTENTTYPE, APPLICATIONJSON)
		// use this for cors
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept-Language")
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
		next.ServeHTTP(w, r)
		timer.ObserveDuration()
	})
}

func startHttpServer(c connectors.Clients) *http.Server {
	srv := &http.Server{Addr: ":" + os.Getenv("SERVER_PORT")}

	r := mux.NewRouter()
	r.Use(prometheusMiddleware)
	r.Path("/api/v2/metrics").Handler(promhttp.Handler())

	r.HandleFunc("/api/v1/batch", func(w http.ResponseWriter, req *http.Request) {
		handlers.BatchPayloadHandler(w, req, c)
	}).Methods("POST", "OPTIONS")

	r.HandleFunc("/api/v2/isalive", handlers.IsAlive).Methods("GET")

	http.Handle("/", r)

	fmt.Println("[INFO] : service starting ", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		fmt.Println("Httpserver: ListenAndServe() error: " + err.Error())
		os.Exit(1)
	}
	return srv
}

func main() {
	log := clog.New("info")
	err := validator.ValidateEnvars(log)
	if err != nil {
		os.Exit(-1)
	}
	log.Level(os.Getenv("LOG_LEVEL"))
	cons := connectors.NewClientConnections(log)
	startHttpServer(cons)
}

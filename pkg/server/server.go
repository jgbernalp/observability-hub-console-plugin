package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	plugins "github.com/openshift/observability-ui-hub/pkg/api/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

var slog = logrus.WithField("module", "server")

type Config struct {
	Port             int
	CertFile         string
	PrivateKeyFile   string
	Features         map[string]bool
	StaticPath       string
	ConfigPath       string
	PluginConfigPath string
}

type PluginConfig struct {
	UseTenantInHeader               bool          `json:"useTenantInHeader,omitempty" yaml:"useTenantInHeader,omitempty"`
	IsStreamingEnabledInDefaultPage bool          `json:"isStreamingEnabledInDefaultPage,omitempty" yaml:"isStreamingEnabledInDefaultPage,omitempty"`
	LokiTenanLabelKey               string        `json:"lokiTenanLabelKey,omitempty" yaml:"lokiTenanLabelKey,omitempty"`
	Timeout                         time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	LogsLimit                       int           `json:"logsLimit,omitempty" yaml:"logsLimit,omitempty"`
}

func (pluginConfig *PluginConfig) MarshalJSON() ([]byte, error) {
	type Alias PluginConfig
	return json.Marshal(&struct {
		Timeout float64 `json:"timeout,omitempty"`
		*Alias
	}{
		Timeout: pluginConfig.Timeout.Seconds(),
		Alias:   (*Alias)(pluginConfig),
	})
}

func Start(cfg *Config) {
	config, err := rest.InClusterConfig()
	if err != nil {
		slog.WithError(err).Error("cannot get in cluster config")
		panic(err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		slog.WithError(err).Error("cannot create k8s client")
		panic(err)
	}

	router := setupRoutes(cfg, dynamicClient)
	router.Use(corsHeaderMiddleware(cfg))

	loggedRouter := handlers.LoggingHandler(slog.Logger.Out, router)

	// clients must use TLS 1.2 or higher
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	httpServer := &http.Server{
		Handler:      loggedRouter,
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		TLSConfig:    tlsConfig,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	if cfg.CertFile != "" && cfg.PrivateKeyFile != "" {
		slog.Infof("listening on https://:%d", cfg.Port)
		panic(httpServer.ListenAndServeTLS(cfg.CertFile, cfg.PrivateKeyFile))
	} else {
		slog.Infof("listening on http://:%d", cfg.Port)
		panic(httpServer.ListenAndServe())
	}
}

func setupRoutes(cfg *Config, dynamicClient *dynamic.DynamicClient) *mux.Router {
	r := mux.NewRouter()

	r.PathPrefix("/health").HandlerFunc(healthHandler())

	// serve plugin manifest according to enabled features
	r.Path("/plugin-manifest.json").Handler(manifestHandler(cfg))

	r.HandleFunc("/api/v1/plugins/enable", plugins.EnablePluginHandler(dynamicClient)).Methods("POST")
	r.HandleFunc("/api/v1/plugins/{name}", plugins.GetPluginHandler(dynamicClient)).Methods("GET")
	r.HandleFunc("/api/v1/plugins/{name}", plugins.DeletePluginHandler(dynamicClient)).Methods("DELETE")
	r.HandleFunc("/api/v1/plugins", plugins.ListPluginHandler(dynamicClient)).Methods("GET")

	// serve front end files
	r.PathPrefix("/").Handler(filesHandler(http.Dir(cfg.StaticPath)))

	return r
}

func filesHandler(root http.FileSystem) http.Handler {
	fileServer := http.FileServer(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filePath := r.URL.Path

		// disable caching for plugin entry point
		if strings.HasPrefix(filePath, "/plugin-entry.js") {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Expires", "0")
		}

		if strings.HasPrefix(filePath, "/plugin-manifest.json") {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Expires", "0")
		}

		fileServer.ServeHTTP(w, r)
	})
}

func healthHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
}

func corsHeaderMiddleware(cfg *Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headers := w.Header()
			headers.Set("Access-Control-Allow-Origin", "*")
			next.ServeHTTP(w, r)
		})
	}
}

package service

import (
	"demo/cmd"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/up1/microservices-workshop-with-go/common/monitoring"
	"github.com/up1/microservices-workshop-with-go/common/tracing"
)

type Server struct {
	cfg *cmd.Config
	r   *chi.Mux
	h   *Handler
}

func NewServer(cfg *cmd.Config, h *Handler) *Server {
	return &Server{cfg: cfg, h: h}
}

func (s *Server) Close() {

}

func (s *Server) Start() {
	logrus.Infof("Starting HTTP server on %v", ":"+s.cfg.Port)
	err := http.ListenAndServe(":"+s.cfg.Port, s.r)
	if err != nil {
		logrus.WithError(err).Fatal("error starting HTTP server")
	}
}

func (s *Server) SetupRoutes() {

	s.r = chi.NewRouter()
	s.r.Use(middleware.RequestID)
	s.r.Use(middleware.RealIP)
	s.r.Use(middleware.Logger)
	s.r.Use(middleware.Recoverer)
	s.r.Use(middleware.Timeout(time.Minute))

	// Sub-routers with monitoring
	s.r.Route("/accounts", func(r chi.Router) {
		r.With(Trace("Get_Account")).
			With(Monitor(s.cfg.Name, "GetAccount", "GET /accounts/{accountId}")).
			Get("/{accountId}", s.h.GetAccount)
	})

	s.r.Get("/health", s.h.HealthCheck)
	s.r.Get("/metrics", promhttp.Handler().ServeHTTP)

	logrus.Info("Successfully routes")
}

func Monitor(serviceName, routeName, signature string) func(http.Handler) http.Handler {
	summaryVec := monitoring.BuildSummaryVec(serviceName, routeName, signature)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			start := time.Now()
			next.ServeHTTP(rw, req)
			duration := time.Since(start)

			// Store duration of request
			summaryVec.WithLabelValues("duration").Observe(duration.Seconds())

			// Store size of response, if possible.
			size, err := strconv.Atoi(rw.Header().Get("Content-Length"))
			if err == nil {
				summaryVec.WithLabelValues("size").Observe(float64(size))
			}
		})
	}
}

func Trace(opName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			logrus.Infof("starting span for %v", opName)
			span := tracing.StartHTTPTrace(req, opName)
			ctx := tracing.UpdateContext(req.Context(), span)
			next.ServeHTTP(rw, req.WithContext(ctx))

			span.Finish()
			logrus.Infof("finished span for %v", opName)
		})
	}
}

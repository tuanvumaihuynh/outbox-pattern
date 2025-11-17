package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"

	"github.com/tuanvumaihuynh/outbox-pattern/internal/apperr"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/config"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/http/apierr"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/http/gen"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/http/metric"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/http/middleware"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/http/swagger"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/service"
)

var tracer = otel.Tracer("internal/http")

// Service represents the HTTP service.
type Service struct {
	cfg     config.HTTP
	logger  *slog.Logger
	metrics *metric.Metrics

	productSvc service.ProductService
}

type CleanupFunc func(ctx context.Context) error

func New(
	cfg config.HTTP,
	log *slog.Logger,
	productSvc service.ProductService,
) *Service {
	return &Service{
		cfg:        cfg,
		logger:     log.With(slog.String("service", "http")),
		metrics:    metric.New(),
		productSvc: productSvc,
	}
}

func (s *Service) Run(ctx context.Context) (CleanupFunc, error) {
	r := chi.NewRouter()
	s.RegisterMiddlewares(r)

	if s.cfg.Swagger {
		swagger.Register(r)
	}

	s.RegisterHandlers(r)

	return s.RunWithServer(ctx, r)
}

func (s *Service) RunWithServer(ctx context.Context, handler http.Handler) (CleanupFunc, error) {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", s.cfg.Port),
		Handler:           handler,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 16, // 64 KB
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	}, nil
}

func (s *Service) RegisterMiddlewares(r chi.Router) {
	r.Use(
		middleware.Recoverer(s.logger),
		middleware.Trace(tracer),
		middleware.Metrics(s.metrics),
		middleware.CorrelationID(),
		middleware.Cors(),
		middleware.Logging(s.logger),
	)
}

func (s *Service) RegisterHandlers(r chi.Router) {
	handler := s.newHandler()
	strictHandlers := gen.NewStrictHandlerWithOptions(
		handler,
		[]gen.StrictMiddlewareFunc{},
		gen.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  s.handleRequestError,
			ResponseErrorHandlerFunc: s.handleResponseError,
		},
	)

	gen.HandlerWithOptions(strictHandlers, gen.ChiServerOptions{
		BaseRouter:       r,
		ErrorHandlerFunc: s.handleResponseError,
		Middlewares:      []gen.MiddlewareFunc{},
	})

	r.Handle(middleware.MetricsPath, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
		ErrorLog: log.Default(),
	}))
}

func (s *Service) handleRequestError(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	err = apperr.ValidationErr.WrapParent(err)
	res := apierr.New(err)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.logger.WarnContext(r.Context(), "error encoding error request",
			slog.Any("error", err))
	}
}

func (s *Service) handleResponseError(w http.ResponseWriter, r *http.Request, err error) {
	res := apierr.New(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(res.StatusCode)

	logLevel := slog.LevelInfo
	if res.StatusCode >= 500 {
		logLevel = slog.LevelError
	} else if res.StatusCode >= 400 {
		logLevel = slog.LevelWarn
	}
	s.logger.Log(r.Context(), logLevel, "http response error", slog.Any("error", err))

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.logger.ErrorContext(r.Context(), "error encoding error response",
			slog.Any("error", err))
	}
}

var _ gen.StrictServerInterface = (*handler)(nil)

type handler struct {
	*productHandler
}

func (s *Service) newHandler() *handler {
	return &handler{
		productHandler: newProductHandler(s.productSvc),
	}
}

package http_server

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"net/http"
)

type Context interface {
	echo.Context
}

type HandlerFunc func(Context) error

type MiddlewareFunc func(HandlerFunc) HandlerFunc

type Config struct {
	Addr      string
	SecretKey string
}

type Server struct {
	config Config
	server *echo.Echo
	logger *zap.Logger
}

func (s *Server) OnStart(_ context.Context) error {
	go func(addr string) {
		err := s.server.Start(addr)
		if err != nil {
			s.logger.Fatal("failed to start server", zap.Error(err))
		}
	}(s.config.Addr)

	return nil
}

func (s *Server) OnStop(ctx context.Context) error {
	err := s.server.Shutdown(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) GET(path string, handler HandlerFunc, mws ...MiddlewareFunc) {
	s.Add(http.MethodGet, path, handler, mws...)
}

func (s *Server) POST(path string, handler HandlerFunc, mws ...MiddlewareFunc) {
	s.Add(http.MethodPost, path, handler, mws...)
}

func (s *Server) PUT(path string, handler HandlerFunc, mws ...MiddlewareFunc) {
	s.Add(http.MethodPut, path, handler, mws...)
}

func (s *Server) DELETE(path string, handler HandlerFunc, mws ...MiddlewareFunc) {
	s.Add(http.MethodDelete, path, handler, mws...)
}

func toEchoMiddlewareFunc(mws ...MiddlewareFunc) []echo.MiddlewareFunc {
	return lo.Map(mws, func(mw MiddlewareFunc, _ int) echo.MiddlewareFunc {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(ec echo.Context) error {
				return mw(func(c Context) error {
					return next(c)
				})(ec)
			}
		}
	})
}

func (s *Server) Any(path string, handler HandlerFunc, mws ...MiddlewareFunc) {
	s.server.Any(
		path,
		func(c echo.Context) error {
			return handler(c)
		},
		toEchoMiddlewareFunc(mws...)...,
	)
}

func (s *Server) Add(method string, path string, handler HandlerFunc, mws ...MiddlewareFunc) {
	s.server.Add(
		method,
		path,
		func(c echo.Context) error {
			return handler(c)
		},
		toEchoMiddlewareFunc(mws...)...,
	)
}

func (s *Server) Use(mws ...MiddlewareFunc) {
	s.server.Use(toEchoMiddlewareFunc(mws...)...)
}

func (s *Server) Pre(mws ...MiddlewareFunc) {
	s.server.Pre(toEchoMiddlewareFunc(mws...)...)
}

func (s *Server) Echo() *echo.Echo {
	return s.server
}

func NewServer(config Config, v *validator.Validate, logger *zap.Logger) *Server {
	s := echo.New()
	s.HideBanner = true
	s.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info("request",
				zap.String("URI", v.URI),
				zap.Int("status", v.Status),
			)

			return nil
		},
		HandleError: true,
	}))
	s.Use(middleware.Recover())
	s.Validator = &customValidator{validator: v}

	return &Server{
		config: config,
		server: s,
		logger: logger,
	}
}

type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

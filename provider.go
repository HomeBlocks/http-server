package server

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go.uber.org/zap"
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
		return errors.Wrap(err, "HTTPServer.OnStop")
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

func NewServer(config Config, logger *zap.Logger) *Server {
	echoServer := echo.New()
	echoServer.HideBanner = true
	echoServer.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		Skipper:        nil,
		BeforeNextFunc: nil,
		LogValuesFunc: func(_ echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info("request",
				zap.String("URI", v.URI),
				zap.Int("status", v.Status),
			)

			return nil
		},
		HandleError:      true,
		LogLatency:       false,
		LogProtocol:      false,
		LogRemoteIP:      false,
		LogHost:          false,
		LogMethod:        false,
		LogURI:           true,
		LogURIPath:       false,
		LogRoutePath:     false,
		LogRequestID:     false,
		LogReferer:       false,
		LogUserAgent:     false,
		LogStatus:        true,
		LogError:         false,
		LogContentLength: false,
		LogResponseSize:  false,
		LogHeaders:       nil,
		LogQueryParams:   nil,
		LogFormValues:    nil,
	}))
	echoServer.Use(middleware.Recover())

	return &Server{
		config: config,
		server: echoServer,
		logger: logger,
	}
}

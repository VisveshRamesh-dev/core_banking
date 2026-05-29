package server

import (
	"account/internal/conf"
	"account/internal/service"
	utmw "utils/middleware"

	accountv1 "github.com/visvesh-ramesh/corebank/v1/account"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

func NewHTTPServer(c *conf.Server, svc *service.AccountService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			utmw.RequestID(),
			utmw.Logging(logger),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	accountv1.RegisterAccountServiceHTTPServer(srv, svc)
	return srv
}

// Package v1 manages the different versions of the API.
package v1

import (
	"github.com/nhaancs/bhms/app/services/api/v1/handlers/checkgrp"
	"github.com/nhaancs/bhms/app/services/api/v1/handlers/usergrp"
	"github.com/nhaancs/bhms/business/core/user"
	"github.com/nhaancs/bhms/business/core/user/stores/usercache"
	"github.com/nhaancs/bhms/business/core/user/stores/userdb"
	mid2 "github.com/nhaancs/bhms/business/web/mid"
	"github.com/nhaancs/bhms/foundation/sms"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/nhaancs/bhms/business/web/auth"
	"github.com/nhaancs/bhms/foundation/logger"
	"github.com/nhaancs/bhms/foundation/web"
	"go.opentelemetry.io/otel/trace"
)

// Options represent optional parameters.
type Options struct {
	corsOrigin string
}

// WithCORS provides configuration options for CORS.
func WithCORS(origin string) func(opts *Options) {
	return func(opts *Options) {
		opts.corsOrigin = origin
	}
}

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	Build    string
	Shutdown chan os.Signal
	Log      *logger.Logger
	Auth     *auth.Auth
	DB       *sqlx.DB
	Tracer   trace.Tracer
	KeyID    string
	SMS      *sms.SMS
}

// APIMux constructs a http.Handler with all application routes defined.
func APIMux(cfg APIMuxConfig, options ...func(opts *Options)) http.Handler {
	const version = "v1"
	var opts Options
	for _, option := range options {
		option(&opts)
	}

	app := web.NewApp(
		cfg.Shutdown,
		cfg.Tracer,
		mid2.Logger(cfg.Log),
		mid2.Errors(cfg.Log),
		mid2.Metrics(),
		mid2.Panics(),
	)

	if opts.corsOrigin != "" {
		app.EnableCORS(mid2.Cors(opts.corsOrigin))
	}

	// -------------------------------------------------------------------------
	// Check routes
	checkHdl := checkgrp.New(cfg.Build, cfg.DB)
	app.HandleNoMiddleware(http.MethodGet, version, "/readiness", checkHdl.Readiness)
	app.HandleNoMiddleware(http.MethodGet, version, "/liveness", checkHdl.Liveness)

	// -------------------------------------------------------------------------
	// User routes
	usrCore := user.NewCore(cfg.Log, usercache.NewStore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB)))
	usrHdl := usergrp.New(usrCore, cfg.Auth, cfg.KeyID, cfg.SMS)
	app.Handle(http.MethodPost, version, "/users/register", usrHdl.Register)
	app.Handle(http.MethodPost, version, "/users/verify-otp", usrHdl.VerifyOTP)
	app.Handle(http.MethodGet, version, "/users/token", usrHdl.Token)

	return app
}

package app

import (
	"goonairplanes/core"
	"net/http"
)

func ConfigureMiddleware(app *core.GonAirApp) {
	// Initialize global middleware with logging
	app.Router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			app.Logger.InfoLog.Printf("🛡️ Global Middleware: Processing request for %s", r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})

	// Add core middleware
	app.Router.Use(core.LoggingMiddleware(app.Logger))
	app.Router.Use(core.RecoveryMiddleware(app.Logger))
	app.Router.Use(core.SecureHeadersMiddleware())

	// Configure CORS if enabled
	if app.Config.EnableCORS {
		app.Router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				app.Logger.InfoLog.Printf("🌐 CORS Middleware: Processing request from %s", r.Header.Get("Origin"))
				next.ServeHTTP(w, r)
			})
		})
		app.Router.Use(core.CORSMiddleware(app.Config.AllowedOrigins))
	}

	// Configure rate limiting if enabled
	if app.Config.RateLimit > 0 {
		app.Router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				app.Logger.InfoLog.Printf("⏱️ Rate Limiting Middleware: Processing request from %s", r.RemoteAddr)
				next.ServeHTTP(w, r)
			})
		})
		app.Router.Use(core.RateLimitMiddleware(app.Config.RateLimit))
	}

	// Example of route-specific middleware with logging
	app.Router.AddRoute("/dashboard", nil, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			app.Logger.InfoLog.Printf("🔒 Auth Middleware: Checking access for %s", r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}, core.AuthMiddleware(func(token string) bool {
		app.Logger.InfoLog.Printf("🔑 Token validation for dashboard access")
		return true
	}))

	// Example of API route with middleware and logging
	app.Router.AddAPIRoute("/api/secure", nil,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				app.Logger.InfoLog.Printf("🔒 API Auth Middleware: Checking access for %s", r.URL.Path)
				next.ServeHTTP(w, r)
			})
		},
		core.AuthMiddleware(func(token string) bool {
			app.Logger.InfoLog.Printf("🔑 Token validation for API access")
			return true
		}),
		core.RateLimitMiddleware(10),
	)

	app.Logger.InfoLog.Printf("✅ Middleware configuration completed")
}

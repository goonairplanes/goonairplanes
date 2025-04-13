package core

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type AppLogger struct {
	InfoLog  *log.Logger
	ErrorLog *log.Logger
	WarnLog  *log.Logger
}

type GonAirApp struct {
	Router      *Router
	FileWatcher *FileWatcher
	Config      *Config
	Logger      *AppLogger
}

func NewApp() *GonAirApp {
	logger := &AppLogger{
		InfoLog:  log.New(os.Stdout, "✈️ \033[36mINFO\033[0m  ", log.Ldate|log.Ltime),
		ErrorLog: log.New(os.Stderr, "✈️ \033[31mERROR\033[0m ", log.Ldate|log.Ltime),
		WarnLog:  log.New(os.Stdout, "✈️ \033[33mWARN\033[0m  ", log.Ldate|log.Ltime),
	}

	router := NewRouter(logger)

	return &GonAirApp{
		Router: router,
		Config: &AppConfig,
		Logger: logger,
	}
}

func (app *GonAirApp) Init() error {
	startTime := time.Now()

	app.Logger.InfoLog.Printf("Initializing Go on Airplanes...")

	err := app.Router.InitRoutes()
	if err != nil {
		app.Logger.ErrorLog.Printf("Failed to initialize routes: %v", err)
		return fmt.Errorf("failed to initialize routes: %w", err)
	}
	app.Logger.InfoLog.Printf("Routes initialized successfully")

	if AppConfig.InMemoryJS {
		app.Logger.InfoLog.Printf("Initializing JavaScript library cache...")
		if err := FetchAndCacheJSLibraries(); err != nil {
			app.Logger.WarnLog.Printf("Failed to cache JavaScript libraries: %v", err)
			app.Logger.WarnLog.Printf("Falling back to CDN for JavaScript libraries")
		} else {
			app.Logger.InfoLog.Printf("JavaScript libraries cached successfully")
		}
	}

	configureMiddleware := app.getConfigureMiddlewareFunc()
	if configureMiddleware != nil {
		configureMiddleware(app)
		app.Logger.InfoLog.Printf("Middleware configured successfully")
	}

	if app.Config.DevMode && app.Config.LiveReload {
		watcher, err := NewFileWatcher(app.Router, app.Logger)
		if err != nil {
			app.Logger.ErrorLog.Printf("Failed to create file watcher: %v", err)
			return fmt.Errorf("failed to create file watcher: %w", err)
		}
		app.Logger.InfoLog.Printf("File watcher created successfully")
		app.FileWatcher = watcher
	}

	elapsedTime := time.Since(startTime)
	app.Logger.InfoLog.Printf("Go on Airplanes initialized in %v", elapsedTime.Round(time.Millisecond))

	return nil
}

func (app *GonAirApp) getConfigureMiddlewareFunc() func(*GonAirApp) {
	middlewareConfigPath := filepath.Join(app.Config.AppDir, "middleware.go")
	if _, err := os.Stat(middlewareConfigPath); os.IsNotExist(err) {
		app.Logger.WarnLog.Printf("Middleware configuration file not found at %s", middlewareConfigPath)
		return nil
	}

	return func(app *GonAirApp) {
		app.Router.Use(LoggingMiddleware(app.Logger))
		app.Router.Use(RecoveryMiddleware(app.Logger))

		if app.Config.EnableCORS {
			app.Router.Use(CORSMiddleware(app.Config.AllowedOrigins))
		}

		if app.Config.SSGEnabled {
			app.Router.Use(SSGMiddleware(app.Logger))
			app.Logger.InfoLog.Printf("SSG enabled, static files will be generated in %s", app.Config.SSGDir)
		}
	}
}

func (app *GonAirApp) Start() error {
	if app.FileWatcher != nil {
		app.FileWatcher.Start()
		defer app.FileWatcher.Stop()
		app.Logger.InfoLog.Printf("Live reload enabled - watching for file changes")
	}

	port := app.Config.Port

	app.printBanner(port)

	app.printErrorSummary()

	http.DefaultTransport = &http.Transport{
		MaxIdleConns:        1024,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     0,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableCompression:  true,
	}

	mux := http.NewServeMux()

	if app.Config.DevMode && app.FileWatcher != nil {

		app.FileWatcher.RegisterSocketHandler(mux)
	}

	mux.Handle("/", app.Router)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,

		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,

		MaxHeaderBytes: 1 << 20,
	}

	app.Logger.InfoLog.Printf("Press Ctrl+C to stop the server")

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		app.Logger.ErrorLog.Printf("Server error: %v", err)
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func (app *GonAirApp) printErrorSummary() {

	var templateErrorCount int
	var templateErrors []string

	app.Router.Marley.mutex.RLock()
	for routePath, err := range app.Router.Marley.TemplateErrors {
		templateErrorCount++
		templateErrors = append(templateErrors, fmt.Sprintf("  • %s: %v", routePath, err))
	}
	app.Router.Marley.mutex.RUnlock()

	var pageErrorCount int
	var pageErrors []string

	pageErrorMap.Range(func(key, value interface{}) bool {
		routePath := key.(string)
		pageError := value.(*PageError)

		app.Router.Marley.mutex.RLock()
		_, alreadyReported := app.Router.Marley.TemplateErrors[routePath]
		app.Router.Marley.mutex.RUnlock()

		if !alreadyReported {
			pageErrorCount++
			errorDetails := pageError.Details
			if len(errorDetails) > 100 {
				errorDetails = errorDetails[:100] + "..."
			}
			pageErrors = append(pageErrors, fmt.Sprintf("  • %s: %s", routePath, errorDetails))
		}
		return true
	})

	var apiErrorCount int
	var apiErrors []string

	apiErrorMap.Range(func(key, value interface{}) bool {
		apiError := value.(*APIError)

		apiErrorCount++
		errorMsg := apiError.ErrorMsg
		if len(errorMsg) > 100 {
			errorMsg = errorMsg[:100] + "..."
		}
		apiErrors = append(apiErrors, fmt.Sprintf("  • %s %s: %s",
			apiError.Method, apiError.Path, errorMsg))
		return true
	})

	if templateErrorCount > 0 || pageErrorCount > 0 || apiErrorCount > 0 {
		fmt.Println()
		app.Logger.WarnLog.Printf("┌─────────────────────────────────────────────────┐")
		app.Logger.WarnLog.Printf("│               ERROR SUMMARY REPORT              │")
		app.Logger.WarnLog.Printf("└─────────────────────────────────────────────────┘")

		if templateErrorCount > 0 {
			app.Logger.WarnLog.Printf("Template Loading Errors: %d", templateErrorCount)
			for _, errMsg := range templateErrors {
				app.Logger.WarnLog.Printf("%s", errMsg)
			}
			fmt.Println()
		}

		if pageErrorCount > 0 {
			app.Logger.WarnLog.Printf("Page Rendering Errors: %d", pageErrorCount)
			for _, errMsg := range pageErrors {
				app.Logger.WarnLog.Printf("%s", errMsg)
			}
			fmt.Println()
		}

		if apiErrorCount > 0 {
			app.Logger.WarnLog.Printf("API Endpoint Errors: %d", apiErrorCount)
			for _, errMsg := range apiErrors {
				app.Logger.WarnLog.Printf("%s", errMsg)
			}
			fmt.Println()
		}

		app.Logger.WarnLog.Printf("Page routes with errors will display the error page when accessed.")
		app.Logger.WarnLog.Printf("API endpoints with errors will return error responses.")
		app.Logger.WarnLog.Printf("Fix the issues and restart the server to resolve them.")
		fmt.Println()
	} else if app.Config.LogLevel == "debug" {
		app.Logger.InfoLog.Printf("No template, page or API errors detected. All routes healthy!")
	}
}

func (app *GonAirApp) printBanner(port string) {
	banner := `
	██████╗  ██████╗   
	██╔════╝ ██╔═══██╗  
	██║  ███╗██║   ██║    
	██║   ██║██║   ██║   
	╚██████╔╝╚██████╔╝    
	╚═════╝  ╚═════╝      
`
	fmt.Print(banner)
	app.Logger.InfoLog.Printf("Go on Airplanes ready for takeoff!")
	app.Logger.InfoLog.Printf("Local:   http://localhost:%s", port)

	interfaces, _ := getNetworkInterfaces()
	if len(interfaces) > 0 {
		for _, ip := range interfaces {
			app.Logger.InfoLog.Printf("Network: http://%s:%s", ip, port)
		}
	}
}

func getNetworkInterfaces() ([]string, error) {
	var ips []string

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {

		if iface.Flags&net.FlagUp == 0 ||
			iface.Flags&net.FlagLoopback != 0 ||
			strings.Contains(iface.Name, "vmnet") ||
			strings.Contains(iface.Name, "vEthernet") ||
			strings.Contains(iface.Name, "vboxnet") {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}

			ips = append(ips, ip.String())
		}
	}

	return ips, nil
}

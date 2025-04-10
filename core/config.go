package core

type Config struct {
	AppDir         string
	StaticDir      string
	Port           string
	DevMode        bool
	LiveReload     bool
	DefaultCDNs    bool
	TailwindCDN    string
	JQueryCDN      string
	LayoutPath     string
	ComponentDir   string
	AppName        string
	Version        string
	LogLevel       string
	TemplateCache  bool
	EnableCORS     bool
	AllowedOrigins []string
	RateLimit      int
}

var AppConfig = Config{
	AppDir:         "app",
	StaticDir:      "static",
	Port:           "3000",
	DevMode:        true,
	LiveReload:     true,
	DefaultCDNs:    true,
	TailwindCDN:    "https://cdn.tailwindcss.com",
	JQueryCDN:      "https://code.jquery.com/jquery-3.7.1.min.js",
	LayoutPath:     "app/layout.html",
	ComponentDir:   "app/components",
	AppName:        "Go on Airplanes",
	Version:        "0.3.0",
	LogLevel:       "info",
	TemplateCache:  true,
	EnableCORS:     false,
	AllowedOrigins: []string{"*"},
	RateLimit:      100,
}

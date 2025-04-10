# 🛫 Go on Airplanes Framework

<div align="center">
  <img src="img/goonairplane2.png" alt="Go on Airplanes Logo" width="180" />
  <br><br>
  <img src="https://img.shields.io/badge/Go-1.18+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go version"/>
  <img src="https://img.shields.io/badge/License-MIT-blue?style=for-the-badge" alt="License"/>
  <img src="https://img.shields.io/badge/Status-Alpha-orange?style=for-the-badge" alt="Status"/>
</div>

Go on Airplanes is a lightweight, high-performance fullstack web framework for Go with HTML file-based routing. It's designed for simplicity, speed, and a great developer experience.

I created Go on Airplanes 'cause I got tired of writing a whole damn colossus in Next.js just to build a simple CRUD.

> ✨ Zero configuration to start. Soar through development with ease.

## ✈️ Key Features

- **File-Based Routing** - Create pages by adding HTML files to your app directory
- **Component System** - Build reusable UI components with Go templates
- **Hot Reloading** - Changes refresh automatically in development mode
- **Performance Optimized** - Concurrent template loading and efficient caching
- **Minimal Dependencies** - No bloated external packages
- **Modern UI Support** - Tailwind CSS and jQuery included by default
- **Developer-Friendly Logs** - Clear, colorful console outputs
- **Zero Build Process** - Just write Go and HTML - no transpilation needed

## 🚀 Quick Start

1. Clone this repository
```bash
git clone https://github.com/yourusername/goonairplanes.git
cd goonairplanes
```

2. Run the server
```bash
go run main.go
```

3. View your site at `http://localhost:3000`

## 📂 Project Structure

```
project/
├── main.go                # Application entry point
├── core/                  # Framework internals
│   ├── app.go             # Application setup and lifecycle
│   ├── config.go          # Configuration
│   ├── marley.go          # Template rendering engine
│   ├── router.go          # Request handling and routing
│   └── watcher.go         # File watching for hot reload
├── app/                   # Your application
│   ├── layout.html        # Base layout template
│   ├── index.html         # Homepage ("/")
│   ├── about.html         # About page ("/about")
│   ├── dashboard/         # Dashboard section
│   │   └── index.html     # Dashboard homepage ("/dashboard")
│   ├── user/[id]/         # Dynamic route with parameters
│   │   └── index.html     # User profile page ("/user/123")
│   ├── components/        # Reusable UI components
│   │   ├── navbar.html    # Navigation component
│   │   └── card.html      # Card component
│   └── api/               # API endpoints
│       └── users/         # Users API
│           └── route.go   # Handler for "/api/users"
├── static/                # Static assets
│   ├── css/               # Stylesheets
│   ├── js/                # JavaScript files
│   └── images/            # Image assets
└── go.mod                 # Go module definition
```

## 📑 Page Creation

### Basic Pages

Create HTML files in the `app` directory to define routes:

- `app/about.html` → `/about`
- `app/contact.html` → `/contact`
- `app/blog/index.html` → `/blog`
- `app/blog/post.html` → `/blog/post`

### Dynamic Routes

Create folders with names in square brackets for dynamic segments:

- `app/product/[id]/index.html` → `/product/123`, `/product/abc`
- `app/blog/[category]/[slug].html` → `/blog/tech/go-web-dev`

Access parameters in templates:
```html
<h1>Product: {{.Params.id}}</h1>
```

### Nested Routes

Organize routes in subfolders for better structure:
```
app/
├── dashboard/
│   ├── index.html         # "/dashboard"
│   ├── settings.html      # "/dashboard/settings"
│   └── analytics/
│       └── index.html     # "/dashboard/analytics"
```

## 🧩 Components & Templates

### Creating Components

Define reusable components in the `app/components` directory:

```html
<!-- app/components/alert.html -->
{{define "alert"}}
<div class="bg-yellow-100 border-l-4 border-yellow-500 p-4 mb-4">
  <p class="font-bold">Note</p>
  <p>{{.}}</p>
</div>
{{end}}
```

### Using Components

Include components in your pages:

```html
<!-- app/index.html -->
{{define "content"}}
  <h1>Welcome to Go on Airplanes</h1>
  
  {{template "alert" "This framework is currently in alpha."}}
  
  <p>Start building your application!</p>
{{end}}
```

### Layout Template

The `app/layout.html` file defines the base layout used by all pages:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Go on Airplanes</title>
  {{if .Config.DefaultCDNs}}
    <script src="{{.Config.TailwindCDN}}"></script>
    <script src="{{.Config.JQueryCDN}}"></script>
  {{end}}
</head>
<body class="bg-gray-100 min-h-screen">
  <main class="container mx-auto py-6 px-4">
    {{template "content" .}}
  </main>
</body>
</html>
```

## 🔧 Configuration

Edit `core/config.go` to modify framework behavior:

```go
var AppConfig = Config{
  AppDir:        "app",
  StaticDir:     "static",
  Port:          "3000",
  DevMode:       true,        // Set to false in production
  LiveReload:    true,        // Hot reload in development
  DefaultCDNs:   true,        // Use Tailwind and jQuery CDNs
  AppName:       "Go on Airplanes",
  Version:       "0.1.0",
  LogLevel:      "info",      // Options: debug, info, warn, error
  TemplateCache: true,        // Cache templates for better performance
}
```

## 🔌 API Routes

Create API endpoints by placing Go files in the `app/api` directory:

```go
// app/api/hello/route.go
package main

import (
  "encoding/json"
  "net/http"
  "time"
)

func Handler(w http.ResponseWriter, r *http.Request) {
  response := map[string]interface{}{
    "message": "Hello from Go on Airplanes API!",
    "time":    time.Now().Format(time.RFC3339),
  }
  
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(response)
}
```

## 🧰 Advanced Features

### Custom Error Pages

Create specialized error pages:
- `app/404.html` - Custom not found page
- `app/500.html` - Server error page

### Environment Variables

Set configuration through environment variables:
```bash
PORT=8080 go run main.go
```

### Static File Serving

All files in the `static` directory are served at `/static/`:
```html
<img src="/static/images/logo.png">
<link rel="stylesheet" href="/static/css/styles.css">
<script src="/static/js/app.js"></script>
```

Example favicon implementation in layout.html:
```html
<link rel="icon" type="image/png" href="/static/favicon.ico">
```

## 📈 Performance Tips

- Enable template caching in production by setting `TemplateCache: true` in your config
- The framework uses concurrent template loading for faster startup times
- Keep components small and focused for better reusability and performance
- Marley template engine caches components for efficient rendering
- Set appropriate LogLevel in production (`"info"` or `"error"`) to reduce logging overhead
- Static assets are efficiently served through dedicated file server handlers

## 📜 License

MIT

---

<div align="center">
  <p>Built with ❤️ by the Jklee</p>
</div> 
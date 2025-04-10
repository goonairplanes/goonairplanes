# 🚀 Getting Started

This guide will help you get up and running with Go on Airplanes quickly.

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Project Structure](#project-structure)
- [First Steps](#first-steps)
- [Next Steps](#next-steps)

## Installation

1. **Prerequisites**
   - Go 1.18 or later
   - Git

2. **Install the Framework**
   ```bash
   git clone https://github.com/yourusername/goonairplanes.git
   cd goonairplanes
   ```

3. **Verify Installation**
   ```bash
   go run main.go
   ```
   Visit `http://localhost:3000` to see the default page.

## Quick Start

1. **Create a New Page**
   ```bash
   mkdir -p app/about
   touch app/about/index.html
   ```

2. **Add Content**
   ```html
   <!-- app/about/index.html -->
   {{define "content"}}
   <h1>About Us</h1>
   <p>Welcome to our about page!</p>
   {{end}}
   ```

3. **Add a Component**
   ```html
   <!-- app/components/alert.html -->
   {{define "alert"}}
   <div class="bg-yellow-100 p-4">
     {{.}}
   </div>
   {{end}}
   ```

4. **Use the Component**
   ```html
   <!-- app/about/index.html -->
   {{define "content"}}
   <h1>About Us</h1>
   {{template "alert" "This is an important message!"}}
   <p>Welcome to our about page!</p>
   {{end}}
   ```

## Project Structure

```
project/
├── main.go                # Application entry point
├── core/                  # Framework internals
│   ├── app.go             # Application setup
│   ├── config.go          # Configuration
│   ├── middleware.go      # Middleware system
│   ├── router.go          # Routing
│   └── watcher.go         # Hot reloading
├── app/                   # Your application
│   ├── layout.html        # Base layout
│   ├── index.html         # Homepage
│   ├── components/        # UI components
│   └── api/               # API endpoints
└── static/                # Static assets
```

## First Steps

1. **Configure Your App**
   Edit `core/config.go`:
   ```go
   var AppConfig = Config{
       AppName: "My App",
       Port: "8080",
       DevMode: true,
   }
   ```

2. **Set Up Middleware**
   Create `app/middleware.go`:
   ```go
   package app

   import "goonairplanes/core"

   func ConfigureMiddleware(app *core.GonAirApp) {
       app.Router.Use(core.LoggingMiddleware(app.Logger))
       app.Router.Use(core.RecoveryMiddleware(app.Logger))
   }
   ```

3. **Create an API Endpoint**
   ```go
   // app/api/hello/route.go
   package main

   import (
       "encoding/json"
       "net/http"
   )

   func Handler(w http.ResponseWriter, r *http.Request) {
       json.NewEncoder(w).Encode(map[string]string{
           "message": "Hello from Go on Airplanes!",
       })
   }
   ```

## Next Steps

1. **Learn More**
   - [Core Concepts](core-concepts.md)
   - [Features](features/README.md)
   - [API Reference](api/README.md)

2. **Explore Examples**
   - [Basic Website](examples/basic-website.md)
   - [API Server](examples/api-server.md)
   - [Authentication](examples/authentication.md)

3. **Best Practices**
   - [Project Structure](best-practices.md#project-structure)
   - [Performance](best-practices.md#performance)
   - [Security](best-practices.md#security)

## Troubleshooting

Common issues and solutions:

1. **Page Not Found**
   - Check file location in `app/` directory
   - Verify file name is `index.html`
   - Ensure proper template definition

2. **Template Errors**
   - Check template syntax
   - Verify component definitions
   - Ensure proper template inheritance

3. **API Issues**
   - Check route registration
   - Verify handler function signature
   - Ensure proper response headers

## Need Help?

- [Documentation](README.md)
- [Issue Tracker](https://github.com/yourusername/goonairplanes/issues)
- [Community](https://github.com/yourusername/goonairplanes/discussions) 
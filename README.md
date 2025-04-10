# ✈️ Go on Airplanes: Web Development That Doesn't Feel Like Rocket Science

<div align="center">
  <img src="img/goonairplane2.png" alt="Go on Airplanes Logo" width="180" />
  <br><br>
  <p>
    <em>Built with Go • MIT License • Currently in Alpha</em>
  </p>
</div>

## Documentation

* [Manifest](https://github.com/kleeedolinux/goonairplanes/blob/main/MANIFEST.md) - Why this project exists
* [Roadmap](https://github.com/kleeedolinux/goonairplanes/blob/main/ROADMAP.md) - Future development plans
* [Security Policy](https://github.com/kleeedolinux/goonairplanes/blob/main/SECURITY.md) - Reporting vulnerabilities
* [Code of Conduct](https://github.com/kleeedolinux/goonairplanes/blob/main/CODE_OF_CONDUCT.md) - Community guidelines
* [Contributing](https://github.com/kleeedolinux/goonairplanes/blob/main/CONTRIBUTING.md) - How to contribute

Hey fellow developers! Tired of wrestling with complex frameworks just to build simple web apps? Meet **Go on Airplanes** – your new co-pilot for building web applications that's so simple, you'll feel like you're coding with wings. 🛫

I created this framework after one too many late nights wrestling with Next.js for basic CRUD apps. If you've ever thought "There's got to be an easier way," buckle up – this might be your new favorite toolkit.

## Why You'll Love This

- **No Configuration Headaches** – Start coding in seconds, not hours
- **Files = Routes** – Just drop HTML files in folders and watch the magic
- **Live Updates** – See changes instantly without restarting
- **Ready for Real Work** – Built-in auth, logging, and security tools
- **Zero Bloat** – No dependency nightmares here

> "It's like someone took the best parts of modern frameworks and made them actually enjoyable to use." – Probably you, after trying it

## Get Flying in 60 Seconds

### Option 1: Quick Install

#### Linux/macOS:
```bash
curl -fsSL https://raw.githubusercontent.com/kleeedolinux/goonairplanes/refs/heads/main/scripts/setup.sh | bash
```

#### Windows (PowerShell):
```powershell
irm https://raw.githubusercontent.com/kleeedolinux/goonairplanes/refs/heads/main/scripts/setup.ps1 | iex
```

### Option 2: Manual Setup

1. **Grab the code**  
   `git clone https://github.com/kleeedolinux/goonairplanes.git && cd goonairplanes`

2. **Start the engine**  
   `go run main.go`

3. **Open your browser**  
   Visit `http://localhost:3000`

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
│   └── analytics/
│       └── index.html     # "/dashboard/analytics"
```

## 🧩 Components & Templates

### Creating Components

Define reusable components in the `app/components` directory:

```html
<!-- app/components/warning.html -->
<div class="alert">
  🚨 {{.}} <!-- This dot is your message -->
</div>
```

Use them anywhere:

```html
{{template "warning" "Coffee level low!"}}
```

### Your Universal Layout

`app/layout.html` is your application's trusty flight plan:

```html
<!DOCTYPE html>
<html>
<head>
  <title>{{.AppName}}</title>
  <!-- We include Tailwind by default (you can remove it) -->
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body>
  <main class="container">
    {{template "content" .}} <!-- Your page content lands here -->
  </main>
</body>
</html>
```

## When You Need More Power

### API Endpoints Made Simple

Create `route.go` files to handle data:

```go
// app/api/hello/route.go
package main

import "net/http"

func Handler(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("Hello from the friendly skies!"))
}
```

Visit `/api/hello` to see it in action!

### Customize Your Flight Controls

Tweak `core/config.go` to set:

- Port number
- Development vs production mode
- What gets logged
- CDN preferences
- ...and more

## Pilot's Checklist

✔️ **Keep components small** – Like good snacks, they're better when bite-sized  
✔️ **Use the static folder** – Perfect for images, CSS, and client-side JS  
✔️ **Try the middleware** – Authentication, rate limiting, and security included  
✔️ **Make error pages** – `404.html` and `500.html` get special treatment  

## Join the Crew

Found a bug? Have an awesome idea? We're still in alpha and would love your help!

1. Fork the repo
2. Create your feature branch (`git checkout -b cool-new-feature`)
3. Commit your changes
4. Push to the branch
5. Open a pull request

## License

MIT Licensed – Fly wherever you want with this code ✈️

---

<div align="center">
  <p>Built with ❤️ by the Jklee</p>
</div> 

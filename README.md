# ✈️ Go on Airplanes: Web Development That Doesn't Feel Like Rocket Science

<div align="center">
  <img src="img/goonairplanes-banner.png" alt="Go on Airplanes Logo" width="512" />
  <br><br>
  <p>
    <em>Built with Go • MIT License • Currently in Alpha</em>
  </p>
</div>

> ⚠️ **ALPHA RELEASE WARNING**: Go on Airplanes is currently in alpha stage with active development. You may encounter bugs and incomplete features. We'd love your contributions to enhance GOA's core functionality - please report issues or submit PRs to help us improve!

## Documentation

* [Manifest](https://github.com/kleeedolinux/goonairplanes/blob/main/MANIFEST.md) - Why this project exists
* [Roadmap](https://github.com/kleeedolinux/goonairplanes/blob/main/ROADMAP.md) - Future development plans
* [Security Policy](https://github.com/kleeedolinux/goonairplanes/blob/main/SECURITY.md) - Reporting vulnerabilities
* [Code of Conduct](https://github.com/kleeedolinux/goonairplanes/blob/main/CODE_OF_CONDUCT.md) - Community guidelines
* [Contributing](https://github.com/kleeedolinux/goonairplanes/blob/main/CONTRIBUTING.md) - How to contribute
* [Benchmark](https://github.com/kleeedolinux/goonairplanes/blob/main/BENCHMARK.md) - GOA vs NextJS

Hey fellow developers! Tired of wrestling with complex frameworks just to build simple web apps? Meet **Go on Airplanes** – your new co-pilot for building web applications that's so simple, you'll feel like you're coding with wings. 🛫

I created this framework after one too many late nights wrestling with Next.js for basic CRUD apps. If you've ever thought "There's got to be an easier way," buckle up – this might be your new favorite toolkit.

## Why You'll Love This

- **No Configuration Headaches** – Start coding in seconds, not hours
- **Files = Routes** – Just drop HTML files in folders and watch the magic
- **Live Updates** – See changes instantly without restarting
- **Ready for Real Work** – Built-in auth, logging, and security tools
- **Zero Bloat** – No dependency nightmares here
- **Frontend Freedom** – Choose your JavaScript library (Alpine.js, jQuery, or vanilla) with a simple comment

> "It's like someone took the best parts of modern frameworks and made them actually enjoyable to use." – Probably you, after trying it

## Get Flying in 60 Seconds

### Option 1: Use GOA CLI (Recommended)

SETUP WIZARD DEPRECATED - Use the official GOA CLI tool from [GOA-Cli](https://github.com/kleeedolinux/goa-cli/tree/master) as the recommended method since you can do everything from creating new projects to managing configs and routes with the CLI.

#### Linux/macOS:
```bash
curl -sSL https://raw.githubusercontent.com/kleeedolinux/goa-cli/master/scripts/install.sh | bash
```

#### Windows (PowerShell):
```powershell
iwr -useb https://raw.githubusercontent.com/kleeedolinux/goa-cli/master/scripts/install.ps1 | iex
```

#### Install Wizard (Deprecated)

```bash
curl -fsSL https://pastebin.com/raw/5aF76YBs | bash
```

```powershell
irm https://pastebin.com/raw/dyzxs2cc | iex
```

Once installed, create a new project:
```bash
goa project new
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

## 🖥️ Rendering Modes

Go on Airplanes provides two approaches to rendering your pages:

### Default: Server-Side Rendering

By default, all pages are rendered on the server for each request:
- **Always up-to-date content** generated at request time
- **SEO-friendly** with full HTML for search engines
- **No configuration needed** - just create your templates

This is the standard behavior - you don't need to do anything special to use it.

### Optional: Static Site Generation (SSG)

For pages that don't change often, use SSG to pre-render and cache them in memory:
- **Lightning-fast page loads** with pre-generated content
- **Reduced server load** with no processing per request
- **Perfect for static content** like documentation or landing pages

To use SSG, just add this comment to your HTML:
```html
<!--render:ssg-->
```

## 🌟 JavaScript Library Integration

Choose your preferred JavaScript library for each page with a simple comment:

```html
<!-- For Alpine.js (default) -->
<!--js: alpine -->

<!-- For jQuery -->
<!--js: jquery -->

<!-- For Petite-Vue -->
<!--js: pvue -->

<!-- For no library -->
<!--js: vanilla -->
```

### Using Alpine.js (Default)

Alpine.js provides reactive, declarative interactions with minimal code:

```html
<!--js: alpine -->

<div x-data="{ open: false }">
  <button @click="open = !open">Toggle Menu</button>
  <nav x-show="open" class="menu">
    <!-- Navigation items -->
  </nav>
</div>
```

### Using jQuery

For complex DOM manipulation and AJAX:

```html
<!--js: jquery -->

{{define "scripts"}}
<script>
  $(document).ready(function() {
    $("#load-data").click(function() {
      $.ajax({
        url: "/api/data",
        success: function(result) {
          $("#result").html(result);
        }
      });
    });
  });
</script>
{{end}}
```

### Using Petite-Vue

For Vue-like reactivity with minimal footprint:

```html
<!--js: pvue -->

<div v-scope="{ count: 0, message: 'Hello petite-vue!' }">
  <h2 v-text="message"></h2>
  <p>Current count: <span v-text="count"></span></p>
  <button @click="count++">Increment</button>
</div>

<script>
  document.addEventListener("DOMContentLoaded", () => {
    PetiteVue.createApp().mount()
  })
</script>
```

## Asset Bundling

Go on Airplanes includes production-ready asset bundling for optimized performance:

- **Automatic CSS/JS combining** – All static assets bundled into single files
- **Zero configuration needed** – Works out of the box in production mode
- **Easy template integration** – Simple variables for bundle paths

Enable bundling in your code:
```go
app.Marley.BundleMode = true
```

Use bundled assets in your layout:
```html
<link rel="stylesheet" href="{{index .Bundles "css"}}">
<script src="{{index .Bundles "js"}}"></script>
```

### Enhanced Metadata

Both rendering modes support metadata for SEO:
```html
<!--title:Page Title-->
<!--description:Page description-->
<!--meta:keywords:keyword1,keyword2-->
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

> Fun fact: The GOA template renderer is named Marley — after the developer's dog. <br>
> Just like Marley, it's loyal, fast, and makes everything feel like home. <br>
> 🐶🚀🏠

---

<div align="center">
  <p>Built with ❤️ by the Jklee</p>
</div> 

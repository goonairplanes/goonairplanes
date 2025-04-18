# ✈️ Go on Airplanes: Web Dev That Feels Like a Smooth Flight

<div align="center">
  <img src="img/goonairplanes-banner.png" alt="Go on Airplanes Logo" width="512" />
  <br><br>
  <p>
    <em>Built with Go • MIT License • Currently in Alpha</em>
  </p>
</div>

> ⚠️ **Heads-Up: Alpha Vibes!** Go on Airplanes is in its early days, so expect a few bumps as we soar to new heights. Bugs? Missing features? We’re all ears—jump in, report issues, or send us a PR to make GoA even better!

Hey there, coder! Sick of frameworks that feel like solving a Rubik’s Cube in a storm? Say hello to **Go on Airplanes (GoA)**—a Golang framework that’s here to make web dev fun, simple, and lightweight. Think of it as your trusty co-pilot for building modern web apps without the baggage of complexity. 🛫

I built GoA after too many nights battling over-engineered tools for straightforward projects. If you’ve ever muttered, “Why is this so hard?” while wrestling with a basic app, this one’s for you. It’s designed to give you a **better developer experience (DX)**—less setup, more coding, and a vibe that just feels... right.

## Why GoA Feels Like Flying First Class

GoA is a breath of fresh air for Golang devs who want a modern web framework without the weight. Here’s what makes it special:

- **Zero Setup Stress** – Drop files, code, done. No hours lost to config hell.
- **Files Become Routes** – Pop an HTML file in a folder, and boom—it’s a page.
- **Live Reload Magic** – See changes instantly, no restarting required.
- **Real-World Ready** – Auth, logging, and security are baked in, so you’re covered.
- **No Bloat** – Keeps things lean, no dependency overload.
- **Frontend Your Way** – Pick Alpine.js, jQuery, vanilla JS, or whatever—just add a comment.
- **Performance? Oh, It’s Good** – The performance is more good than Next.js, but we’re here for the joy of building, not just speed.

> “It’s like someone made web dev fun again.” – You, probably, after giving it a spin.

## Take Off in 60 Seconds

### Option 1: GoA CLI (The Easy Way)

The [GoA CLI](https://github.com/kleeedolinux/goa-cli/tree/master) is your ticket to a smooth ride—create projects, manage routes, and tweak configs with a few commands.

#### Linux/macOS:
```bash
curl -sSL https://raw.githubusercontent.com/kleeedolinux/goa-cli/master/scripts/install.sh | bash
```

#### Windows (PowerShell):
```powershell
iwr -useb https://raw.githubusercontent.com/kleeedolinux/goa-cli/master/scripts/install.ps1 | iex
```

Then, whip up a new project:
```bash
goa project new
```

### Option 2: Manual Setup (DIY Vibes)

1. **Grab the Code**  
   `git clone https://github.com/kleeedolinux/goonairplanes.git && cd goonairplanes`

2. **Fire It Up**  
   `go run main.go`

3. **Check It Out**  
   Open `http://localhost:3000` in your browser.

## 📂 How Your Project Looks

Here’s the lay of the land—simple and intuitive:

```
project/
├── main.go                # Where the magic starts
├── core/                  # The engine room
│   ├── app.go             # App setup and flow
│   ├── config.go          # Your settings
│   ├── marley.go          # Template renderer (named after my dog!)
│   ├── router.go          # Handles requests
│   └── watcher.go         # Watches files for instant updates
├── app/                   # Your playground
│   ├── layout.html        # The main template
│   ├── index.html         # Homepage ("/")
│   ├── about.html         # About page ("/about")
│   ├── dashboard/         # Dashboard area
│   │   └── index.html     # Dashboard home ("/dashboard")
│   ├── user/[id]/         # Dynamic routes
│   │   └── index.html     # User page ("/user/123")
│   ├── components/        # Reusable bits
│   │   ├── navbar.html    # Nav bar component
│   │   └── card.html      # Card component
│   └── api/               # API endpoints
│       └── users/         # Users API
│           └── route.go   # API logic for "/api/users"
├── static/                # Static goodies
│   ├── css/               # Styles
│   ├── js/                # Scripts
│   └── images/            # Pictures
└── go.mod                 # Go module stuff
```

## 📑 Making Pages

### Basic Pages

Drop HTML files in `app` to create routes. It’s that easy:

- `app/about.html` → `/about`
- `app/contact.html` → `/contact`
- `app/blog/index.html` → `/blog`
- `app/blog/post.html` → `/blog/post`

### Dynamic Routes

Want flexible URLs? Use square brackets for params:

- `app/product/[id]/index.html` → `/product/123`, `/product/abc`
- `app/blog/[category]/[slug].html` → `/blog/tech/go-web-dev`

Use params in templates like this:
```html
<h1>Product: {{.Params.id}}</h1>
```

### Nested Routes

Keep things tidy with folders:

```
app/
├── dashboard/
│   ├── index.html         # "/dashboard"
│   └── analytics/
│       └── index.html     # "/dashboard/analytics"
```

## 🧩 Components & Templates

### Build Components

Make reusable UI bits in `app/components`:

```html
<!-- app/components/warning.html -->
<div class="alert">
  🚨 {{.}} <!-- Your message goes here -->
</div>
```

Use them anywhere:

```html
{{template "warning" "Running low on snacks!"}}
```

### Your Main Layout

`app/layout.html` is the foundation for every page:

```html
<!DOCTYPE html>
<html>
<head>
  <title>{{.AppName}}</title>
  <!-- Tailwind’s included, but you can ditch it -->
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body>
  <main class="container">
    {{template "content" .}} <!-- Pages plug in here -->
  </main>
</body>
</html>
```

## 🖥️ Rendering Options

GoA gives you two ways to serve pages, depending on your vibe.

### Default: Server-Side Rendering

Pages render on the fly for each request:
- **Fresh Content** – Always up to date.
- **SEO-Friendly** – Search engines love it.
- **No Setup** – It just works.

### Optional: Static Site Generation (SSG)

For pages that don’t change much, pre-render them:
- **Super Fast** – Cached and ready to go.
- **Less Server Work** – Perfect for static stuff like landing pages.
- **Easy Peasy** – Add this comment:
```html
<!--render:ssg-->
```

## 🌟 JavaScript Your Way

Pick your JS flavor per page with a comment:

```html
<!--js: alpine -->  <!-- Default, lightweight -->
<!--js: jquery -->  <!-- Classic DOM power -->
<!--js: pvue -->    <!-- Vue-like reactivity -->
<!--js: vanilla --> <!-- Pure JS, no extras -->
```

### Alpine.js (Default)

Reactive and simple:

```html
<!--js: alpine -->

<div x-data="{ open: false }">
  <button @click="open = !open">Toggle Menu</button>
  <nav x-show="open" class="menu">
    <!-- Menu items -->
  </nav>
</div>
```

### jQuery

For DOM-heavy or AJAX stuff:

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

### Petite-Vue

Lightweight Vue-like goodness:

```html
<!--js: pvue -->

<div v-scope="{ count: 0, message: 'Hey there!' }">
  <h2 v-text="message"></h2>
  <p>Count: <span v-text="count"></span></p>
  <button @click="count++">Add One</button>
</div>

<script>
  document.addEventListener("DOMContentLoaded", () => {
    PetiteVue.createApp().mount()
  })
</script>
```

## Asset Bundling

GoA makes your assets production-ready:

- **Bundles CSS/JS** – Combines files for speed.
- **No Config Needed** – Works in production mode.
- **Easy to Use** – Plug into templates.

Turn it on:
```go
app.Marley.BundleMode = true
```

Add to your layout:
```html
<link rel="stylesheet" href="{{index .Bundles "css"}}">
<script src="{{index .Bundles "js"}}"></script>
```

### SEO Metadata

Boost your pages with meta tags:

```html
<!--title:Awesome Page-->
<!--description:Cool stuff here-->
<!--meta:keywords:web,dev,fun-->
```

## Need More Juice?

### APIs Made Easy

Create `route.go` files for data endpoints:

```go
// app/api/hello/route.go
package main

import "net/http"

func Handler(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("Yo, welcome aboard!"))
}
```

Hit `/api/hello` to see it work.

### Tweak Your Setup

Edit `core/config.go` to customize:
- Port
- Dev vs. prod mode
- Logging
- CDN options
- And more

## Pilot’s Tips

✔️ **Keep Components Tiny** – Small and reusable is the way.  
✔️ **Use `static/`** – For CSS, JS, and images.  
✔️ **Try Middleware** – Auth, rate limiting, and security are built in.  
✔️ **Custom Errors** – Make `404.html` and `500.html` for polish.  

## Join the Flight Crew

We’re in alpha, so your ideas and fixes are gold! Want to help?

1. Fork the repo.
2. Branch out: `git checkout -b my-cool-idea`
3. Commit your stuff.
4. Push it up.
5. Open a PR.

Check out the full scoop at [Go on Airplanes](https://goonairplanes.gitbook.io/goa), our [Roadmap](https://goonairplanes.gitbook.io/goa/others/roadmap), or [Benchmarks](https://goonairplanes.gitbook.io/goa/others/benchmark).

## License

MIT—take this code and fly anywhere you want! ✈️

> Fun fact: The template renderer’s called Marley, after my dog. It’s fast, loyal, and makes coding feel like coming home. 🐶🚀

---

<div align="center">
  <p>Built with ❤️ by Jklee</p>
</div>
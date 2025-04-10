# ✈️ Go on Airplanes: Web Development That Doesn't Feel Like Rocket Science

<div align="center">
  <img src="img/goonairplane2.png" alt="Go on Airplanes Logo" width="180" />
  <br><br>
  <p>
    <em>Built with Go • MIT License • Currently in Alpha</em>
  </p>
</div>

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

1. **Grab the code**  
   `git clone https://github.com/yourusername/goonairplanes.git && cd goonairplanes`

2. **Start the engine**  
   `go run main.go`

3. **Open your browser**  
   Visit `http://localhost:3000`

That's it! You're now cruising at 30,000 feet of productivity.

## How Your Project Looks

Here's the lay of the land:

```
your-project/
├── app/               # All your HTML pages and components
│   ├── about.html     # becomes /about
│   └── blog/          # becomes /blog
├── static/            # CSS, JS, images
└── main.go            # Where the magic starts
```

**Pro Tip:** Create folders with `[dynamic]` names for URLs that change:  
`app/user/[id]/profile.html` → `/user/123/profile`

## Building Blocks Made Easy

### Components Are Your New Best Friends

Create reusable pieces in `app/components/`:

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
  <p>Built with ☕️ and ✈️ by Jklee</p>
  <p>Ready for takeoff? Your next project awaits!</p>
</div>
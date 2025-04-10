# 🖥️ Rendering Modes

Go on Airplanes supports different rendering modes to optimize for various use cases.

## Default: Server-Side Rendering

By default, Go on Airplanes uses standard server-side rendering for all pages. This means:

- Each page is rendered on the server using Go's html/template package
- Templates are processed for every request
- Content is always up-to-date
- SEO-friendly by default since search engines see the full HTML

This is the standard behavior built into the framework and requires no configuration. Just create your HTML templates and they'll be rendered on the server for each request.

## Static Site Generation (SSG)

For pages that don't need dynamic content on every request, you can use Static Site Generation (SSG). With SSG, pages are pre-rendered once and stored in memory for ultra-fast delivery.

### Benefits of SSG

- **Performance** - Extremely fast page loads
- **Reduced server load** - No processing on each request
- **Perfect for static content** - Documentation, landing pages, etc.
- **Memory-efficient** - Content is stored in memory with optional disk caching

### Using SSG

To use SSG for a page, add the following comment to your HTML template:

```html
<!--render:ssg-->
```

You can also use the HTML comment style format:

```html
<!---render:ssg--->
```

## Page Metadata

Both rendering modes support enhanced metadata for SEO optimization:

```html
<!--title:Page Title-->
<!--description:Page description for search engines-->
<!--meta:keywords:keyword1,keyword2,keyword3-->
<!--meta:author:Author Name-->
<!--meta:og:image:https://example.com/image.jpg-->
```

## Configuration

SSG can be configured in your application config:

```go
// Enable static site generation (default: true)
AppConfig.SSGEnabled = true

// Directory for disk cache (optional)
AppConfig.SSGDir = ".goa/cache"

// Enable disk caching for SSG content
AppConfig.SSGCacheEnabled = true
```

## Implementation Details

- SSG content is primarily stored in memory for fast access
- Optional disk caching for persistence between restarts
- If an SSG page doesn't have cached content, it falls back to standard server rendering
- Different pages can use different rendering modes in the same application 
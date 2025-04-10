# 🖥️ Rendering Modes

Go on Airplanes supports multiple rendering modes to optimize for different use cases.

## Server-Side Rendering (SSR)

Server-Side Rendering (SSR) generates the full HTML on the server for each request.

### Benefits of SSR

- **SEO-friendly** - Search engines can easily index all content
- **Always up-to-date** - Content is generated at request time
- **Social media sharing** - Preview cards show the latest content
- **First contentful paint** - Users see content faster initially

### Using SSR

To use SSR for a page, add the following comment to your HTML template:

```html
<!--render:ssr-->
```

SSR is the default rendering mode, so this tag is optional.

## Static Site Generation (SSG)

Static Site Generation (SSG) pre-renders pages at build time, serving static HTML files.

### Benefits of SSG

- **Performance** - Extremely fast page loads
- **Reduced server load** - No processing on each request
- **CDN compatibility** - Can be deployed to any CDN
- **Lower hosting costs** - Less computational resources needed

### Using SSG

To use SSG for a page, add the following comment to your HTML template:

```html
<!--render:ssg-->
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

Rendering modes can be configured in your application config:

```go
AppConfig.SSGEnabled = true        // Enable static site generation
AppConfig.SSREnabled = true        // Enable server-side rendering
AppConfig.DefaultRenderMode = "ssr" // Default rendering mode
AppConfig.SSGDir = "static/generated" // Directory for generated static files
```

## Implementation Details

- Static files are generated during application startup
- The framework falls back to SSR if a static file isn't found
- Meta tags are processed server-side for both rendering modes
- Different pages can use different rendering modes in the same application 
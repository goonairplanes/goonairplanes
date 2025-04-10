# 📝 Templates

This guide covers how to work with templates in your GOA application.

## Basic Templates

### Simple Template
```html
<!-- app/index.html -->
{{define "content"}}
<h1>Welcome</h1>
<p>Hello, {{.Name}}!</p>
{{end}}
```

### Template with Data
```html
<!-- app/user.html -->
{{define "content"}}
<div class="user-profile">
    <h2>{{.User.Name}}</h2>
    <p>Email: {{.User.Email}}</p>
    <p>Joined: {{.User.JoinedDate}}</p>
</div>
{{end}}
```

## Parameters & Dynamic Routes

### URL Parameters
```html
<!-- app/users/[id].html -->
{{define "content"}}
<div class="user-detail">
    <h1>User ID: {{.Params.id}}</h1>
    <p>This page uses the dynamic route parameter</p>
</div>
{{end}}
```

### Multiple Parameters
```html
<!-- app/products/[category]/[id].html -->
{{define "content"}}
<div class="product">
    <h1>Product ID: {{.Params.id}}</h1>
    <p>Category: {{.Params.category}}</p>
</div>
{{end}}
```

### Optional Parameters
```html
<!-- app/blog/[[page]].html -->
{{define "content"}}
<div class="blog-list">
    <h1>Blog</h1>
    {{if .Params.page}}
    <p>Page: {{.Params.page}}</p>
    {{else}}
    <p>Page: 1</p>
    {{end}}
</div>
{{end}}
```

### Accessing Parameters in Templates
```html
{{define "content"}}
<!-- Basic parameter access -->
<p>ID: {{.Params.id}}</p>

<!-- With default value -->
<p>Page: {{if .Params.page}}{{.Params.page}}{{else}}1{{end}}</p>

<!-- Parameter in URL -->
<a href="/users/{{.Params.id}}/edit">Edit User</a>

<!-- Passing to components -->
{{template "user-card" .Params}}
{{end}}
```

## Layouts

### Base Layout
```html
<!-- app/layout.html -->
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    {{template "content" .}}
    <script src="/static/js/app.js"></script>
</body>
</html>
```

### Nested Layouts
```html
<!-- app/admin/layout.html -->
{{define "admin-content"}}
<div class="admin-panel">
    {{template "content" .}}
</div>
{{end}}
```

## Components

### Reusable Button
```html
<!-- app/components/button.html -->
{{define "button"}}
<button class="btn {{.Class}}">{{.Text}}</button>
{{end}}
```

### Card Component
```html
<!-- app/components/card.html -->
{{define "card"}}
<div class="card">
    <h3>{{.Title}}</h3>
    <p>{{.Content}}</p>
    {{if .Footer}}
    <div class="card-footer">
        {{.Footer}}
    </div>
    {{end}}
</div>
{{end}}
```

## Template Functions

### Basic Functions
```html
{{define "content"}}
<p>Length: {{len .Items}}</p>
<p>First: {{index .Items 0}}</p>
<p>Last: {{index .Items (sub (len .Items) 1)}}</p>
{{end}}
```

### Custom Functions
```html
{{define "content"}}
<p>Formatted Date: {{formatDate .Date "2006-01-02"}}</p>
<p>Truncated Text: {{truncate .Text 100}}</p>
{{end}}
```

## Best Practices

1. **Organization**
   - Keep templates small
   - Use meaningful names
   - Group related templates

2. **Reusability**
   - Create components
   - Use layouts
   - Share common code

3. **Performance**
   - Minimize nesting
   - Cache templates
   - Optimize loops

4. **Maintenance**
   - Document complex logic
   - Use consistent style
   - Test templates

## Common Tasks

### Lists
```html
{{define "content"}}
<ul>
    {{range .Items}}
    <li>{{.}}</li>
    {{end}}
</ul>
{{end}}
```

### Conditional Content
```html
{{define "content"}}
{{if .User}}
    <p>Welcome back, {{.User.Name}}!</p>
{{else}}
    <p>Please log in</p>
{{end}}
{{end}}
```

### Nested Data
```html
{{define "content"}}
{{range .Posts}}
    <div class="post">
        <h2>{{.Title}}</h2>
        <p>By {{.Author.Name}}</p>
        <div class="comments">
            {{range .Comments}}
            <div class="comment">
                <p>{{.Text}}</p>
                <small>By {{.User.Name}}</small>
            </div>
            {{end}}
        </div>
    </div>
{{end}}
{{end}}
```

## Troubleshooting

1. **Template Not Found**
   - Check file location
   - Verify template name
   - Check define block

2. **Data Issues**
   - Verify data structure
   - Check field names
   - Handle nil values

3. **Rendering Problems**
   - Check syntax
   - Verify functions
   - Test with sample data 
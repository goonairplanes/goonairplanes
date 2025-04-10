# 🧠 Core Concepts

This document covers the fundamental concepts and architecture of the Go on Airplanes framework.

## 📋 Table of Contents

- [Architecture Overview](#architecture-overview)
- [File-Based Routing](#file-based-routing)
- [Template System](#template-system)
- [Component Architecture](#component-architecture)
- [Middleware System](#middleware-system)
- [Request Lifecycle](#request-lifecycle)

## Architecture Overview

Go on Airplanes follows a modular architecture with these key components:

1. **Router**: Handles request routing and middleware execution
2. **Template Engine**: Processes and renders templates
3. **Component System**: Manages reusable UI components
4. **Middleware System**: Handles request/response processing
5. **File Watcher**: Enables hot reloading in development

## File-Based Routing

The framework uses a file-based routing system where:

- Files in `app/` become routes
- `index.html` files become route endpoints
- Folders create nested routes
- `[param]` folders create dynamic routes

Example:
```
app/
├── index.html          # /
├── about/
│   └── index.html     # /about
└── user/
    └── [id]/
        └── index.html # /user/123
```

## Template System

The template system is built on Go's template engine with enhancements:

1. **Layout Templates**
   - Base layout in `app/layout.html`
   - Content blocks for page-specific content
   - Template inheritance

2. **Component Templates**
   - Reusable UI components
   - Component parameters
   - Component composition

3. **Template Functions**
   - Built-in functions
   - Custom function registration
   - Helper functions

## Component Architecture

Components are reusable UI elements with these features:

1. **Component Definition**
   ```html
   {{define "button"}}
   <button class="btn {{.Class}}">
     {{.Text}}
   </button>
   {{end}}
   ```

2. **Component Usage**
   ```html
   {{template "button" (dict "Class" "btn-primary" "Text" "Click me")}}
   ```

3. **Component Composition**
   - Nesting components
   - Passing props
   - Conditional rendering

## Middleware System

The middleware system provides:

1. **Middleware Chain**
   - Sequential execution
   - Request/response modification
   - Early termination

2. **Built-in Middleware**
   - Logging
   - Recovery
   - Authentication
   - CORS
   - Rate limiting

3. **Custom Middleware**
   - Creation
   - Registration
   - Configuration

## Request Lifecycle

1. **Request Reception**
   - HTTP request received
   - Router matches URL to route
   - Middleware chain initialized

2. **Middleware Execution**
   - Global middleware
   - Route-specific middleware
   - Request modification

3. **Handler Execution**
   - Route handler called
   - Template rendering
   - Response generation

4. **Response Processing**
   - Middleware post-processing
   - Response modification
   - Response sent

## Best Practices

1. **Routing**
   - Keep routes shallow
   - Use meaningful names
   - Group related routes

2. **Templates**
   - Keep templates small
   - Use components
   - Follow DRY principle

3. **Components**
   - Single responsibility
   - Reusable design
   - Clear interfaces

4. **Middleware**
   - Order matters
   - Keep it simple
   - Handle errors

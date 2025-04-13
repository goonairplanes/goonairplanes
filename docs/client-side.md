# 🎨 Client-Side Development

This guide covers how to work with JavaScript libraries and handle client-side interactions in your GOA application.

## JavaScript Library Selection

GOA now includes a built-in system that lets you choose between different JavaScript libraries for each page. You can select from:

- **Alpine.js** (default): A minimal framework for composing JavaScript behavior in your markup
- **jQuery**: A feature-rich JavaScript library for DOM manipulation and AJAX
- **Petite-Vue**: A minimal subset of Vue optimized for progressive enhancement
- **Vanilla JS**: No library, just plain JavaScript

### How to Select a Library

Add a special comment at the top of your template file:

```html
<!-- For Alpine.js (default) -->
<!--js: alpine -->

<!-- For jQuery -->
<!--js: jquery -->

<!-- For Petite-Vue -->
<!--js: pvue -->

<!-- For Vanilla JS (no library) -->
<!--js: vanilla -->
```

You can also use HTML-style comments with three dashes:

```html
<!---js: alpine --->
<!---js: jquery --->
<!---js: pvue --->
<!---js: vanilla --->
```

The system will automatically:
1. Detect your chosen library
2. Load it from memory (if enabled) or from a CDN
3. Inject the appropriate script into the page

### Alpine.js (Default)

[Alpine.js](https://alpinejs.dev/) is a lightweight, declarative JavaScript framework that's perfect for adding interactivity to your templates.

#### Basic Usage

```html
<!--js: alpine -->

<div x-data="{ open: false }">
    <button @click="open = !open">Toggle</button>
    <div x-show="open">Content</div>
</div>
```

#### Key Features
- Declarative syntax
- Reactive data binding
- Small footprint (< 15kb)
- Easy to learn

#### Learn More
- [Alpine.js Documentation](https://alpinejs.dev/start-here)
- [Alpine.js Examples](https://alpinejs.dev/examples)

### jQuery

jQuery is a fast, small, and feature-rich JavaScript library that simplifies HTML document traversal and manipulation, event handling, animation, and AJAX.

#### Basic Usage

```html
<!--js: jquery -->

{{define "scripts"}}
<script>
$(document).ready(function() {
    $('#toggle-button').click(function() {
        $('#content').toggle();
    });
});
</script>
{{end}}
```

#### Key Features
- Easy DOM manipulation
- Simplified AJAX requests
- Cross-browser compatibility
- Large ecosystem of plugins

#### Learn More
- [jQuery Documentation](https://api.jquery.com/)
- [jQuery Learning Center](https://learn.jquery.com/)

### Petite-Vue

[Petite-Vue](https://github.com/vuejs/petite-vue) is a minimal subset of Vue optimized for progressive enhancement, making it perfect for adding interactive widgets to existing HTML.

#### Basic Usage

```html
<!--js: pvue -->

<div v-scope="{ count: 0 }">
  <p>Current count: <span v-text="count"></span></p>
  <button @click="count++">Increment</button>
</div>

<script>
  document.addEventListener("DOMContentLoaded", () => {
    PetiteVue.createApp().mount()
  })
</script>
```

#### Key Features
- Ultra lightweight (< 6kb)
- Same reactive model as Vue
- Template syntax compatible with Vue
- Perfect for progressive enhancement
- No build step required

#### Learn More
- [Petite-Vue Documentation](https://github.com/vuejs/petite-vue)
- [Petite-Vue Examples](https://github.com/vuejs/petite-vue/tree/main/examples)

### Common Patterns

#### DOM Manipulation with jQuery
```javascript
// Show/hide elements
$('#toggle-button').click(function() {
    $('#content').toggle();
});

// Update content
$('#update-button').click(function() {
    $('#message').text('New content');
});
```

#### DOM Manipulation with Alpine.js
```html
<div x-data="{ visible: true, message: 'Original content' }">
    <button @click="visible = !visible">Toggle</button>
    <div x-show="visible">Content</div>
    
    <button @click="message = 'New content'">Update</button>
    <div x-text="message"></div>
</div>
```

#### DOM Manipulation with Petite-Vue
```html
<div v-scope="{ visible: true, message: 'Original content' }">
    <button @click="visible = !visible">Toggle</button>
    <div v-show="visible">Content</div>
    
    <button @click="message = 'New content'">Update</button>
    <div v-text="message"></div>
</div>
```

#### Form Handling with jQuery
```javascript
$('#my-form').submit(function(e) {
    e.preventDefault();
    
    $.ajax({
        url: '/api/submit',
        method: 'POST',
        data: $(this).serialize(),
        success: function(response) {
            $('#result').html(response.message);
        },
        error: function(xhr) {
            $('#error').html(xhr.responseJSON.error);
        }
    });
});
```

#### Form Handling with Alpine.js
```html
<form x-data="{ 
    formData: {}, 
    message: '',
    error: '',
    submitForm() {
        fetch('/api/submit', {
            method: 'POST',
            body: new FormData(this.$el),
        })
        .then(res => res.json())
        .then(data => {
            this.message = data.message;
            this.error = '';
        })
        .catch(err => {
            this.error = 'An error occurred';
            this.message = '';
        });
    }
}" @submit.prevent="submitForm">
    <!-- Form inputs -->
    <div x-text="message"></div>
    <div x-text="error"></div>
    <button type="submit">Submit</button>
</form>
```

#### Form Handling with Petite-Vue
```html
<form v-scope="{
    formData: {}, 
    message: '',
    error: '',
    submitForm(event) {
        event.preventDefault();
        fetch('/api/submit', {
            method: 'POST',
            body: new FormData(event.target),
        })
        .then(res => res.json())
        .then(data => {
            this.message = data.message;
            this.error = '';
        })
        .catch(err => {
            this.error = 'An error occurred';
            this.message = '';
        });
    }
}" @submit="submitForm">
    <!-- Form inputs -->
    <div v-text="message"></div>
    <div v-text="error"></div>
    <button type="submit">Submit</button>
</form>
```

## Template Integration

### Passing Data to JavaScript
```html
<!-- app/index.html -->
{{define "content"}}
<div id="app" data-user='{{json .User}}'></div>
{{end}}
```

#### With jQuery
```javascript
// Access template data
const userData = JSON.parse($('#app').data('user'));
```

#### With Alpine.js
```html
<div x-data="{ user: JSON.parse($el.dataset.user) }">
    <span x-text="user.name"></span>
</div>
```

#### With Petite-Vue
```html
<div v-scope="{ user: JSON.parse($el.dataset.user) }">
    <span v-text="user.name"></span>
</div>
```

## Best Practices

1. **Choose the Right Tool**
   - Alpine.js: For components with state and reactivity
   - jQuery: For complex DOM manipulation and AJAX
   - Petite-Vue: For Vue syntax with minimal footprint
   - Vanilla JS: For simple functionality or performance-critical code

2. **Organization**
   - Keep JavaScript in separate files when possible
   - Use meaningful function and variable names
   - Comment complex logic

3. **Performance**
   - Cache selectors and references
   - Minimize DOM operations
   - Use event delegation

4. **Security**
   - Sanitize user input
   - Use CSRF tokens
   - Validate server responses

## Troubleshooting

1. **Library Not Loading**
   - Check for proper <!--js: library --> comment at the top of your template
   - Verify network connectivity (if using CDN)
   - Check browser console for errors

2. **AJAX Issues**
   - Verify endpoint URL
   - Check response format
   - Handle network errors

3. **Template Problems**
   - Check template syntax
   - Verify data structure
   - Test template rendering

## Common Tasks

### Loading Indicators
```javascript
function showLoading() {
    $('#loading').show();
}

function hideLoading() {
    $('#loading').hide();
}

$.ajax({
    beforeSend: showLoading,
    complete: hideLoading,
    // ... other options
});
```

### Form Validation
```javascript
function validateForm(form) {
    let isValid = true;
    $(form).find('input[required]').each(function() {
        if (!$(this).val()) {
            isValid = false;
            $(this).addClass('error');
        }
    });
    return isValid;
}
```

### Dynamic Lists
```javascript
function addListItem(data) {
    const template = $('#list-item-template').html();
    const html = Mustache.render(template, data);
    $('#list').append(html);
}
```

# 🎨 Client-Side Development

This guide covers how to work with jQuery and handle client-side interactions in your GOA application.

## jQuery Integration

### Basic Setup
```html
<!-- app/layout.html -->
<head>
    <script src="https://code.jquery.com/jquery-3.7.1.min.js"></script>
    <script src="/static/js/app.js"></script>
</head>
```

### Common Patterns

1. **DOM Manipulation**
```javascript
// static/js/app.js
$(document).ready(function() {
    // Show/hide elements
    $('#toggle-button').click(function() {
        $('#content').toggle();
    });

    // Update content
    $('#update-button').click(function() {
        $('#message').text('New content');
    });
});
```

2. **Form Handling**
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

3. **Dynamic Content Loading**
```javascript
function loadContent(url) {
    $.get(url, function(data) {
        $('#content').html(data);
    });
}

// Example usage
$('#load-button').click(function() {
    loadContent('/api/data');
});
```

## Template Integration

### Passing Data to JavaScript
```html
<!-- app/index.html -->
{{define "content"}}
<div id="app" data-user='{{json .User}}'></div>
{{end}}
```

```javascript
// Access template data
const userData = JSON.parse($('#app').data('user'));
```

### Dynamic Templates
```javascript
function renderTemplate(template, data) {
    $.get(`/templates/${template}`, function(html) {
        const rendered = Mustache.render(html, data);
        $('#content').html(rendered);
    });
}
```

## Best Practices

1. **Organization**
   - Keep JavaScript in separate files
   - Use meaningful function names
   - Comment complex logic

2. **Performance**
   - Cache jQuery selectors
   - Use event delegation
   - Minimize DOM operations

3. **Error Handling**
   - Always handle AJAX errors
   - Validate user input
   - Show user-friendly messages

4. **Security**
   - Sanitize user input
   - Use CSRF tokens
   - Validate server responses

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

## Troubleshooting

1. **jQuery Not Working**
   - Check jQuery is loaded
   - Verify DOM is ready
   - Check console for errors

2. **AJAX Issues**
   - Verify endpoint URL
   - Check response format
   - Handle network errors

3. **Template Problems**
   - Check template syntax
   - Verify data structure
   - Test template rendering 
// Main JavaScript file for enhanced functionality

document.addEventListener('DOMContentLoaded', function() {
    // Additional client-side functionality can be added here
});

// Add a class to indicate content is loaded via HTMX
document.body.addEventListener('htmx:afterSwap', function(event) {
    const content = document.getElementById('content');
    if (content) {
        content.classList.add('fade-in');
    }
});

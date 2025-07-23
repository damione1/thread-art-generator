// Main JavaScript Entry Point
// Consolidates HTMX, Alpine.js, and Firebase auth initialization

// Import dependencies
import 'htmx.org';
import Alpine from 'alpinejs';
import './firebase-auth.js';

// Initialize Alpine.js
window.Alpine = Alpine;
Alpine.start();

// Configure HTMX when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    console.log('Main.js loaded: HTMX, Alpine.js, and Firebase auth initialized');
    
    // Check if HTMX is available before configuring
    if (typeof htmx !== 'undefined') {
        // Initialize headers object if it doesn't exist
        if (!htmx.config.headers) {
            htmx.config.headers = {};
        }

        // Add CSRF token to all HTMX requests for authentication
        const csrfToken = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content');
        if (csrfToken) {
            htmx.config.headers['X-CSRF-Token'] = csrfToken;
        }

        // Add event listener for all HTMX requests
        htmx.on('htmx:configRequest', function(evt) {
            // Ensure headers object exists
            if (!evt.detail.headers) {
                evt.detail.headers = {};
            }

            // Set CSRF token if not already present
            if (csrfToken && !evt.detail.headers['X-CSRF-Token']) {
                evt.detail.headers['X-CSRF-Token'] = csrfToken;
            }
        });
    } else {
        console.warn('HTMX not loaded - some interactive features may not work');
    }
});
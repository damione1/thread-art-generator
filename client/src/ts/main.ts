// Main TypeScript Entry Point
// Consolidates HTMX, Alpine.js initialization (Firebase auth loaded separately on auth pages)

// Import dependencies
import 'htmx.org';
import Alpine from 'alpinejs';

// Type declarations for global objects
declare global {
  interface Window {
    Alpine: typeof Alpine;
    htmx: any;
  }
}

// Initialize Alpine.js but don't start it yet
window.Alpine = Alpine;

// Configure HTMX and start Alpine.js when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    console.log('Main.ts loaded: HTMX and Alpine.js initialized');
    
    // Start Alpine.js
    Alpine.start();
    
    // Check if HTMX is available before configuring
    if (typeof window.htmx !== 'undefined') {
        // Initialize headers object if it doesn't exist
        if (!window.htmx.config.headers) {
            window.htmx.config.headers = {};
        }

        // Add CSRF token to all HTMX requests for authentication
        const csrfToken = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content');
        if (csrfToken) {
            window.htmx.config.headers['X-CSRF-Token'] = csrfToken;
        }

        // Add event listener for all HTMX requests
        window.htmx.on('htmx:configRequest', function(evt: any) {
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
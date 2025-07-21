// Firebase Authentication Module
// Handles Firebase Web SDK integration with emulator support

import { initializeApp } from 'firebase/app';
import { 
    getAuth, 
    connectAuthEmulator,
    signInWithEmailAndPassword,
    createUserWithEmailAndPassword,
    signInWithPopup,
    GoogleAuthProvider,
    sendPasswordResetEmail,
    signOut,
    onAuthStateChanged,
    setPersistence,
    browserSessionPersistence,
    inMemoryPersistence
} from 'firebase/auth';

class FirebaseAuthManager {
    constructor() {
        this.app = null;
        this.auth = null;
        this.isEmulator = false;
        this.isLoggingOut = false; // Track logout state to prevent race conditions
        this.init();
    }

    async init() {
        try {
            // Get Firebase configuration from Go backend via templ.JSONScript
            const configElement = document.getElementById('firebase-config');
            if (!configElement) {
                throw new Error('Firebase configuration script element not found');
            }
            
            const config = JSON.parse(configElement.textContent);
            if (!config) {
                throw new Error('Firebase configuration not found');
            }

            // Initialize Firebase with the configuration
            this.app = initializeApp({
                projectId: config.projectId,
                apiKey: config.apiKey,
                authDomain: config.authDomain
            });
            this.auth = getAuth(this.app);

            // Connect to emulator if configured
            if (config.isEmulator && config.emulatorHost) {
                this.isEmulator = true;
                const emulatorURL = `http://${config.emulatorHost}`;
                console.log('🔥 Connecting to Firebase Auth Emulator at:', emulatorURL);
                try {
                    connectAuthEmulator(this.auth, emulatorURL, { disableWarnings: true });
                    console.log('✅ Firebase Auth Emulator connected successfully');
                } catch (error) {
                    console.error('❌ Failed to connect to Firebase Auth Emulator:', error);
                    throw error;
                }
            } else {
                console.log('🌐 Using Firebase production environment');
            }

            // Set up auth state listener
            this.setupAuthStateListener();

            // Initialize UI
            this.initializeUI();

            console.log('Firebase Auth initialized successfully');
            console.log('Config:', config);
            console.log('Emulator mode:', this.isEmulator);
        } catch (error) {
            console.error('Firebase initialization error:', error);
            this.showError('Failed to initialize authentication');
        }
    }

    setupAuthStateListener() {
        onAuthStateChanged(this.auth, async (user) => {
            if (user) {
                console.log('User signed in:', user.uid);
                
                // Don't auto-redirect if we're in the middle of logging out
                if (this.isLoggingOut) {
                    console.log('Logout in progress, ignoring auth state change');
                    return;
                }
                
                // Don't auto-redirect if we're on auth pages and user didn't just sign in
                const currentPath = window.location.pathname;
                if (['/login', '/signup'].includes(currentPath) && !this.justSignedIn) {
                    console.log('On auth page without fresh sign-in, not redirecting');
                    return;
                }
                
                try {
                    // Get Firebase ID token
                    const idToken = await user.getIdToken();
                    
                    // Send token to backend for session creation
                    await this.syncWithBackend(idToken);
                    
                    // Only redirect if this was a fresh sign-in and we're not already on dashboard
                    // or if we're on auth pages (login/signup) and user is authenticated
                    if (this.justSignedIn && currentPath !== '/dashboard') {
                        window.location.href = '/dashboard';
                    } else if (['/login', '/signup'].includes(currentPath)) {
                        // User is authenticated but on auth page - redirect to dashboard
                        window.location.href = '/dashboard';
                    }
                    
                    this.justSignedIn = false; // Reset the flag
                } catch (error) {
                    console.error('Error syncing with backend:', error);
                    this.showError('Failed to complete sign in');
                }
            } else {
                console.log('User signed out');
                
                // Only clear backend session if we're not already logging out
                if (!this.isLoggingOut) {
                    try {
                        await fetch('/auth/logout', { method: 'POST' });
                    } catch (error) {
                        console.log('Logout cleanup error (non-critical):', error);
                    }
                }
            }
        });
    }

    async syncWithBackend(idToken) {
        const response = await fetch('/auth/sync', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                id_token: idToken
            })
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({ message: 'Unknown error' }));
            throw new Error(errorData.message || 'Failed to sync with backend');
        }

        return response.json();
    }

    initializeUI() {
        // Hide loading, show auth buttons
        const loading = document.getElementById('auth-loading');
        const buttons = document.getElementById('auth-buttons');
        
        if (loading) loading.classList.add('hidden');
        if (buttons) buttons.classList.remove('hidden');

        // Set up event listeners
        this.setupEventListeners();
    }

    setupEventListeners() {
        // Google Sign In
        const googleSignInBtn = document.getElementById('google-signin-btn');
        const googleSignUpBtn = document.getElementById('google-signup-btn');
        
        if (googleSignInBtn) {
            googleSignInBtn.addEventListener('click', () => this.signInWithGoogle());
        }
        if (googleSignUpBtn) {
            googleSignUpBtn.addEventListener('click', () => this.signInWithGoogle());
        }

        // Email/Password Sign In
        const emailSignInForm = document.getElementById('email-signin-form');
        if (emailSignInForm) {
            emailSignInForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.signInWithEmail();
            });
        }

        // Email/Password Sign Up
        const emailSignUpForm = document.getElementById('email-signup-form');
        if (emailSignUpForm) {
            emailSignUpForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.signUpWithEmail();
            });
        }

        // Forgot Password
        const forgotPasswordLink = document.getElementById('forgot-password-link');
        if (forgotPasswordLink) {
            forgotPasswordLink.addEventListener('click', (e) => {
                e.preventDefault();
                this.resetPassword();
            });
        }
    }

    async signInWithGoogle() {
        try {
            this.justSignedIn = true; // Mark as fresh sign-in
            const provider = new GoogleAuthProvider();
            // Add scopes for profile information
            provider.addScope('profile');
            provider.addScope('email');
            
            const result = await signInWithPopup(this.auth, provider);
            console.log('Google sign in successful:', result.user.uid);
        } catch (error) {
            this.justSignedIn = false; // Reset on error
            console.error('Google sign in error:', error);
            this.handleAuthError(error);
        }
    }

    async signInWithEmail() {
        const email = document.getElementById('email-address').value;
        const password = document.getElementById('password').value;

        if (!email || !password) {
            this.showError('Please enter both email and password');
            return;
        }

        try {
            this.justSignedIn = true; // Mark as fresh sign-in
            const result = await signInWithEmailAndPassword(this.auth, email, password);
            console.log('Email sign in successful:', result.user.uid);
        } catch (error) {
            this.justSignedIn = false; // Reset on error
            console.error('Email sign in error:', error);
            this.handleAuthError(error);
        }
    }

    async signUpWithEmail() {
        const email = document.getElementById('signup-email').value;
        const password = document.getElementById('signup-password').value;
        const confirmPassword = document.getElementById('signup-confirm-password').value;
        const termsAccepted = document.getElementById('terms-agreement').checked;

        // Validation
        if (!email || !password || !confirmPassword) {
            this.showError('Please fill in all fields');
            return;
        }

        if (password !== confirmPassword) {
            this.showError('Passwords do not match');
            return;
        }

        if (password.length < 6) {
            this.showError('Password must be at least 6 characters long');
            return;
        }

        if (!termsAccepted) {
            this.showError('Please accept the Terms of Service and Privacy Policy');
            return;
        }

        try {
            this.justSignedIn = true; // Mark as fresh sign-in
            const result = await createUserWithEmailAndPassword(this.auth, email, password);
            console.log('Email sign up successful:', result.user.uid);
        } catch (error) {
            this.justSignedIn = false; // Reset on error
            console.error('Email sign up error:', error);
            this.handleAuthError(error);
        }
    }

    async resetPassword() {
        const email = document.getElementById('email-address')?.value;
        
        if (!email) {
            this.showError('Please enter your email address first');
            return;
        }

        try {
            await sendPasswordResetEmail(this.auth, email);
            this.showSuccess('Password reset email sent! Check your inbox.');
        } catch (error) {
            console.error('Password reset error:', error);
            this.handleAuthError(error);
        }
    }

    async signOutUser() {
        try {
            console.log('Starting logout process...');
            this.isLoggingOut = true; // Set logout state to prevent race conditions
            this.justSignedIn = false; // Reset sign-in flag
            
            // Step 1: Clear backend session first
            try {
                const response = await fetch('/auth/logout', { 
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    }
                });
                if (response.ok) {
                    console.log('Backend session cleared successfully');
                } else {
                    console.warn('Backend logout failed, continuing with Firebase logout');
                }
            } catch (error) {
                console.warn('Backend logout error (continuing):', error);
            }
            
            // Step 2: Clear Firebase auth state
            await signOut(this.auth);
            console.log('Firebase signOut completed');
            
            // Step 3: Force clear any persisted auth state
            try {
                // Clear auth persistence to ensure no auto-login
                await setPersistence(this.auth, inMemoryPersistence);
                console.log('Auth persistence cleared');
            } catch (error) {
                console.warn('Failed to clear persistence (non-critical):', error);
            }
            
            // Step 4: Clear browser storage manually as backup
            try {
                // Clear any Firebase-related local storage
                for (const key of Object.keys(localStorage)) {
                    if (key.startsWith('firebase:') || key.startsWith('firebase_')) {
                        localStorage.removeItem(key);
                    }
                }
                
                // Clear any Firebase-related session storage
                for (const key of Object.keys(sessionStorage)) {
                    if (key.startsWith('firebase:') || key.startsWith('firebase_')) {
                        sessionStorage.removeItem(key);
                    }
                }
                console.log('Browser storage cleared');
            } catch (error) {
                console.warn('Failed to clear browser storage (non-critical):', error);
            }
            
            // Step 5: Redirect after logout complete
            console.log('Logout process completed successfully');
            window.location.href = '/?logout=success';
            
        } catch (error) {
            console.error('Sign out error:', error);
            this.showError('Failed to sign out. Please try again.');
        } finally {
            // Always reset logout state
            this.isLoggingOut = false;
        }
    }

    handleAuthError(error) {
        let message = 'An authentication error occurred';
        
        switch (error.code) {
            case 'auth/user-not-found':
            case 'auth/wrong-password':
                message = 'Invalid email or password';
                break;
            case 'auth/email-already-in-use':
                message = 'An account with this email already exists';
                break;
            case 'auth/weak-password':
                message = 'Password is too weak';
                break;
            case 'auth/invalid-email':
                message = 'Invalid email address';
                break;
            case 'auth/operation-not-allowed':
                message = 'This sign-in method is not enabled';
                break;
            case 'auth/popup-closed-by-user':
                message = 'Sign-in popup was closed';
                break;
            case 'auth/popup-blocked':
                message = 'Sign-in popup was blocked by browser';
                break;
            case 'auth/too-many-requests':
                message = 'Too many failed attempts. Please try again later';
                break;
            default:
                console.error('Unhandled auth error:', error);
                message = error.message || message;
        }
        
        this.showError(message);
    }

    showError(message) {
        const errorDiv = document.getElementById('auth-error');
        const errorMessage = document.getElementById('auth-error-message');
        
        if (errorDiv && errorMessage) {
            errorMessage.textContent = message;
            errorDiv.classList.remove('hidden');
            
            // Auto-hide after 10 seconds
            setTimeout(() => {
                errorDiv.classList.add('hidden');
            }, 10000);
        } else {
            // Fallback to alert if error div not found
            alert(message);
        }
    }

    showSuccess(message) {
        // Create success message element if it doesn't exist
        let successDiv = document.getElementById('auth-success');
        if (!successDiv) {
            successDiv = document.createElement('div');
            successDiv.id = 'auth-success';
            successDiv.className = 'mt-4 p-4 border border-green-300 rounded-md bg-green-50';
            successDiv.innerHTML = `
                <div class="flex">
                    <div class="flex-shrink-0">
                        <svg class="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor">
                            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
                        </svg>
                    </div>
                    <div class="ml-3">
                        <h3 class="text-sm font-medium text-green-800">Success</h3>
                        <p id="auth-success-message" class="mt-1 text-sm text-green-700"></p>
                    </div>
                </div>
            `;
            
            const container = document.getElementById('firebase-auth-container');
            if (container) {
                container.appendChild(successDiv);
            }
        }
        
        const successMessage = document.getElementById('auth-success-message');
        if (successMessage) {
            successMessage.textContent = message;
            successDiv.classList.remove('hidden');
            
            // Auto-hide after 5 seconds
            setTimeout(() => {
                successDiv.classList.add('hidden');
            }, 5000);
        }
    }
}

// Initialize Firebase Auth when page loads
document.addEventListener('DOMContentLoaded', () => {
    window.firebaseAuth = new FirebaseAuthManager();
});

// Export for global access
window.FirebaseAuthManager = FirebaseAuthManager;

// Global logout function for UI components
window.logout = function() {
    if (window.firebaseAuth) {
        window.firebaseAuth.signOutUser();
    } else {
        // Fallback to backend logout if Firebase not initialized
        window.location.href = '/auth/logout';
    }
};
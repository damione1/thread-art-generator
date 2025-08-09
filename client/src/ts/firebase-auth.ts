// Firebase Authentication Module
// Handles Firebase Web SDK integration with emulator support

import { initializeApp, FirebaseApp } from "firebase/app";
import {
  Auth,
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
  inMemoryPersistence,
  User,
  AuthError,
} from "firebase/auth";
import {
  FirebaseStorage,
  getStorage,
  connectStorageEmulator,
  ref,
  uploadBytes,
  uploadBytesResumable,
  getDownloadURL,
  StorageReference,
} from "firebase/storage";

interface FirebaseConfig {
  projectId: string;
  apiKey: string;
  authDomain: string;
  isEmulator: boolean;
  emulatorHost?: string;
  emulatorUI?: string;
}

interface SyncResponse {
  success: boolean;
  message?: string;
}

interface TokenCache {
  token: string;
  expires: number;
  user: User;
}

class FirebaseAuthManager {
  private app: FirebaseApp | null = null;
  private auth: Auth | null = null;
  private storage: FirebaseStorage | null = null;
  private isEmulator: boolean = false;
  private isLoggingOut: boolean = false; // Track logout state to prevent race conditions
  private justSignedIn: boolean = false; // Track fresh sign-ins to control redirects

  // Token caching and optimization
  private tokenCache: TokenCache | null = null;
  private refreshTimer: NodeJS.Timeout | null = null;
  private syncInProgress: boolean = false;
  private lastSyncTime: number = 0;
  private readonly SYNC_DEBOUNCE_MS = 1000; // Prevent multiple sync calls within 1 second
  private readonly TOKEN_REFRESH_BUFFER_MS = 5 * 60 * 1000; // Refresh 5 minutes before expiry

  constructor() {
    this.init();
  }

  async init(): Promise<void> {
    try {
      console.log('🔥 Firebase Auth initialization starting...');

      // Get Firebase configuration from Go backend via templ.JSONScript
      const configElement = document.getElementById("firebase-config");
      console.log('🔍 Config element found:', !!configElement);

      if (!configElement) {
        console.error('❌ Firebase configuration script element not found');
        this.showError("Firebase configuration not found. Please refresh the page.");
        this.hideLoadingShowUI();
        return;
      }

      console.log('📄 Config element content:', configElement.textContent);

      const config: FirebaseConfig = JSON.parse(configElement.textContent || '{}');
      console.log('⚙️ Parsed config:', config);

      if (!config) {
        throw new Error("Firebase configuration not found");
      }

      if (!config.projectId || !config.apiKey) {
        console.error('❌ Invalid Firebase config - missing required fields:', config);
        this.showError("Invalid Firebase configuration. Please check your setup.");
        this.hideLoadingShowUI();
        return;
      }

      // Initialize Firebase with the configuration
      this.app = initializeApp({
        projectId: config.projectId,
        apiKey: config.apiKey,
        authDomain: config.authDomain,
        storageBucket: `${config.projectId}.appspot.com`, // Add default storage bucket
      });
      this.auth = getAuth(this.app);
      this.storage = getStorage(this.app);

      // Connect to emulators if configured
      if (config.isEmulator && config.emulatorHost) {
        this.isEmulator = true;
        await this.connectToEmulators(config.emulatorHost);
      } else {
        console.log("🌐 Using Firebase production environment");
      }

      // Set up auth state listener
      this.setupAuthStateListener();

      // Wait for initial auth state to settle before showing UI
      console.log("⏳ Waiting for initial auth state to settle...");
      await this.auth.authStateReady();
      console.log("✅ Initial auth state ready");

      // Initialize UI
      this.initializeUI();

      console.log("✅ Firebase Auth initialized successfully");
      console.log("📋 Config:", config);
      console.log("🔧 Emulator mode:", this.isEmulator);
      console.log("👤 Current user:", this.auth.currentUser?.uid || 'none');
    } catch (error) {
      console.error("❌ Firebase initialization error:", error);
      this.showError("Failed to initialize authentication");
      this.hideLoadingShowUI();
    }
  }

  private setupAuthStateListener(): void {
    if (!this.auth) return;

    let authStateChangeTimeout: NodeJS.Timeout | null = null; // Debounce rapid auth state changes

    onAuthStateChanged(this.auth, async (user: User | null) => {
      // Debounce rapid auth state changes to prevent race conditions
      if (authStateChangeTimeout) {
        clearTimeout(authStateChangeTimeout);
      }

      authStateChangeTimeout = setTimeout(async () => {
        if (user) {
          console.log("User signed in:", user.uid);

          // Don't auto-redirect if we're in the middle of logging out
          if (this.isLoggingOut) {
            console.log("Logout in progress, ignoring auth state change");
            return;
          }

          // Don't auto-redirect if we're on auth pages and user didn't just sign in
          const currentPath = window.location.pathname;
          if (["/login", "/signup"].includes(currentPath) && !this.justSignedIn) {
            console.log("On auth page without fresh sign-in, not redirecting");
            return;
          }

          try {
            // Use cached token if available and valid, otherwise get fresh token
            await this.ensureValidToken(user);

            // Only redirect if this was a fresh sign-in and we're not already on dashboard
            // or if we're on auth pages (login/signup) and user is authenticated
            if (this.justSignedIn && currentPath !== "/dashboard") {
              window.location.href = "/dashboard";
            } else if (["/login", "/signup"].includes(currentPath)) {
              // User is authenticated but on auth page - redirect to dashboard
              window.location.href = "/dashboard";
            }

            this.justSignedIn = false; // Reset the flag
          } catch (error) {
            console.error("Error in auth state handling:", error);

            // Check if this is an "authEvent" related error and handle appropriately
            if (error instanceof Error && error.message.includes('authEvent')) {
              console.error("Auth event communication error detected:", error.message);
              this.showError("Authentication service temporarily unavailable. Please try refreshing the page.");

              // Don't immediately sign out for authEvent errors - give user a chance to retry
              return;
            }

            this.showError("Failed to complete sign in. Please try again.");

            // On sync failure, sign out to prevent ghost state (but not for authEvent errors)
            if (!(error instanceof Error) || !error.message?.includes('authEvent')) {
              try {
                await this.signOutUser();
              } catch (signOutError) {
                console.error("Failed to sign out after sync error:", signOutError);
              }
            }
          }
        } else {
          console.log("User signed out");

          // Clear token cache and refresh timer
          this.clearTokenCache();

          // Only clear backend session if we're not already logging out
          if (!this.isLoggingOut) {
            try {
              const response = await fetch("/auth/logout", {
                method: "POST",
                headers: {
                  "Content-Type": "application/json",
                }
              });

              if (!response.ok) {
                console.warn("Backend session cleanup failed (non-critical)");
              }
            } catch (error) {
              console.log("Logout cleanup error (non-critical):", error);
            }
          }
        }
      }, 100); // 100ms debounce to prevent rapid state changes
    });
  }

  private async syncWithBackend(idToken: string): Promise<SyncResponse> {
    const response = await fetch("/auth/sync", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        id_token: idToken,
      }),
    });

    if (!response.ok) {
      const errorData = await response
        .json()
        .catch(() => ({ message: "Unknown error" }));
      throw new Error(errorData.message || "Failed to sync with backend");
    }

    return response.json();
  }

  private async connectToEmulators(emulatorHost: string, maxRetries: number = 3): Promise<void> {
    if (!this.auth || !this.storage) return;

    // Check if emulators are already connected to prevent duplicate connections
    if ((this.auth as any)._delegate?._config?.emulator) {
      console.log("🔥 Firebase Auth Emulator already connected");
      return;
    }

    const authEmulatorURL = `http://${emulatorHost}`;

    // Connect to Auth Emulator with retry and improved error handling
    let authConnected = false;
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        console.log(`🔥 Connecting to Firebase Auth Emulator at: ${authEmulatorURL} (attempt ${attempt}/${maxRetries})`);

        // Use a timeout for the connection attempt
        const connectionPromise = new Promise<void>((resolve, reject) => {
          try {
            connectAuthEmulator(this.auth!, authEmulatorURL, {
              disableWarnings: true,
            });
            resolve();
          } catch (error) {
            reject(error);
          }
        });

        const timeoutPromise = new Promise<never>((_, reject) =>
          setTimeout(() => reject(new Error('Connection timeout')), 5000)
        );

        await Promise.race([connectionPromise, timeoutPromise]);

        console.log("✅ Firebase Auth Emulator connected successfully");
        authConnected = true;
        break;
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown error';
        console.warn(`❌ Auth emulator connection attempt ${attempt} failed:`, errorMessage);
        if (errorMessage.includes('already connected') || errorMessage.includes('Auth Emulator has already been started')) {
          console.log("✅ Firebase Auth Emulator already connected (detected via error)");
          authConnected = true;
          break;
        }

        if (attempt === maxRetries) {
          console.error("❌ Failed to connect to Firebase Auth Emulator after all retries. Continuing with production config.");
          this.showError("Warning: Could not connect to local authentication. This may cause sign-in issues.");
          this.isEmulator = false;
          return; // Exit early if auth emulator fails
        } else {
          // Exponential backoff with jitter
          const delay = Math.min(1000 * Math.pow(2, attempt - 1) + Math.random() * 1000, 10000);
          console.log(`Retrying in ${Math.round(delay)}ms...`);
          await new Promise(resolve => setTimeout(resolve, delay));
        }
      }
    }

    // Connect to Storage Emulator with retry (only if Auth emulator connected)
    if (authConnected && this.isEmulator) {
      // Check if storage emulator is already connected
      if ((this.storage as any)._delegate?._host?.includes('localhost:9199')) {
        console.log("🔥 Firebase Storage Emulator already connected");
        return;
      }

      for (let attempt = 1; attempt <= maxRetries; attempt++) {
        try {
          console.log(`🔥 Connecting to Firebase Storage Emulator at: localhost:9199 (attempt ${attempt}/${maxRetries})`);

          const connectionPromise = new Promise<void>((resolve, reject) => {
            try {
              connectStorageEmulator(this.storage!, "localhost", 9199);
              resolve();
            } catch (error) {
              reject(error);
            }
          });

          const timeoutPromise = new Promise<never>((_, reject) =>
            setTimeout(() => reject(new Error('Storage connection timeout')), 5000)
          );

          await Promise.race([connectionPromise, timeoutPromise]);

          console.log("✅ Firebase Storage Emulator connected successfully");
          break;
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Unknown error';
          console.warn(`❌ Storage emulator connection attempt ${attempt} failed:`, errorMessage);
          if (errorMessage.includes('already connected') || errorMessage.includes('Storage Emulator has already been started')) {
            console.log("✅ Firebase Storage Emulator already connected (detected via error)");
            break;
          }

          if (attempt === maxRetries) {
            console.error("❌ Failed to connect to Firebase Storage Emulator after all retries");
            this.showError("Warning: Could not connect to local storage. File uploads may not work properly.");
          } else {
            const delay = Math.min(1000 * Math.pow(2, attempt - 1) + Math.random() * 1000, 8000);
            await new Promise(resolve => setTimeout(resolve, delay));
          }
        }
      }
    }
  }

  // Token caching and management methods
  private async ensureValidToken(user: User): Promise<string> {
    const now = Date.now();

    // Check if we have a valid cached token
    if (this.tokenCache && this.tokenCache.expires > now + this.TOKEN_REFRESH_BUFFER_MS) {
      console.log("Using cached Firebase token");
      return this.tokenCache.token;
    }

    // Debounce multiple concurrent sync requests
    if (this.syncInProgress) {
      console.log("Token sync already in progress, waiting...");
      // Wait for ongoing sync to complete
      while (this.syncInProgress) {
        await new Promise(resolve => setTimeout(resolve, 100));
      }
      // Return cached token if available after wait
      if (this.tokenCache && this.tokenCache.expires > now) {
        return this.tokenCache.token;
      }
    }

    // Check sync debounce
    if (now - this.lastSyncTime < this.SYNC_DEBOUNCE_MS) {
      console.log("Sync debounced, using existing token");
      if (this.tokenCache) {
        return this.tokenCache.token;
      }
    }

    return await this.refreshTokenAndSync(user);
  }

  private async refreshTokenAndSync(user: User): Promise<string> {
    this.syncInProgress = true;
    this.lastSyncTime = Date.now();

    try {
      console.log("Refreshing Firebase token...");

      // Get fresh token with shorter timeout
      const tokenPromise = user.getIdToken(true);
      const timeoutPromise = new Promise<never>((_, reject) =>
        setTimeout(() => reject(new Error('Token retrieval timeout')), 5000) // Reduced from 10s to 5s
      );

      const idToken = await Promise.race([tokenPromise, timeoutPromise]);

      // Parse token to get expiration
      const tokenPayload = JSON.parse(atob(idToken.split('.')[1]));
      const expires = tokenPayload.exp * 1000; // Convert to milliseconds

      // Cache the token
      this.tokenCache = {
        token: idToken,
        expires: expires,
        user: user
      };

      // Schedule background refresh
      this.scheduleTokenRefresh(expires);

      // Send token to backend for session creation with retry logic
      await this.syncWithBackendRetry(idToken);

      console.log("Token refreshed and cached successfully");
      return idToken;

    } finally {
      this.syncInProgress = false;
    }
  }

  private scheduleTokenRefresh(expires: number): void {
    // Clear existing refresh timer
    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer);
    }

    // Schedule refresh 5 minutes before expiration
    const refreshTime = expires - Date.now() - this.TOKEN_REFRESH_BUFFER_MS;

    if (refreshTime > 0) {
      console.log(`Scheduling token refresh in ${Math.round(refreshTime / 1000 / 60)} minutes`);
      this.refreshTimer = setTimeout(async () => {
        if (this.auth?.currentUser) {
          try {
            await this.refreshTokenAndSync(this.auth.currentUser);
          } catch (error) {
            console.warn("Background token refresh failed:", error);
          }
        }
      }, refreshTime);
    }
  }

  private clearTokenCache(): void {
    this.tokenCache = null;
    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer);
      this.refreshTimer = null;
    }
  }

  private async syncWithBackendRetry(idToken: string, maxRetries: number = 3): Promise<SyncResponse> {
    let lastError: Error | null = null;

    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        console.log(`Syncing with backend (attempt ${attempt}/${maxRetries})`);
        return await this.syncWithBackend(idToken);
      } catch (error) {
        lastError = error instanceof Error ? error : new Error('Unknown error');
        console.warn(`Backend sync attempt ${attempt} failed:`, lastError.message);

        if (attempt < maxRetries) {
          // Exponential backoff: wait 1s, 2s, 4s
          const delay = Math.pow(2, attempt - 1) * 1000;
          console.log(`Retrying in ${delay}ms...`);
          await new Promise(resolve => setTimeout(resolve, delay));
        }
      }
    }

    console.error(`Backend sync failed after ${maxRetries} attempts`);
    throw lastError;
  }

  private initializeUI(): void {
    console.log("🎨 Initializing Firebase Auth UI...");
    this.hideLoadingShowUI();
    this.setupEventListeners();
  }

  private hideLoadingShowUI(): void {
    // Hide loading, show auth buttons
    const loading = document.getElementById("auth-loading");
    const buttons = document.getElementById("auth-buttons");

    console.log("🎯 UI elements found - loading:", !!loading, "buttons:", !!buttons);

    if (loading) {
      loading.classList.add("hidden");
      console.log("⚡ Loading spinner hidden");
    }
    if (buttons) {
      buttons.classList.remove("hidden");
      console.log("⚡ Auth buttons shown");
    }
  }

  private setupEventListeners(): void {
    // Google Sign In
    const googleSignInBtn = document.getElementById("google-signin-btn");
    const googleSignUpBtn = document.getElementById("google-signup-btn");

    if (googleSignInBtn) {
      googleSignInBtn.addEventListener("click", () => this.signInWithGoogle());
    }
    if (googleSignUpBtn) {
      googleSignUpBtn.addEventListener("click", () => this.signInWithGoogle());
    }

    // Email/Password Sign In
    const emailSignInForm = document.getElementById("email-signin-form");
    if (emailSignInForm) {
      emailSignInForm.addEventListener("submit", (e) => {
        e.preventDefault();
        this.signInWithEmail();
      });
    }

    // Email/Password Sign Up
    const emailSignUpForm = document.getElementById("email-signup-form");
    if (emailSignUpForm) {
      emailSignUpForm.addEventListener("submit", (e) => {
        e.preventDefault();
        this.signUpWithEmail();
      });
    }

    // Forgot Password
    const forgotPasswordLink = document.getElementById("forgot-password-link");
    if (forgotPasswordLink) {
      forgotPasswordLink.addEventListener("click", (e) => {
        e.preventDefault();
        this.resetPassword();
      });
    }
  }

  async signInWithGoogle(): Promise<void> {
    if (!this.auth) return;

    try {
      this.justSignedIn = true; // Mark as fresh sign-in
      const provider = new GoogleAuthProvider();
      // Add scopes for profile information
      provider.addScope("profile");
      provider.addScope("email");

      const result = await signInWithPopup(this.auth, provider);
      console.log("Google sign in successful:", result.user.uid);
    } catch (error) {
      this.justSignedIn = false; // Reset on error
      console.error("Google sign in error:", error);
      this.handleAuthError(error as AuthError);
    }
  }

  async signInWithEmail(): Promise<void> {
    if (!this.auth) return;

    const emailElement = document.getElementById("email-address") as HTMLInputElement;
    const passwordElement = document.getElementById("password") as HTMLInputElement;

    const email = emailElement?.value;
    const password = passwordElement?.value;

    if (!email || !password) {
      this.showError("Please enter both email and password");
      return;
    }

    try {
      this.justSignedIn = true; // Mark as fresh sign-in
      const result = await signInWithEmailAndPassword(this.auth, email, password);
      console.log("Email sign in successful:", result.user.uid);
    } catch (error) {
      this.justSignedIn = false; // Reset on error
      console.error("Email sign in error:", error);
      this.handleAuthError(error as AuthError);
    }
  }

  async signUpWithEmail(): Promise<void> {
    if (!this.auth) return;

    const emailElement = document.getElementById("signup-email") as HTMLInputElement;
    const passwordElement = document.getElementById("signup-password") as HTMLInputElement;
    const confirmPasswordElement = document.getElementById("signup-confirm-password") as HTMLInputElement;
    const termsElement = document.getElementById("terms-agreement") as HTMLInputElement;

    const email = emailElement?.value;
    const password = passwordElement?.value;
    const confirmPassword = confirmPasswordElement?.value;
    const termsAccepted = termsElement?.checked;

    // Validation
    if (!email || !password || !confirmPassword) {
      this.showError("Please fill in all fields");
      return;
    }

    if (password !== confirmPassword) {
      this.showError("Passwords do not match");
      return;
    }

    if (password.length < 6) {
      this.showError("Password must be at least 6 characters long");
      return;
    }

    if (!termsAccepted) {
      this.showError("Please accept the Terms of Service and Privacy Policy");
      return;
    }

    try {
      this.justSignedIn = true; // Mark as fresh sign-in
      const result = await createUserWithEmailAndPassword(this.auth, email, password);
      console.log("Email sign up successful:", result.user.uid);
    } catch (error) {
      this.justSignedIn = false; // Reset on error
      console.error("Email sign up error:", error);
      this.handleAuthError(error as AuthError);
    }
  }

  async resetPassword(): Promise<void> {
    if (!this.auth) return;

    const emailElement = document.getElementById("email-address") as HTMLInputElement;
    const email = emailElement?.value;

    if (!email) {
      this.showError("Please enter your email address first");
      return;
    }

    try {
      await sendPasswordResetEmail(this.auth, email);
      this.showSuccess("Password reset email sent! Check your inbox.");
    } catch (error) {
      console.error("Password reset error:", error);
      this.handleAuthError(error as AuthError);
    }
  }

  async signOutUser(): Promise<void> {
    if (!this.auth) return;

    try {
      console.log("Starting logout process...");
      this.isLoggingOut = true; // Set logout state to prevent race conditions
      this.justSignedIn = false; // Reset sign-in flag

      // Step 1: Clear token cache and timers
      this.clearTokenCache();

      // Step 2: Clear backend session first
      try {
        const response = await fetch("/auth/logout", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
        });
        if (response.ok) {
          console.log("Backend session cleared successfully");
        } else {
          console.warn("Backend logout failed, continuing with Firebase logout");
        }
      } catch (error) {
        console.warn("Backend logout error (continuing):", error);
      }

      // Step 3: Clear Firebase auth state
      await signOut(this.auth);
      console.log("Firebase signOut completed");

      // Step 4: Force clear any persisted auth state
      try {
        // Clear auth persistence to ensure no auto-login
        await setPersistence(this.auth, inMemoryPersistence);
        console.log("Auth persistence cleared");
      } catch (error) {
        console.warn("Failed to clear persistence (non-critical):", error);
      }

      // Step 5: Clear browser storage manually as backup
      try {
        // Clear any Firebase-related local storage
        for (const key of Object.keys(localStorage)) {
          if (key.startsWith("firebase:") || key.startsWith("firebase_")) {
            localStorage.removeItem(key);
          }
        }

        // Clear any Firebase-related session storage
        for (const key of Object.keys(sessionStorage)) {
          if (key.startsWith("firebase:") || key.startsWith("firebase_")) {
            sessionStorage.removeItem(key);
          }
        }
        console.log("Browser storage cleared");
      } catch (error) {
        console.warn("Failed to clear browser storage (non-critical):", error);
      }

      // Step 6: Redirect after logout complete
      console.log("Logout process completed successfully");
      window.location.href = "/?logout=success";
    } catch (error) {
      console.error("Sign out error:", error);
      this.showError("Failed to sign out. Please try again.");
    } finally {
      // Always reset logout state
      this.isLoggingOut = false;
    }
  }

  private handleAuthError(error: AuthError | Error): void {
    let message = "An authentication error occurred";

    // Check for specific emulator communication errors first
    if (error.message) {
      if (error.message.includes('authEvent') || error.message.includes('Sending authEvent failed')) {
        message = "Authentication service communication error. Please check that the Firebase emulator is running and try refreshing the page.";
        console.error("Firebase Auth Emulator communication failure:", error);
        this.showError(message);
        return;
      }

      if (error.message.includes('Connection timeout') || error.message.includes('Network Error')) {
        message = "Authentication service is temporarily unavailable. Please check your connection and try again.";
        this.showError(message);
        return;
      }

      if (error.message.includes('Auth Emulator has already been started')) {
        console.log("Auth emulator already started - this is not an error");
        return; // Don't show error for this case
      }
    }

    const authError = error as AuthError;
    switch (authError.code) {
      case "auth/user-not-found":
      case "auth/wrong-password":
        message = "Invalid email or password";
        break;
      case "auth/email-already-in-use":
        message = "An account with this email already exists";
        break;
      case "auth/weak-password":
        message = "Password is too weak";
        break;
      case "auth/invalid-email":
        message = "Invalid email address";
        break;
      case "auth/operation-not-allowed":
        message = "This sign-in method is not enabled";
        break;
      case "auth/popup-closed-by-user":
        message = "Sign-in popup was closed";
        break;
      case "auth/popup-blocked":
        message = "Sign-in popup was blocked by browser";
        break;
      case "auth/too-many-requests":
        message = "Too many failed attempts. Please try again later";
        break;
      case "auth/network-request-failed":
        message = "Network error. Please check your internet connection and try again.";
        break;
      case "auth/internal-error":
        message = "Authentication service error. Please try again or contact support if the problem persists.";
        break;
      default:
        console.error("Unhandled auth error:", error);
        // For emulator-specific errors, provide more helpful guidance
        if (this.isEmulator) {
          message = "Local authentication error. Please ensure the Firebase emulator is running properly and try refreshing the page.";
        } else {
          message = error.message || message;
        }
    }

    this.showError(message);
  }

  private showError(message: string): void {
    const errorDiv = document.getElementById("auth-error");
    const errorMessage = document.getElementById("auth-error-message");

    if (errorDiv && errorMessage) {
      errorMessage.textContent = message;
      errorDiv.classList.remove("hidden");

      // Auto-hide after 10 seconds
      setTimeout(() => {
        errorDiv.classList.add("hidden");
      }, 10000);
    } else {
      // Fallback to alert if error div not found
      alert(message);
    }
  }

  private showSuccess(message: string): void {
    // Create success message element if it doesn't exist
    let successDiv = document.getElementById("auth-success");
    if (!successDiv) {
      successDiv = document.createElement("div");
      successDiv.id = "auth-success";
      successDiv.className = "mt-4 p-4 border border-green-300 rounded-md bg-green-50";
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

      const container = document.getElementById("firebase-auth-container");
      if (container) {
        container.appendChild(successDiv);
      }
    }

    const successMessage = document.getElementById("auth-success-message");
    if (successMessage) {
      successMessage.textContent = message;
      successDiv.classList.remove("hidden");

      // Auto-hide after 5 seconds
      setTimeout(() => {
        successDiv.classList.add("hidden");
      }, 5000);
    }
  }

  // Public methods for accessing Firebase services
  public getAuth(): Auth | null {
    return this.auth;
  }

  public getStorage(): FirebaseStorage | null {
    return this.storage;
  }

  public getStorageRef(path: string): StorageReference | null {
    if (this.storage) {
      return ref(this.storage, path);
    }
    return null;
  }
}

// Initialize Firebase Auth when page loads, only if config is available
document.addEventListener("DOMContentLoaded", () => {
  console.log('🚀 DOM Content Loaded - initializing Firebase Auth...');

  const configElement = document.getElementById("firebase-config");
  console.log('🔍 Firebase config element found:', !!configElement);

  if (configElement) {
    console.log('📄 Config element content:', configElement.textContent);
    console.log('✅ Starting FirebaseAuthManager...');
    (window as any).firebaseAuth = new FirebaseAuthManager();
  } else {
    console.error("❌ Firebase config element not found - Firebase auth not initialized");
    console.log('🔍 Available script elements:', Array.from(document.querySelectorAll('script[type="application/json"]')).map(el => ({ id: el.id, content: el.textContent?.substring(0, 100) })));

    // Show user-friendly error
    const errorDiv = document.getElementById("auth-error");
    const errorMessage = document.getElementById("auth-error-message");
    if (errorDiv && errorMessage) {
      errorMessage.textContent = "Firebase configuration is missing. Please refresh the page or contact support.";
      errorDiv.classList.remove("hidden");
    }

    // Hide loading spinner
    const loading = document.getElementById("auth-loading");
    if (loading) loading.classList.add("hidden");
  }
});

// Export for global access and prevent tree-shaking
(window as any).FirebaseAuthManager = FirebaseAuthManager;
export { FirebaseAuthManager };

// Make Firebase services available globally for templates
(window as any).firebase = {
  auth: () => (window as any).firebaseAuth?.getAuth(),
  storage: () => (window as any).firebaseAuth?.getStorage(),
  // Helper functions for common Firebase operations
  storageRef: (path: string) => {
    if ((window as any).firebaseAuth?.getStorage()) {
      return ref((window as any).firebaseAuth.getStorage(), path);
    }
    return null;
  },
  uploadBytes: uploadBytes,
  uploadBytesResumable: uploadBytesResumable,
  getDownloadURL: getDownloadURL,
};

// Global logout function for UI components
(window as any).logout = function () {
  if ((window as any).firebaseAuth) {
    (window as any).firebaseAuth.signOutUser();
  } else {
    // Fallback to backend logout if Firebase not initialized
    window.location.href = "/auth/logout";
  }
};

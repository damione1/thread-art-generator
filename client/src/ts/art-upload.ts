// TypeScript interface for secure upload functionality via BFF
declare global {
  interface Window {
    firebase: {
      auth(): any;
      storage(): any;
      storageRef(path: string): any;
      uploadBytesResumable(ref: any, file: File, options?: any): any;
    };
  }
}

export interface ArtUploadConfig {
  artId: string;
  internalUserId: string;
}

export interface ArtUploadState {
  dragOver: boolean;
  uploading: boolean;
  uploaded: boolean;
  error: boolean;
  errorMessage: string;
  uploadProgress: number;
  firebaseReady: boolean;
}

export interface ArtUploadActions {
  handleDrop(event: DragEvent): void;
  handleFileSelect(event: Event): void;
  uploadFile(file: File): Promise<void>;
  showError(message: string): void;
  resetUpload(): void;
  refreshPage(): void;
  waitForFirebaseReady(): Promise<void>;
}

export type ArtUpload = ArtUploadState & ArtUploadActions;

// Request/Response interfaces for BFF API
interface UploadURLRequest {
  artId: string;
  contentType: string;
  fileSize: number;
  filename?: string;
}

interface UploadURLResponse {
  uploadUrl: string;
  storagePath: string;
  expiresAt: string;
  maxFileSize: number;
}

interface ErrorResponse {
  error: string;
  message?: string;
}

export function createArtUpload(config: ArtUploadConfig): ArtUpload {
  const state: ArtUploadState = {
    dragOver: false,
    uploading: false,
    uploaded: false,
    error: false,
    errorMessage: '',
    uploadProgress: 0,
    firebaseReady: true, // Start as ready - handle Firebase readiness asynchronously
  };

  const actions: ArtUploadActions = {
    async waitForFirebaseReady(): Promise<void> {
      // Quick check for Firebase readiness - no artificial delays
      let attempts = 0;
      const maxAttempts = 100; // 10 seconds max wait time with faster polling
      
      while (attempts < maxAttempts) {
        // Check if Firebase is available and initialized
        if (window.firebase && window.firebase.auth()) {
          const auth = window.firebase.auth();
          if (auth.currentUser) {
            console.log('Firebase ready with authenticated user:', auth.currentUser.uid);
            return; // No need to update state, already ready
          }
        }
        
        // Faster polling - 100ms intervals
        await new Promise(resolve => setTimeout(resolve, 100));
        attempts++;
      }
      
      throw new Error('Firebase authentication not ready. Please refresh the page and try signing in again.');
    },

    handleDrop(event: DragEvent): void {
      state.dragOver = false;
      const files = event.dataTransfer?.files;
      if (files && files.length > 0) {
        actions.uploadFile(files[0]);
      }
    },

    handleFileSelect(event: Event): void {
      const input = event.target as HTMLInputElement;
      const files = input.files;
      if (files && files.length > 0) {
        actions.uploadFile(files[0]);
      }
    },

    async uploadFile(file: File): Promise<void> {
      // Reset any previous errors
      state.error = false;
      state.errorMessage = '';

      try {
        // Validate file type
        if (!file.type.startsWith('image/')) {
          actions.showError('Please select an image file');
          return;
        }

        // Validate file size (10MB limit)
        if (file.size > 10 * 1024 * 1024) {
          actions.showError('File size must be less than 10MB');
          return;
        }

        // Set uploading state
        state.uploading = true;
        state.uploadProgress = 0;

        console.log('Starting secure upload process via BFF...');

        // Step 1: Get signed URL from BFF
        const uploadRequest: UploadURLRequest = {
          artId: config.artId,
          contentType: file.type,
          fileSize: file.size,
          filename: file.name,
        };

        console.log('Requesting signed upload URL from BFF...', uploadRequest);

        const urlResponse = await fetch('/api/storage/upload-url', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(uploadRequest),
          credentials: 'include', // Include session cookies
        });

        if (!urlResponse.ok) {
          const errorData: ErrorResponse = await urlResponse.json().catch(() => ({ error: 'Unknown error' }));
          throw new Error(errorData.message || errorData.error || `HTTP ${urlResponse.status}`);
        }

        const urlData: UploadURLResponse = await urlResponse.json();
        console.log('Received signed upload URL:', {
          storagePath: urlData.storagePath,
          expiresAt: urlData.expiresAt,
          maxFileSize: urlData.maxFileSize
        });

        // Step 2: Upload file directly to Firebase Storage using two-step process for emulator
        console.log('Uploading file to Firebase Storage with proper metadata handling...');

        // Use XMLHttpRequest for progress tracking
        const xhr = new XMLHttpRequest();
        
        // Set up progress tracking
        xhr.upload.addEventListener('progress', (event) => {
          if (event.lengthComputable) {
            const progress = Math.round((event.loaded / event.total) * 100);
            state.uploadProgress = progress;
            console.log('Upload progress:', progress + '%');
          }
        });

        // First, upload the file with PUT (this will set content-type as application/octet-stream)
        await new Promise<void>((resolve, reject) => {
          xhr.addEventListener('load', async () => {
            if (xhr.status >= 200 && xhr.status < 300) {
              console.log('File uploaded successfully, now updating metadata...');
              
              // Step 3: For Firebase Storage emulator, update metadata with PATCH request
              // This fixes the content-type from application/octet-stream to the correct MIME type
              try {
                const isEmulator = urlData.uploadUrl.includes('localhost') || urlData.uploadUrl.includes('127.0.0.1');
                
                if (isEmulator) {
                  // Create metadata object for PATCH request
                  const metadata = {
                    contentType: file.type
                  };
                  
                  // PATCH request to update metadata
                  console.log('Sending PATCH request to update metadata with:', metadata);
                  console.log('PATCH URL:', urlData.uploadUrl);
                  
                  const metadataResponse = await fetch(urlData.uploadUrl, {
                    method: 'PATCH',
                    headers: {
                      'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(metadata)
                  });
                  
                  console.log('PATCH response status:', metadataResponse.status);
                  console.log('PATCH response statusText:', metadataResponse.statusText);
                  
                  if (!metadataResponse.ok) {
                    const errorText = await metadataResponse.text();
                    console.warn('Metadata update failed:', metadataResponse.status, metadataResponse.statusText, 'Response:', errorText);
                  } else {
                    const responseText = await metadataResponse.text();
                    console.log('Metadata updated successfully, response:', responseText);
                    console.log('Correct content-type should now be set');
                  }
                }
                
                resolve();
              } catch (metadataError) {
                console.warn('Metadata update failed:', metadataError, 'but file was uploaded');
                // Don't fail the entire upload if metadata update fails
                resolve();
              }
            } else {
              console.error('Upload failed with status:', xhr.status, xhr.statusText);
              console.error('Response:', xhr.responseText);
              reject(new Error(`Upload failed: ${xhr.status} ${xhr.statusText}`));
            }
          });

          xhr.addEventListener('error', () => {
            console.error('Upload network error');
            reject(new Error('Network error during upload'));
          });

          xhr.addEventListener('abort', () => {
            console.error('Upload was aborted');
            reject(new Error('Upload was canceled'));
          });

          // Open connection with PUT method for Firebase Storage upload
          xhr.open('PUT', urlData.uploadUrl);
          
          // For the initial PUT request, we don't set Content-Type header
          // Firebase Storage emulator will set it as application/octet-stream initially
          xhr.setRequestHeader('Content-Length', file.size.toString());

          // Start upload with raw file data
          xhr.send(file);
        });

        // Upload completed successfully
        console.log('Upload completed successfully');
        state.uploading = false;
        state.uploaded = true;
        
        // Show success message - HTMX will handle page refresh automatically
        console.log('File uploaded successfully! HTMX will refresh when image is ready.');
        
        // No manual refresh needed - HTMX auto-refresh will handle it
        // The page will automatically update when art.GetImageUrl() is populated

      } catch (error: any) {
        console.error('Upload error:', error);
        state.uploading = false;
        
        let errorMessage = 'Upload failed. Please try again.';
        if (error.message) {
          if (error.message.includes('Firebase authentication not ready')) {
            errorMessage = error.message;
          } else if (error.message.includes('Authentication required')) {
            errorMessage = 'Please sign in to upload images.';
          } else if (error.message.includes('Unauthorized')) {
            errorMessage = 'You are not authorized to upload this file. Please sign in and try again.';
          } else if (error.message.includes('File size')) {
            errorMessage = error.message;
          } else if (error.message.includes('Invalid request')) {
            errorMessage = 'Invalid upload request. Please try again.';
          } else {
            errorMessage = `Upload failed: ${error.message}`;
          }
        }
        
        actions.showError(errorMessage);
      }
    },

    showError(message: string): void {
      state.error = true;
      state.errorMessage = message;
      state.uploading = false;
      state.uploaded = false;
      
      // Log error for debugging
      console.error('Upload error displayed to user:', message);
      
      // Auto-hide error after 10 seconds to prevent permanent error state
      setTimeout(() => {
        if (state.error && state.errorMessage === message) {
          actions.resetUpload();
        }
      }, 10000);
    },

    resetUpload(): void {
      state.error = false;
      state.errorMessage = '';
      state.uploading = false;
      state.uploaded = false;
      state.uploadProgress = 0;
      // Keep firebaseReady state
      console.log('Upload state reset');
    },

    refreshPage(): void {
      console.log('Manual page refresh triggered');
      window.location.reload();
    }
  };

  // No preemptive Firebase check - handle on demand during upload
  // This eliminates the artificial loading delay

  return { ...state, ...actions };
}

// Make the function available globally for use in templates
(window as any).createArtUpload = createArtUpload;
import { Code, ConnectError } from '@connectrpc/connect';
import { createArtClient } from './rpc';

const artClient = createArtClient();

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
}

export interface ArtUploadActions {
  handleDrop(event: DragEvent): void;
  handleFileSelect(event: Event): void;
  uploadFile(file: File): Promise<void>;
  showError(message: string): void;
  resetUpload(): void;
  refreshPage(): void;
}

export type ArtUpload = ArtUploadState & ArtUploadActions;

function putFile(
  url: string,
  file: File,
  headers: { [key: string]: string },
  method: string,
  onProgress: (pct: number) => void,
): Promise<void> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();

    xhr.upload.addEventListener('progress', (event) => {
      if (event.lengthComputable) {
        onProgress(Math.round((event.loaded / event.total) * 100));
      }
    });

    xhr.addEventListener('load', () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve();
      } else {
        reject(new Error(`Upload failed: ${xhr.status} ${xhr.statusText}`));
      }
    });

    xhr.addEventListener('error', () => {
      reject(new Error('Network error during upload'));
    });

    xhr.addEventListener('abort', () => {
      reject(new Error('Upload was canceled'));
    });

    xhr.open(method || 'PUT', url);
    for (const [key, value] of Object.entries(headers)) {
      xhr.setRequestHeader(key, value);
    }
    xhr.send(file);
  });
}

function uploadErrorMessage(error: unknown): string {
  const connectErr = ConnectError.from(error);

  switch (connectErr.code) {
    case Code.Unauthenticated:
      return 'Please sign in to upload images.';
    case Code.PermissionDenied:
      return 'You are not authorized to upload this file. Please sign in and try again.';
    case Code.InvalidArgument:
      return connectErr.message || 'Invalid upload request. Please try again.';
    default:
      break;
  }

  if (error instanceof Error && error.message) {
    if (error.message.includes('File size')) {
      return error.message;
    }
    return `Upload failed: ${error.message}`;
  }

  return 'Upload failed. Please try again.';
}

export function createArtUpload(config: ArtUploadConfig): ArtUpload {
  return {
    dragOver: false,
    uploading: false,
    uploaded: false,
    error: false,
    errorMessage: '',
    uploadProgress: 0,

    handleDrop(event: DragEvent): void {
      this.dragOver = false;
      const files = event.dataTransfer?.files;
      if (files && files.length > 0) {
        this.uploadFile(files[0]);
      }
    },

    handleFileSelect(event: Event): void {
      const input = event.target as HTMLInputElement;
      const files = input.files;
      if (files && files.length > 0) {
        this.uploadFile(files[0]);
      }
    },

    async uploadFile(file: File): Promise<void> {
      this.error = false;
      this.errorMessage = '';

      try {
        if (!file.type.startsWith('image/')) {
          this.showError('Please select an image file');
          return;
        }

        if (file.size > 10 * 1024 * 1024) {
          this.showError('File size must be less than 10MB');
          return;
        }

        this.uploading = true;
        this.uploadProgress = 0;

        const name = `users/${config.internalUserId}/arts/${config.artId}`;

        const started = await artClient.startArtUpload({
          name,
          contentType: file.type,
        });

        await putFile(
          started.uploadUrl,
          file,
          started.headers,
          started.method || 'PUT',
          (pct) => {
            this.uploadProgress = pct;
          },
        );

        await artClient.completeArtUpload({ name });

        this.uploading = false;
        this.uploaded = true;
        setTimeout(() => this.refreshPage(), 400);
      } catch (error: unknown) {
        console.error('Upload error:', error);
        this.uploading = false;
        this.showError(uploadErrorMessage(error));
      }
    },

    showError(message: string): void {
      this.error = true;
      this.errorMessage = message;
      this.uploading = false;
      this.uploaded = false;

      setTimeout(() => {
        if (this.error && this.errorMessage === message) {
          this.resetUpload();
        }
      }, 10000);
    },

    resetUpload(): void {
      this.error = false;
      this.errorMessage = '';
      this.uploading = false;
      this.uploaded = false;
      this.uploadProgress = 0;
    },

    refreshPage(): void {
      window.location.reload();
    },
  };
}

(window as any).createArtUpload = createArtUpload;

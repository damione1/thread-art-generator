import { Code, ConnectError } from '@connectrpc/connect';
import { createArtClient } from './rpc';

const artClient = createArtClient();

const MAX_SOURCE_BYTES = 30 * 1024 * 1024;
const MAX_WORKING_EDGE = 4096;
const OUTPUT_SIZE = 1600;
const MAX_ZOOM = 4;
const JPEG_QUALITY = 0.92;

export interface ArtUploadConfig {
  artId: string;
  internalUserId: string;
}

interface Point {
  x: number;
  y: number;
}

interface AlpineCtx {
  $refs: {
    fileInput: HTMLInputElement;
    viewport: HTMLElement;
  };
  $nextTick(): Promise<void>;
  $cleanup?(fn: () => void): void;
}

export interface ArtUploadState {
  dragOver: boolean;
  uploading: boolean;
  uploaded: boolean;
  error: boolean;
  errorMessage: string;
  uploadProgress: number;
  cropping: boolean;
  preparing: boolean;
  dragging: boolean;
  previewUrl: string;
  imgWidth: number;
  imgHeight: number;
  scale: number;
  minScale: number;
  offsetX: number;
  offsetY: number;
  zoomLevel: number;
  cropError: string;
  lastX: number;
  lastY: number;
  pointers: Record<string, Point>;
  pinchStartDist: number;
  pinchStartScale: number;
  workingCanvas: HTMLCanvasElement | null;
  previewObjectUrl: string;
  originalName: string;
}

export interface ArtUploadActions {
  init(): void;
  teardown(): void;
  readonly imageStyle: Record<string, string>;
  openPicker(): void;
  handleDrop(event: DragEvent): void;
  handleFileSelect(event: Event): void;
  beginCrop(file: File): Promise<void>;
  viewportSize(): number;
  fitToCover(): void;
  onResize(): void;
  clampOffsets(): void;
  setScale(nextScale: number, originX?: number, originY?: number): void;
  onZoomInput(event: Event): void;
  onWheel(event: WheelEvent): void;
  pointerCount(): number;
  onPointerDown(event: PointerEvent): void;
  onPointerMove(event: PointerEvent): void;
  onPointerUp(event: PointerEvent): void;
  cancelCrop(): void;
  releasePreview(): void;
  confirmCrop(): Promise<void>;
  exportSquare(): Promise<File>;
  uploadFile(file: File): Promise<void>;
  showError(message: string): void;
  resetUpload(): void;
  refreshPage(): void;
}

export type ArtUpload = ArtUploadState & ArtUploadActions;

function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value));
}

function distance(a: Point, b: Point): number {
  return Math.hypot(a.x - b.x, a.y - b.y);
}

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

async function decodeImage(file: File): Promise<ImageBitmap | HTMLImageElement> {
  if (typeof createImageBitmap === 'function') {
    try {
      return await createImageBitmap(file, { imageOrientation: 'from-image' });
    } catch {
      try {
        return await createImageBitmap(file);
      } catch {
        // Fall through to HTMLImageElement
      }
    }
  }

  return loadHtmlImage(file);
}

function loadHtmlImage(file: File): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const url = URL.createObjectURL(file);
    const img = new Image();
    img.onload = () => {
      URL.revokeObjectURL(url);
      resolve(img);
    };
    img.onerror = () => {
      URL.revokeObjectURL(url);
      reject(new Error('Could not read this image. Try a JPEG, PNG, or WebP file.'));
    };
    img.src = url;
  });
}

function rasterizeToCanvas(source: CanvasImageSource & { width: number; height: number }): HTMLCanvasElement {
  const fit = Math.min(1, MAX_WORKING_EDGE / Math.max(source.width, source.height));
  const width = Math.max(1, Math.round(source.width * fit));
  const height = Math.max(1, Math.round(source.height * fit));
  const canvas = document.createElement('canvas');
  canvas.width = width;
  canvas.height = height;
  const ctx = canvas.getContext('2d', { alpha: false });
  if (!ctx) {
    throw new Error('Could not process this image');
  }
  ctx.imageSmoothingEnabled = true;
  ctx.imageSmoothingQuality = 'high';
  ctx.filter = 'grayscale(1)';
  ctx.drawImage(source, 0, 0, width, height);
  ctx.filter = 'none';
  const pixels = ctx.getImageData(0, 0, width, height);
  applyGrayscale(pixels.data);
  ctx.putImageData(pixels, 0, 0);
  if ('close' in source && typeof source.close === 'function') {
    source.close();
  }
  return canvas;
}

function applyGrayscale(data: Uint8ClampedArray): void {
  for (let i = 0; i < data.length; i += 4) {
    const y = 0.299 * data[i] + 0.587 * data[i + 1] + 0.114 * data[i + 2];
    data[i] = y;
    data[i + 1] = y;
    data[i + 2] = y;
  }
}

export function createArtUpload(config: ArtUploadConfig): ArtUpload {
  return {
    dragOver: false,
    uploading: false,
    uploaded: false,
    error: false,
    errorMessage: '',
    uploadProgress: 0,
    cropping: false,
    preparing: false,
    dragging: false,
    previewUrl: '',
    imgWidth: 0,
    imgHeight: 0,
    scale: 1,
    minScale: 1,
    offsetX: 0,
    offsetY: 0,
    zoomLevel: 1,
    cropError: '',
    lastX: 0,
    lastY: 0,
    pointers: {},
    pinchStartDist: 0,
    pinchStartScale: 1,
    workingCanvas: null,
    previewObjectUrl: '',
    originalName: 'art.jpg',

    init() {
      const onKeyDown = (event: KeyboardEvent) => {
        if (event.key === 'Escape') {
          this.cancelCrop();
        }
      };
      window.addEventListener('keydown', onKeyDown);
      const alpine = this as ArtUpload & AlpineCtx;
      alpine.$cleanup?.(() => {
        window.removeEventListener('keydown', onKeyDown);
        this.teardown();
      });
    },

    teardown() {
      this.releasePreview();
      document.body.style.overflow = '';
    },

    get imageStyle() {
      return {
        width: `${this.imgWidth}px`,
        height: `${this.imgHeight}px`,
        transform: `translate(${this.offsetX}px, ${this.offsetY}px) scale(${this.scale})`,
        transformOrigin: '0 0',
        filter: 'grayscale(1)',
      };
    },

    openPicker() {
      if (this.cropping || this.uploading || this.uploaded || this.preparing) {
        return;
      }
      (this as ArtUpload & AlpineCtx).$refs.fileInput.click();
    },

    handleDrop(event: DragEvent): void {
      this.dragOver = false;
      if (this.cropping || this.uploading || this.uploaded || this.preparing) {
        return;
      }
      const files = event.dataTransfer?.files;
      if (files && files.length > 0) {
        void this.beginCrop(files[0]);
      }
    },

    handleFileSelect(event: Event): void {
      const input = event.target as HTMLInputElement;
      const files = input.files;
      if (files && files.length > 0) {
        void this.beginCrop(files[0]);
      }
      input.value = '';
    },

    async beginCrop(file: File): Promise<void> {
      this.error = false;
      this.errorMessage = '';
      this.preparing = true;

      if (!file.type.startsWith('image/')) {
        this.preparing = false;
        this.showError('Please select an image file');
        return;
      }

      if (file.size > MAX_SOURCE_BYTES) {
        this.preparing = false;
        this.showError('File size must be less than 30MB');
        return;
      }

      try {
        const decoded = await decodeImage(file);
        const canvas = rasterizeToCanvas(decoded);
        this.releasePreview();
        this.workingCanvas = canvas;
        this.imgWidth = canvas.width;
        this.imgHeight = canvas.height;
        this.originalName = file.name || 'art.jpg';
        this.cropError = '';
        const blob = await new Promise<Blob>((resolve, reject) => {
          canvas.toBlob(
            (result) => (result ? resolve(result) : reject(new Error('Failed to process image'))),
            'image/jpeg',
            0.85,
          );
        });
        this.previewObjectUrl = URL.createObjectURL(blob);
        this.previewUrl = this.previewObjectUrl;
        this.cropping = true;
        this.preparing = false;
        document.body.style.overflow = 'hidden';
        const alpine = this as ArtUpload & AlpineCtx;
        await alpine.$nextTick();
        await new Promise<void>((resolve) => {
          requestAnimationFrame(() => requestAnimationFrame(() => resolve()));
        });
        this.fitToCover();
      } catch (err) {
        this.preparing = false;
        this.showError(err instanceof Error ? err.message : 'Could not open this image');
      }
    },

    viewportSize(): number {
      const el = (this as ArtUpload & AlpineCtx).$refs.viewport;
      if (el && el.clientWidth > 0) {
        return el.clientWidth;
      }
      return Math.min(420, window.innerWidth - 48);
    },

    fitToCover(): void {
      if (!this.imgWidth || !this.imgHeight) {
        return;
      }
      const size = this.viewportSize();
      this.minScale = size / Math.min(this.imgWidth, this.imgHeight);
      this.scale = this.minScale;
      this.zoomLevel = 1;
      this.offsetX = (size - this.imgWidth * this.scale) / 2;
      this.offsetY = (size - this.imgHeight * this.scale) / 2;
      this.clampOffsets();
    },

    onResize(): void {
      if (!this.cropping) {
        return;
      }
      const previousMin = this.minScale;
      const previousZoom = previousMin > 0 ? this.scale / previousMin : 1;
      const size = this.viewportSize();
      this.minScale = size / Math.min(this.imgWidth, this.imgHeight);
      this.setScale(this.minScale * previousZoom, size / 2, size / 2);
    },

    clampOffsets(): void {
      const size = this.viewportSize();
      const displayedWidth = this.imgWidth * this.scale;
      const displayedHeight = this.imgHeight * this.scale;
      this.offsetX = clamp(this.offsetX, size - displayedWidth, 0);
      this.offsetY = clamp(this.offsetY, size - displayedHeight, 0);
    },

    setScale(nextScale: number, originX?: number, originY?: number): void {
      const size = this.viewportSize();
      const cx = originX ?? size / 2;
      const cy = originY ?? size / 2;
      const clamped = clamp(nextScale, this.minScale, this.minScale * MAX_ZOOM);
      const imgX = (cx - this.offsetX) / this.scale;
      const imgY = (cy - this.offsetY) / this.scale;
      this.scale = clamped;
      this.offsetX = cx - imgX * this.scale;
      this.offsetY = cy - imgY * this.scale;
      this.zoomLevel = this.scale / this.minScale;
      this.clampOffsets();
    },

    onZoomInput(event: Event): void {
      const zoom = Number((event.target as HTMLInputElement).value);
      this.setScale(this.minScale * zoom);
    },

    onWheel(event: WheelEvent): void {
      const rect = (this as ArtUpload & AlpineCtx).$refs.viewport.getBoundingClientRect();
      const originX = event.clientX - rect.left;
      const originY = event.clientY - rect.top;
      const factor = Math.exp(-event.deltaY * 0.0015);
      this.setScale(this.scale * factor, originX, originY);
    },

    pointerCount(): number {
      return Object.keys(this.pointers).length;
    },

    onPointerDown(event: PointerEvent): void {
      event.preventDefault();
      (this as ArtUpload & AlpineCtx).$refs.viewport.setPointerCapture(event.pointerId);
      this.pointers[String(event.pointerId)] = { x: event.clientX, y: event.clientY };

      if (this.pointerCount() === 1) {
        this.dragging = true;
        this.lastX = event.clientX;
        this.lastY = event.clientY;
      } else if (this.pointerCount() === 2) {
        this.dragging = false;
        const pts = Object.values(this.pointers) as Point[];
        this.pinchStartDist = distance(pts[0], pts[1]);
        this.pinchStartScale = this.scale;
      }
    },

    onPointerMove(event: PointerEvent): void {
      const key = String(event.pointerId);
      if (!this.pointers[key]) {
        return;
      }
      this.pointers[key] = { x: event.clientX, y: event.clientY };

      if (this.pointerCount() === 2) {
        const pts = Object.values(this.pointers) as Point[];
        const dist = distance(pts[0], pts[1]);
        if (this.pinchStartDist > 0) {
          const rect = (this as ArtUpload & AlpineCtx).$refs.viewport.getBoundingClientRect();
          const midX = (pts[0].x + pts[1].x) / 2 - rect.left;
          const midY = (pts[0].y + pts[1].y) / 2 - rect.top;
          this.setScale(this.pinchStartScale * (dist / this.pinchStartDist), midX, midY);
        }
        return;
      }

      if (!this.dragging) {
        return;
      }

      this.offsetX += event.clientX - this.lastX;
      this.offsetY += event.clientY - this.lastY;
      this.lastX = event.clientX;
      this.lastY = event.clientY;
      this.clampOffsets();
    },

    onPointerUp(event: PointerEvent): void {
      delete this.pointers[String(event.pointerId)];
      if (this.pointerCount() < 2) {
        this.pinchStartDist = 0;
      }
      if (this.pointerCount() === 0) {
        this.dragging = false;
      } else if (this.pointerCount() === 1) {
        const remaining = (Object.values(this.pointers) as Point[])[0];
        this.dragging = true;
        this.lastX = remaining.x;
        this.lastY = remaining.y;
      }
    },

    cancelCrop(): void {
      if (this.uploading || !this.cropping) {
        return;
      }
      this.cropping = false;
      this.dragging = false;
      this.cropError = '';
      this.pointers = {};
      this.releasePreview();
      document.body.style.overflow = '';
    },

    releasePreview(): void {
      if (this.previewObjectUrl) {
        URL.revokeObjectURL(this.previewObjectUrl);
      }
      this.previewObjectUrl = '';
      this.previewUrl = '';
      this.workingCanvas = null;
    },

    async confirmCrop(): Promise<void> {
      if (this.uploading || !this.workingCanvas) {
        return;
      }

      this.uploading = true;
      this.error = false;
      this.cropError = '';
      this.uploadProgress = 0;

      try {
        const file = await this.exportSquare();
        await this.uploadFile(file);
        this.cropping = false;
        document.body.style.overflow = '';
        this.releasePreview();
      } catch (err) {
        this.uploading = false;
        this.cropError = err instanceof Error ? err.message : 'Upload failed';
      }
    },

    async exportSquare(): Promise<File> {
      if (!this.workingCanvas) {
        throw new Error('No image to crop');
      }

      const size = this.viewportSize();
      let sourceSize = size / this.scale;
      let sourceX = -this.offsetX / this.scale;
      let sourceY = -this.offsetY / this.scale;

      sourceX = clamp(sourceX, 0, Math.max(0, this.imgWidth - sourceSize));
      sourceY = clamp(sourceY, 0, Math.max(0, this.imgHeight - sourceSize));
      sourceSize = Math.min(sourceSize, this.imgWidth, this.imgHeight);

      const canvas = document.createElement('canvas');
      canvas.width = OUTPUT_SIZE;
      canvas.height = OUTPUT_SIZE;
      const ctx = canvas.getContext('2d', { alpha: false });
      if (!ctx) {
        throw new Error('Failed to crop image');
      }
      ctx.imageSmoothingEnabled = true;
      ctx.imageSmoothingQuality = 'high';
      ctx.fillStyle = '#ffffff';
      ctx.fillRect(0, 0, OUTPUT_SIZE, OUTPUT_SIZE);
      ctx.save();
      ctx.beginPath();
      ctx.arc(OUTPUT_SIZE / 2, OUTPUT_SIZE / 2, OUTPUT_SIZE / 2, 0, Math.PI * 2);
      ctx.closePath();
      ctx.clip();
      ctx.filter = 'grayscale(1)';
      ctx.drawImage(
        this.workingCanvas,
        sourceX,
        sourceY,
        sourceSize,
        sourceSize,
        0,
        0,
        OUTPUT_SIZE,
        OUTPUT_SIZE,
      );
      ctx.restore();

      const pixels = ctx.getImageData(0, 0, OUTPUT_SIZE, OUTPUT_SIZE);
      applyGrayscale(pixels.data);
      ctx.putImageData(pixels, 0, 0);

      const blob = await new Promise<Blob>((resolve, reject) => {
        canvas.toBlob(
          (result) => (result ? resolve(result) : reject(new Error('Failed to crop image'))),
          'image/jpeg',
          JPEG_QUALITY,
        );
      });

      const base = this.originalName.replace(/\.[^.]+$/, '') || 'art';
      return new File([blob], `${base}.jpg`, { type: 'image/jpeg' });
    },

    async uploadFile(file: File): Promise<void> {
      this.error = false;
      this.errorMessage = '';

      if (!file.type.startsWith('image/')) {
        throw new Error('Please select an image file');
      }

      if (file.size > 10 * 1024 * 1024) {
        throw new Error('File size must be less than 10MB');
      }

      this.uploading = true;
      this.uploadProgress = 0;

      try {
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
        throw new Error(uploadErrorMessage(error));
      }
    },

    showError(message: string): void {
      this.error = true;
      this.errorMessage = message;
      this.uploading = false;
      this.uploaded = false;
      this.cropping = false;
      document.body.style.overflow = '';

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
      this.cropError = '';
      this.releasePreview();
    },

    refreshPage(): void {
      window.location.reload();
    },
  };
}

(window as any).createArtUpload = createArtUpload;

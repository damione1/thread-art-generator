import { removeBackground } from '@imgly/background-removal';

export async function stripBackground(
  source: Blob,
  onProgress?: (pct: number) => void,
): Promise<Blob> {
  return removeBackground(source, {
    model: 'isnet_quint8',
    output: { format: 'image/png', quality: 0.9 },
    progress: (_key, current, total) => {
      if (onProgress && total > 0) {
        onProgress(Math.round((current / total) * 100));
      }
    },
  });
}

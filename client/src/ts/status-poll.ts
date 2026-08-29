import { ArtStatus, CompositionStatus } from '../gen/art_pb';
import { createArtClient } from './rpc';

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export async function pollArtUntilReady(name: string): Promise<void> {
  const client = createArtClient();
  for (;;) {
    const art = await client.getArt({ name });
    if (art.status === ArtStatus.COMPLETE && art.imageUrl) {
      window.location.reload();
      return;
    }
    if (art.status === ArtStatus.FAILED) {
      window.location.reload();
      return;
    }
    await sleep(3000);
  }
}

export async function pollCompositionUntilDone(name: string): Promise<void> {
  const client = createArtClient();
  for (;;) {
    const composition = await client.getComposition({ name });
    if (
      composition.status === CompositionStatus.COMPLETE ||
      composition.status === CompositionStatus.FAILED
    ) {
      window.location.reload();
      return;
    }
    await sleep(2000);
  }
}

document.addEventListener('DOMContentLoaded', () => {
  const artEl = document.querySelector('[data-poll-art]');
  const artName = artEl?.getAttribute('data-poll-art');
  if (artName) {
    void pollArtUntilReady(artName);
  }

  const compositionEl = document.querySelector('[data-poll-composition]');
  const compositionName = compositionEl?.getAttribute('data-poll-composition');
  if (compositionName) {
    void pollCompositionUntilDone(compositionName);
  }
});

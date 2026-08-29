import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { ArtGeneratorService } from '../gen/services_pb';

export const rpcTransport = createConnectTransport({
  baseUrl: '/rpc',
  useBinaryFormat: true,
  fetch: (input, init) => {
    const headers = new Headers(init?.headers);
    const csrf = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content');
    if (csrf) {
      headers.set('X-CSRF-Token', csrf);
    }
    return fetch(input, { ...init, credentials: 'include', headers });
  },
});

export function createArtClient() {
  return createClient(ArtGeneratorService, rpcTransport);
}

import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { ArtGeneratorService } from '../gen/services_pb';

export const rpcTransport = createConnectTransport({
  baseUrl: '/rpc',
  useBinaryFormat: true,
  fetch: (input, init) => fetch(input, { ...init, credentials: 'include' }),
});

export function createArtClient() {
  return createClient(ArtGeneratorService, rpcTransport);
}

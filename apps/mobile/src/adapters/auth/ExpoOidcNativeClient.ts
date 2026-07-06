import * as AuthSession from 'expo-auth-session';
import * as WebBrowser from 'expo-web-browser';
import {
  ExpoOidcFacade,
  ExpoOidcNativeClientCore,
  NativeAccessTokenRequestConfig,
  NativeAuthRequestConfig,
  NativeDiscoveryDocument,
  NativeRefreshTokenRequestConfig
} from './ExpoOidcNativeClientCore';

const expoOidcFacade: ExpoOidcFacade = {
  completeAuthSession: () => {
    WebBrowser.maybeCompleteAuthSession();
  },
  fetchDiscovery: (issuer) => AuthSession.fetchDiscoveryAsync(issuer),
  createAuthRequest: (config) =>
    new AuthSession.AuthRequest({
      ...config,
      responseType: AuthSession.ResponseType.Code,
      scopes: [...config.scopes]
    }),
  exchangeCode: (config, discovery) =>
    AuthSession.exchangeCodeAsync(
      {
        ...config,
        scopes: [...config.scopes]
      } satisfies AuthSession.AccessTokenRequestConfig,
      discovery as AuthSession.DiscoveryDocument
    ),
  refresh: (config, discovery) =>
    AuthSession.refreshAsync(
      {
        ...config,
        scopes: [...config.scopes]
      } satisfies AuthSession.RefreshTokenRequestConfig,
      discovery as AuthSession.DiscoveryDocument
    )
};

export class ExpoOidcNativeClient extends ExpoOidcNativeClientCore {
  constructor(facade: ExpoOidcFacade = expoOidcFacade) {
    super(facade);
  }
}

export type {
  ExpoOidcFacade,
  NativeAccessTokenRequestConfig,
  NativeAuthRequestConfig,
  NativeDiscoveryDocument,
  NativeRefreshTokenRequestConfig
};

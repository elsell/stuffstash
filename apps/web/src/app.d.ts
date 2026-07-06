declare global {
  namespace App {}

  interface ImportMetaEnv {
    readonly VITE_STUFF_STASH_WEB_ORIGIN?: string;
    readonly VITE_STUFF_STASH_API_BASE_URL?: string;
    readonly VITE_STUFF_STASH_OIDC_ISSUER?: string;
    readonly VITE_STUFF_STASH_OIDC_CLIENT_ID?: string;
    readonly VITE_STUFF_STASH_OIDC_REDIRECT_URI?: string;
  }

  interface ImportMeta {
    readonly env: ImportMetaEnv;
  }
}

export {};

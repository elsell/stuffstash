module.exports = {
  expo: {
    name: 'Stuff Stash',
    slug: 'stuff-stash',
    scheme: 'stuffstash',
    version: '0.0.0',
    orientation: 'portrait',
    icon: './assets/brand/stuff-stash-glyph.png',
    userInterfaceStyle: 'automatic',
    plugins: ['expo-router'],
    ios: {
      supportsTablet: true,
      bundleIdentifier: 'app.stuffstash.mobile'
    },
    android: {
      package: 'app.stuffstash.mobile',
      adaptiveIcon: {
        foregroundImage: './assets/brand/stuff-stash-glyph.png',
        backgroundColor: '#F7FAFB'
      }
    },
    splash: {
      image: './assets/brand/stuff-stash-glyph.png',
      resizeMode: 'contain',
      backgroundColor: '#F7FAFB'
    },
    extra: {
      stuffStash: {
        apiBaseUrl: process.env.EXPO_PUBLIC_STUFF_STASH_API_BASE_URL ?? '',
        tenantId: process.env.EXPO_PUBLIC_STUFF_STASH_TENANT_ID ?? '',
        devToken: process.env.EXPO_PUBLIC_STUFF_STASH_DEV_TOKEN ?? ''
      }
    }
  }
};

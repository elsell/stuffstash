module.exports = {
  expo: {
    name: 'Stuff Stash',
    slug: 'stuff-stash',
    scheme: 'stuffstash',
    version: '0.0.0',
    orientation: 'portrait',
    icon: './assets/brand/stuff-stash-glyph.png',
    userInterfaceStyle: 'automatic',
    plugins: [
      'expo-router',
      'expo-secure-store',
      'expo-web-browser',
      [
        'expo-image-picker',
        {
          photosPermission: 'Stuff Stash uses your photo library so you can attach household item photos.',
          cameraPermission: 'Stuff Stash uses your camera so you can attach household item photos.',
          microphonePermission: 'Stuff Stash uses the microphone when you start a voice interaction.'
        }
      ]
    ],
    ios: {
      supportsTablet: true,
      bundleIdentifier: 'app.stuffstash.mobile',
      splash: {
        image: './assets/brand/stuff-stash-glyph.png',
        resizeMode: 'contain',
        backgroundColor: '#F7FAFB',
        dark: {
          image: './assets/brand/stuff-stash-glyph.png',
          resizeMode: 'contain',
          backgroundColor: '#111416'
        }
      },
      infoPlist: {
        NSPhotoLibraryUsageDescription:
          'Stuff Stash uses your photo library so you can attach household item photos.',
        NSCameraUsageDescription: 'Stuff Stash uses your camera so you can attach household item photos.',
        NSMicrophoneUsageDescription:
          'Stuff Stash uses the microphone when you start a voice interaction.'
      }
    },
    android: {
      package: 'app.stuffstash.mobile',
      splash: {
        image: './assets/brand/stuff-stash-glyph.png',
        resizeMode: 'contain',
        backgroundColor: '#F7FAFB',
        dark: {
          image: './assets/brand/stuff-stash-glyph.png',
          resizeMode: 'contain',
          backgroundColor: '#111416'
        }
      },
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
        voiceDeveloperDiagnosticsEnabled:
          process.env.EXPO_PUBLIC_STUFF_STASH_VOICE_DIAGNOSTICS_ENABLED ?? '',
        directUploadLocalDevelopmentTargetsEnabled:
          process.env.EXPO_PUBLIC_STUFF_STASH_DIRECT_UPLOAD_LOCAL_TARGETS_ENABLED ?? ''
      }
    }
  }
};

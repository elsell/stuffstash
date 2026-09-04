const invitationAllowInsecureLocalHTTP = configuredBoolean(
  'EXPO_PUBLIC_STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP',
  process.env.EXPO_PUBLIC_STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP
);
const invitationOrigin = configuredInvitationOrigin(
  process.env.EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN,
  invitationAllowInsecureLocalHTTP
);
const productionBuild = configuredBoolean(
  'STUFF_STASH_MOBILE_PRODUCTION_BUILD',
  process.env.STUFF_STASH_MOBILE_PRODUCTION_BUILD
);
const generalDistributionBuild = configuredBoolean(
  'STUFF_STASH_MOBILE_GENERAL_DISTRIBUTION_BUILD',
  process.env.STUFF_STASH_MOBILE_GENERAL_DISTRIBUTION_BUILD
);
const invitationLinksRequired = configuredBoolean(
  'STUFF_STASH_MOBILE_REQUIRE_INVITATION_LINKS',
  process.env.STUFF_STASH_MOBILE_REQUIRE_INVITATION_LINKS
);
if (invitationLinksRequired && !invitationOrigin) {
  throw new Error('EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN is required when invitation links are enabled.');
}
if ((productionBuild || invitationLinksRequired) && invitationOrigin && new URL(invitationOrigin).protocol !== 'https:') {
  throw new Error('EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN must use HTTPS for production mobile builds.');
}
const invitationURL = invitationOrigin ? new URL(invitationOrigin) : undefined;
const invitationHost = invitationURL?.protocol === 'https:' ? invitationURL.hostname : undefined;
const { resolveMobileReleaseVersion } = require('./scripts/release-version.js');
const releaseVersion = resolveMobileReleaseVersion({
  tag: process.env.STUFF_STASH_MOBILE_RELEASE_TAG,
  production: productionBuild
});
const apiBaseUrl = process.env.EXPO_PUBLIC_STUFF_STASH_API_BASE_URL?.trim() ?? '';
const tenantId = process.env.EXPO_PUBLIC_STUFF_STASH_TENANT_ID?.trim() ?? '';
const voiceDeveloperDiagnosticsEnabled = configuredBoolean(
  'EXPO_PUBLIC_STUFF_STASH_VOICE_DIAGNOSTICS_ENABLED',
  process.env.EXPO_PUBLIC_STUFF_STASH_VOICE_DIAGNOSTICS_ENABLED
);
const directUploadLocalDevelopmentTargetsEnabled = configuredBoolean(
  'EXPO_PUBLIC_STUFF_STASH_DIRECT_UPLOAD_LOCAL_TARGETS_ENABLED',
  process.env.EXPO_PUBLIC_STUFF_STASH_DIRECT_UPLOAD_LOCAL_TARGETS_ENABLED
);
if (
  generalDistributionBuild &&
  (apiBaseUrl ||
    tenantId ||
    invitationOrigin ||
    invitationAllowInsecureLocalHTTP ||
    voiceDeveloperDiagnosticsEnabled ||
    directUploadLocalDevelopmentTargetsEnabled)
) {
  throw new Error('General mobile distribution builds cannot embed deployment or developer configuration.');
}

module.exports = {
  expo: {
    name: 'Stuff Stash',
    slug: 'stuff-stash',
    scheme: 'stuffstash',
    version: releaseVersion,
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
      bundleIdentifier: 'org.stuffstash.mobile',
      icon: './ios/StuffStash/Images.xcassets/AppIcon.appiconset/App-Icon-1024x1024@1x.png',
      associatedDomains: invitationHost ? [`applinks:${invitationHost}`] : [],
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
      intentFilters: invitationHost
        ? [
            {
              action: 'VIEW',
              autoVerify: true,
              data: [{ scheme: 'https', host: invitationHost, path: '/invitations/accept' }],
              category: ['BROWSABLE', 'DEFAULT']
            }
          ]
        : [],
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
        apiBaseUrl,
        tenantId,
        voiceDeveloperDiagnosticsEnabled:
          process.env.EXPO_PUBLIC_STUFF_STASH_VOICE_DIAGNOSTICS_ENABLED ?? '',
        directUploadLocalDevelopmentTargetsEnabled:
          process.env.EXPO_PUBLIC_STUFF_STASH_DIRECT_UPLOAD_LOCAL_TARGETS_ENABLED ?? '',
        invitationOrigin: invitationOrigin ?? '',
        invitationAllowInsecureLocalHTTP:
          process.env.EXPO_PUBLIC_STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP ?? ''
      }
    }
  }
};

function configuredInvitationOrigin(value, allowInsecureLocalHTTP) {
  const trimmed = value?.trim();
  if (!trimmed) return undefined;
  try {
    const parsed = new URL(trimmed);
    if (
      (parsed.protocol !== 'https:' && !(allowInsecureLocalHTTP && isPrivateLocalHTTPOrigin(parsed))) ||
      (parsed.protocol === 'https:' && parsed.port !== '') ||
      parsed.pathname !== '/' ||
      parsed.search ||
      parsed.hash ||
      parsed.username ||
      parsed.password
    ) {
      throw new Error('invalid');
    }
    return parsed.origin;
  } catch {
    throw new Error('EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN must be a standard-port HTTPS origin.');
  }
}

function isPrivateLocalHTTPOrigin(origin) {
  if (origin.protocol !== 'http:') return false;
  if (['localhost', '127.0.0.1', '[::1]'].includes(origin.hostname)) return true;
  const parts = origin.hostname.split('.');
  if (parts.length !== 4 || parts.some((part) => !/^\d{1,3}$/.test(part))) return false;
  const octets = parts.map(Number);
  if (octets.some((octet) => octet > 255)) return false;
  return octets[0] === 10 ||
    (octets[0] === 172 && octets[1] >= 16 && octets[1] <= 31) ||
    (octets[0] === 192 && octets[1] === 168);
}

function configuredBoolean(name, value) {
  const normalized = value?.trim().toLowerCase();
  if (!normalized) return false;
  if (['1', 'true', 'yes'].includes(normalized)) return true;
  if (['0', 'false', 'no'].includes(normalized)) return false;
  throw new Error(`${name} must be a boolean.`);
}

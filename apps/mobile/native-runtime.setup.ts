import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const nativeRequire = createRequire(require.resolve('react-native/package.json'));
// Match react-native/Libraries/Core/setUpXHR.js, using its locked dependency.
const nativeAbort = nativeRequire('abort-controller/dist/abort-controller');
globalThis.AbortController = nativeAbort.AbortController;
globalThis.AbortSignal = nativeAbort.AbortSignal;

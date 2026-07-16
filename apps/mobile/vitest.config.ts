import { fileURLToPath, URL } from 'node:url';
import { defineConfig } from 'vitest/config';

const support = (file: string) => fileURLToPath(new URL(`./src/test-support/${file}`, import.meta.url));

export default defineConfig({
  resolve: {
    alias: [
      { find: /^react-native$/, replacement: support('react-native.ts') },
      { find: /^react-native-safe-area-context$/, replacement: support('react-native-safe-area-context.tsx') },
      { find: /^lucide-react-native$/, replacement: support('lucide-react-native.ts') },
      { find: /^expo-router$/, replacement: support('expo-router.ts') },
      { find: /^@react-navigation\/native$/, replacement: support('react-navigation-native.ts') },
      { find: /^@expo\/ui\/community\/segmented-control$/, replacement: support('expo-segmented-control.ts') },
      { find: /^react-native-svg$/, replacement: support('react-native-svg.ts') }
    ]
  }
});

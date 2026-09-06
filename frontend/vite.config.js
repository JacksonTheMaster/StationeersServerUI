import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte()],
  build: {
    outDir: '../SSUI/onboard_bundled/v2', // Change output directory to ../dist
    rollupOptions: {
      output: {
        // Set the name of the JS bundle
        entryFileNames: 'assets/ssui.js',
        // Set the name of the CSS bundle and other assets
        assetFileNames: (assetInfo) => {
          // Using the non-deprecated names property which is an array
          if (assetInfo.names && assetInfo.names.some(name => name.endsWith('.css'))) {
            return 'assets/ssui.css'
          }
          return 'assets/[name].[ext]'
        }
      }
    }
  },
  server: {
    host: '0.0.0.0', // Bind to all interfaces
    port: 5173,      // Match your expected port
    strictPort: true // Fail if port 5173 is taken
  }
})
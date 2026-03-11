import { defineConfig } from 'vite';

export default defineConfig({
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    target: 'esnext',
    assetsInlineLimit: 0,
    rollupOptions: {
      output: {
        entryFileNames: `[name].js`,
        chunkFileNames: `[name].js`,
        assetFileNames: `[name].[ext]`,
        manualChunks: {
          'hls': ['hls.js']
        }
      }
    }
  },
  server: {
    port: 5173
  }
});

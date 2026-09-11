// vite.config.ts - Vue 3 + Wails 适配
// Wails 期望 build 产物输出到 ./dist/，dev 由 wails.json 里的 dev:serverUrl: "auto" 自动接管
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import { fileURLToPath, URL } from 'node:url';

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  base: './', // 必须相对路径，wails 是 file:// 加载
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    target: 'es2020',
    sourcemap: false,
    rollupOptions: {
      output: {
        // Wails 期望单一入口，拆 chunks 容易触发 webview 缓存问题
        manualChunks: undefined,
      },
    },
  },
  server: {
    port: 34115,
    strictPort: true,
  },
});

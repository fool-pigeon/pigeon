import path from "node:path";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import tailwindcss from '@tailwindcss/vite';
import wails from "@wailsio/runtime/plugins/vite";
import { defineConfig } from "vite";
// https://vitejs.dev/config/
export default defineConfig({
  server: {
    host: "127.0.0.1",
    port: Number(process.env.WAILS_VITE_PORT) || 9245,
    strictPort: true,
  },
  plugins: [
    svelte(),
    wails("./bindings"),
    tailwindcss()
  ],
  resolve: {
    alias: {
      '@bindings': path.resolve('./bindings'),
      '@services': path.resolve('./bindings/github.com/fool-pigeon/pigeon/cmd/pg-gui'),
      '@lib': path.resolve('./src/lib'),
      '@components': path.resolve('./src/lib/components')
    }
  }
});

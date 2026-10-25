// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

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
      '@bindings': path.resolve('./bindings/'),
      '@pg': path.resolve('./bindings/github.com/fool-pigeon/pigeon/cmd/pg-gui/'),
      '@services': path.resolve('./bindings/github.com/fool-pigeon/pigeon/cmd/pg-gui/service'),
      '@lib': path.resolve('./src/lib'),
      '@type': path.resolve('./src/lib/types/'),
      '@comp': path.resolve('./src/lib/components')
    }
  }
});

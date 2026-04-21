import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import { copyFileSync, mkdirSync } from "node:fs";
import { resolve } from "node:path";

export default defineConfig({
  plugins: [
    react(),
    {
      name: "copy-background-service-worker",
      closeBundle() {
        const distDir = resolve(__dirname, "dist");
        mkdirSync(distDir, { recursive: true });
        copyFileSync(resolve(__dirname, "src/background.js"), resolve(distDir, "background.js"));
      }
    }
  ],
  define: {
    "process.env.NODE_ENV": JSON.stringify("production")
  },
  build: {
    outDir: resolve(__dirname, "dist"),
    emptyOutDir: true,
    sourcemap: true,
    lib: {
      entry: resolve(__dirname, "src/content/main.tsx"),
      name: "HubleadContentScript",
      formats: ["iife"],
      fileName: () => "content.js"
    },
    rollupOptions: {
      output: {
        inlineDynamicImports: true
      }
    }
  }
});

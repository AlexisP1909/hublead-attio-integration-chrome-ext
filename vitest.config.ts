import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    environment: "jsdom",
    globals: true,
    include: ["extension/src/**/*.test.ts", "extension/src/**/*.test.tsx"],
    restoreMocks: true
  }
});

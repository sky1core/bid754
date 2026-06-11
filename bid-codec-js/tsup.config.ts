import { defineConfig } from "tsup";

export default defineConfig([
  // CJS build
  {
    entry: ["src/index.ts"],
    format: ["cjs"],
    dts: false,
    outDir: "dist/cjs",
    outExtension() {
      return { js: ".cjs" };
    },
    clean: true,
    target: "es2020",
    splitting: false,
    sourcemap: true,
  },
  // ESM build + types
  {
    entry: ["src/index.ts"],
    format: ["esm"],
    dts: true,
    outDir: "dist/esm",
    outExtension() {
      return { js: ".js", dts: ".d.ts" };
    },
    clean: true,
    target: "es2020",
    splitting: false,
    sourcemap: true,
  },
  // IIFE (browser global) build
  {
    entry: { "bid-codec": "src/index.ts" },
    format: ["iife"],
    globalName: "BidCodec",
    outDir: "dist",
    outExtension() {
      return { js: ".global.js" };
    },
    clean: false,
    target: "es2020",
    splitting: false,
    sourcemap: false,
    minify: true,
  },
]);

import { cp, mkdir, rm } from "node:fs/promises";
import { resolve } from "node:path";

const root = resolve(import.meta.dirname);
const src = resolve(root, "src");
const build = resolve(root, "build");

await rm(build, { recursive: true, force: true });
await mkdir(build, { recursive: true });
await cp(src, build, { recursive: true });

console.log(`Built test extension at ${build}`);

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const projectRoot = path.resolve(__dirname, '../../');
const targetDocsDir = path.resolve(__dirname, '../src/docs');

const docMappings = [
  { src: 'operator-instructions.md', dest: 'operator-instructions.md' },
  { src: 'manager-instructions.md', dest: 'manager-instructions.md' },
  { src: 'backend/ARCHITECTURE.md', dest: 'ARCHITECTURE.md' },
  { src: 'TRAEFIK-README.md', dest: 'TRAEFIK-README.md' },
  { src: 'MONITORING-README.md', dest: 'MONITORING-README.md' },
  { src: 'AvtoplanetaApp/README.md', dest: 'mobile-README.md' },
  { src: 'frontend/README.md', dest: 'frontend-README.md' },
  { src: 'scripts/migration/README.md', dest: 'migration-README.md' },
  { src: 'README.md', dest: 'root-README.md' },
];

if (!fs.existsSync(targetDocsDir)) {
  fs.mkdirSync(targetDocsDir, { recursive: true });
}

let copiedCount = 0;
for (const item of docMappings) {
  const sourcePath = path.resolve(projectRoot, item.src);
  const targetPath = path.resolve(targetDocsDir, item.dest);

  if (fs.existsSync(sourcePath)) {
    try {
      fs.copyFileSync(sourcePath, targetPath);
      copiedCount++;
    } catch (err) {
      console.warn(`[sync-docs] Could not copy ${sourcePath}:`, err.message);
    }
  }
}

if (copiedCount > 0) {
  console.log(`[sync-docs] Synchronized ${copiedCount} documentation files to ${targetDocsDir}`);
} else {
  console.log('[sync-docs] Root files not accessible (container environment), using existing docs.');
}

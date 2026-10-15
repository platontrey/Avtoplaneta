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

// Автогенерация мобильной документации перед копированием (если доступен Dart)
const mobileToolPath = path.resolve(projectRoot, 'AvtoplanetaApp/tool/generate_docs.dart');
if (fs.existsSync(mobileToolPath)) {
  try {
    const { execSync } = await import('node:child_process');
    execSync('dart run tool/generate_docs.dart', {
      cwd: path.resolve(projectRoot, 'AvtoplanetaApp'),
      stdio: 'pipe',
    });
    console.log('[sync-docs] Generated fresh mobile-README.md via dart tool');
  } catch (err) {
    // В Docker или средах без Dart используем уже сгенерированный README.md
    console.log('[sync-docs] Note: Could not run dart generator, using existing mobile README.md');
  }
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

// Синхронизация OpenAPI спецификации swagger.json из генератора бэкенда
const swaggerSource = path.resolve(projectRoot, 'backend/gen/swagger/api.swagger.json');
const swaggerTarget = path.resolve(__dirname, '../src/data/swagger.json');

if (fs.existsSync(swaggerSource)) {
  try {
    const swaggerTargetDir = path.dirname(swaggerTarget);
    if (!fs.existsSync(swaggerTargetDir)) {
      fs.mkdirSync(swaggerTargetDir, { recursive: true });
    }
    fs.copyFileSync(swaggerSource, swaggerTarget);
    console.log(`[sync-docs] Synchronized swagger.json to ${swaggerTarget}`);
  } catch (err) {
    console.warn('[sync-docs] Could not copy swagger.json:', err.message);
  }
}

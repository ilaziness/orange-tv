#!/usr/bin/env node
/**
 * Permanently set "version" in web package.json files to match Makefile VERSION.
 * Used by `make pack` before frontend build. Does not roll back on failure.
 *
 * Usage: node scripts/sync-frontend-version.js <version>
 */
const fs = require('fs')
const path = require('path')

const version = process.argv[2]
if (!version) {
  console.error('Usage: node scripts/sync-frontend-version.js <version>')
  process.exit(1)
}

const root = path.resolve(__dirname, '..')
const files = [
  'web/package.json',
  'web/apps/admin/package.json',
  'web/apps/client/package.json',
  'web/packages/shared/package.json',
]

const versionRe = /("version"\s*:\s*")[^"]*(")/

for (const rel of files) {
  const filePath = path.join(root, rel)
  const content = fs.readFileSync(filePath, 'utf8')
  if (!versionRe.test(content)) {
    console.error(`missing "version" field in ${rel}`)
    process.exit(1)
  }
  const next = content.replace(versionRe, `$1${version}$2`)
  fs.writeFileSync(filePath, next)
  console.log(`  ${rel} -> ${version}`)
}

#!/usr/bin/env node
// Structural lint for Gherkin feature files.
//
// Formatting rules (indentation, trailing spaces, eof newline, tag spacing)
// are enforced by the `@cucumber/gherkin-utils format` + git-diff check in
// the workflow. This script covers the structural rules that the formatter
// does not, matching the previous .gherkin-lintrc:
//   - file-name (kebab-case)
//   - no-unnamed-features
//   - no-unnamed-scenarios
//   - no-dupe-scenario-names
//   - no-scenario-outlines-without-examples
//   - no-unused-variables (outline placeholders <-> Examples columns)
//
// Usage: node e2e/features/lint.mjs <glob-expanded file list...>
// Exits non-zero and prints "file: message" for every violation.

import { Parser, AstBuilder, GherkinClassicTokenMatcher } from '@cucumber/gherkin'
import { IdGenerator } from '@cucumber/messages'
import { readFileSync } from 'node:fs'
import { basename } from 'node:path'

const files = process.argv.slice(2)
if (files.length === 0) {
  console.error('usage: lint.mjs <file.feature...>')
  process.exit(2)
}

const errors = []
const err = (file, msg) => errors.push(`${file}: ${msg}`)
const kebab = /^[a-z0-9]+(-[a-z0-9]+)*\.feature$/

for (const file of files) {
  const name = basename(file)
  if (!kebab.test(name)) {
    err(file, `file-name is not kebab-case ("${name}")`)
  }

  let doc
  try {
    const parser = new Parser(new AstBuilder(IdGenerator.incrementing()), new GherkinClassicTokenMatcher())
    doc = parser.parse(readFileSync(file, 'utf8'))
  } catch (e) {
    err(file, `parse error: ${e.message}`)
    continue
  }

  const feature = doc.feature
  if (!feature) continue // empty file; format check already flags structure

  if (!feature.name || !feature.name.trim()) {
    err(file, 'feature has no name (no-unnamed-features)')
  }

  const seen = new Set()
  // Scenarios live directly under the feature or nested inside Rule blocks.
  const scenarios = feature.children.flatMap((c) =>
    c.rule ? c.rule.children.map((rc) => rc.scenario) : [c.scenario])
  for (const sc of scenarios) {
    if (!sc) continue // Background has no name requirement here

    if (!sc.name || !sc.name.trim()) {
      err(file, `${sc.keyword.trim()} at line ${sc.location.line} has no name (no-unnamed-scenarios)`)
    } else if (seen.has(sc.name)) {
      err(file, `duplicate scenario name "${sc.name}" (no-dupe-scenario-names)`)
    } else {
      seen.add(sc.name)
    }

    const isOutline = sc.keyword.trim() === 'Scenario Outline'
    if (isOutline && (!sc.examples || sc.examples.length === 0)) {
      err(file, `Scenario Outline "${sc.name}" has no Examples (no-scenario-outlines-without-examples)`)
    }

    // no-unused-variables: every <placeholder> must appear as an Examples
    // column, and every Examples column must be referenced by a <placeholder>.
    if (isOutline && sc.examples?.length) {
      const placeholders = new Set()
      const scan = (text) => {
        if (!text) return
        for (const m of text.matchAll(/<([^>]+)>/g)) placeholders.add(m[1])
      }
      for (const step of sc.steps) {
        scan(step.text)
        if (step.dataTable) {
          for (const row of step.dataTable.rows) for (const c of row.cells) scan(c.value)
        }
        if (step.docString) scan(step.docString.content)
      }
      for (const ex of sc.examples) {
        const cols = (ex.tableHeader?.cells || []).map((c) => c.value)
        for (const col of cols) {
          if (!placeholders.has(col)) {
            err(file, `Examples column "${col}" in "${sc.name}" is never used in a <placeholder> (no-unused-variables)`)
          }
        }
        for (const ph of placeholders) {
          if (!cols.includes(ph)) {
            err(file, `placeholder <${ph}> in "${sc.name}" has no matching Examples column (no-unused-variables)`)
          }
        }
      }
    }
  }
}

if (errors.length) {
  console.error(`Gherkin lint: ${errors.length} violation(s):`)
  for (const e of errors) console.error('  ' + e)
  process.exit(1)
}
console.log(`Gherkin lint: ${files.length} file(s) OK`)

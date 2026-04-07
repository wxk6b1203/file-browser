import { describe, expect, it } from 'vitest'
import en from '../locales/en.json'
import ja from '../locales/ja.json'
import zh from '../locales/zh.json'

function flattenKeys(value: unknown, prefix = ''): string[] {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    return prefix ? [prefix] : []
  }

  return Object.entries(value as Record<string, unknown>).flatMap(([key, child]) => {
    const nextPrefix = prefix ? `${prefix}.${key}` : key
    return flattenKeys(child, nextPrefix)
  })
}

describe('locale message parity', () => {
  it('keeps zh/en/ja message keys aligned', () => {
    const enKeys = flattenKeys(en).sort()

    expect(flattenKeys(zh).sort()).toEqual(enKeys)
    expect(flattenKeys(ja).sort()).toEqual(enKeys)
  })
})

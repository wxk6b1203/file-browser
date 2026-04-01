import { describe, expect, it } from 'vitest'
import IMdiCodeJson from '~icons/mdi/code-json'
import IMdiFileCogOutline from '~icons/mdi/file-cog-outline'
import IMdiFileImageOutline from '~icons/mdi/file-image-outline'
import IMdiFileOutline from '~icons/mdi/file-outline'
import IMdiFolder from '~icons/mdi/folder'
import IMdiFolderOpen from '~icons/mdi/folder-open'
import IMdiLanguageTypescript from '~icons/mdi/language-typescript'
import { DIRECTORY_ENTRY_TYPE, extractFileExtension, resolveFileIcon } from '../useFileIcons'

describe('useFileIcons', () => {
  it('returns folder icons for directory entries', () => {
    expect(resolveFileIcon({ name: 'docs', type: DIRECTORY_ENTRY_TYPE })).toBe(IMdiFolder)
    expect(resolveFileIcon({ name: 'docs', type: DIRECTORY_ENTRY_TYPE }, { opened: true })).toBe(IMdiFolderOpen)
  })

  it('maps common code extensions', () => {
    expect(resolveFileIcon({ name: 'main.ts', type: 1 })).toBe(IMdiLanguageTypescript)
    expect(resolveFileIcon('src/App.ts')).toBe(IMdiLanguageTypescript)
  })

  it('maps special config file names before extension fallback', () => {
    expect(resolveFileIcon('.env.local')).toBe(IMdiFileCogOutline)
    expect(resolveFileIcon('package.json')).toBe(IMdiCodeJson)
  })

  it('maps media extensions and falls back to generic file icon', () => {
    expect(resolveFileIcon({ name: 'cover.png', type: 1 })).toBe(IMdiFileImageOutline)
    expect(resolveFileIcon('unknown.customext')).toBe(IMdiFileOutline)
  })

  it('extracts normalized extensions', () => {
    expect(extractFileExtension('README.MD')).toBe('md')
    expect(extractFileExtension('.env')).toBe('')
  })
})

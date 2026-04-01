import { markRaw, type Component } from 'vue'
import IMdiCodeBraces from '~icons/mdi/code-braces'
import IMdiCodeJson from '~icons/mdi/code-json'
import IMdiDatabaseOutline from '~icons/mdi/database-outline'
import IMdiFileCertificateOutline from '~icons/mdi/file-certificate-outline'
import IMdiFileCodeOutline from '~icons/mdi/file-code-outline'
import IMdiFileCogOutline from '~icons/mdi/file-cog-outline'
import IMdiFileDocumentOutline from '~icons/mdi/file-document-outline'
import IMdiFileExcelBox from '~icons/mdi/file-excel-box'
import IMdiFileImageOutline from '~icons/mdi/file-image-outline'
import IMdiFileKeyOutline from '~icons/mdi/file-key-outline'
import IMdiFileOutline from '~icons/mdi/file-outline'
import IMdiFilePdfBox from '~icons/mdi/file-pdf-box'
import IMdiFilePowerpointBox from '~icons/mdi/file-powerpoint-box'
import IMdiFileTableOutline from '~icons/mdi/file-table-outline'
import IMdiFileWordBox from '~icons/mdi/file-word-box'
import IMdiFilmstripBoxMultiple from '~icons/mdi/filmstrip-box-multiple'
import IMdiFolder from '~icons/mdi/folder'
import IMdiFolderOpen from '~icons/mdi/folder-open'
import IMdiLanguageC from '~icons/mdi/language-c'
import IMdiLanguageCpp from '~icons/mdi/language-cpp'
import IMdiLanguageCsharp from '~icons/mdi/language-csharp'
import IMdiLanguageCss3 from '~icons/mdi/language-css3'
import IMdiLanguageGo from '~icons/mdi/language-go'
import IMdiLanguageHtml5 from '~icons/mdi/language-html5'
import IMdiLanguageJava from '~icons/mdi/language-java'
import IMdiLanguageJavascript from '~icons/mdi/language-javascript'
import IMdiLanguagePhp from '~icons/mdi/language-php'
import IMdiLanguagePython from '~icons/mdi/language-python'
import IMdiLanguageRuby from '~icons/mdi/language-ruby'
import IMdiLanguageRust from '~icons/mdi/language-rust'
import IMdiLanguageTypescript from '~icons/mdi/language-typescript'
import IMdiMusicNoteOutline from '~icons/mdi/music-note-outline'
import IMdiScriptTextOutline from '~icons/mdi/script-text-outline'
import IMdiVuejs from '~icons/mdi/vuejs'
import IMdiZipBoxOutline from '~icons/mdi/zip-box-outline'

export const DIRECTORY_ENTRY_TYPE = 2

export interface FileIconTarget {
  name?: string
  path?: string
  type?: number | null
}

export interface ResolveFileIconOptions {
  opened?: boolean
}

const folderIcon = markRaw(IMdiFolder) as Component
const folderOpenIcon = markRaw(IMdiFolderOpen) as Component
const defaultFileIcon = markRaw(IMdiFileOutline) as Component
const documentIcon = markRaw(IMdiFileDocumentOutline) as Component
const codeIcon = markRaw(IMdiFileCodeOutline) as Component
const imageIcon = markRaw(IMdiFileImageOutline) as Component
const pdfIcon = markRaw(IMdiFilePdfBox) as Component
const wordIcon = markRaw(IMdiFileWordBox) as Component
const excelIcon = markRaw(IMdiFileExcelBox) as Component
const powerpointIcon = markRaw(IMdiFilePowerpointBox) as Component
const archiveIcon = markRaw(IMdiZipBoxOutline) as Component
const audioIcon = markRaw(IMdiMusicNoteOutline) as Component
const videoIcon = markRaw(IMdiFilmstripBoxMultiple) as Component
const configIcon = markRaw(IMdiFileCogOutline) as Component
const scriptIcon = markRaw(IMdiScriptTextOutline) as Component
const jsonIcon = markRaw(IMdiCodeJson) as Component
const bracesIcon = markRaw(IMdiCodeBraces) as Component
const tableIcon = markRaw(IMdiFileTableOutline) as Component
const databaseIcon = markRaw(IMdiDatabaseOutline) as Component
const certificateIcon = markRaw(IMdiFileCertificateOutline) as Component
const keyIcon = markRaw(IMdiFileKeyOutline) as Component

const extensionIcons = new Map<string, Component>([
  ['ts', markRaw(IMdiLanguageTypescript) as Component],
  ['tsx', markRaw(IMdiLanguageTypescript) as Component],
  ['mts', markRaw(IMdiLanguageTypescript) as Component],
  ['cts', markRaw(IMdiLanguageTypescript) as Component],
  ['js', markRaw(IMdiLanguageJavascript) as Component],
  ['jsx', markRaw(IMdiLanguageJavascript) as Component],
  ['mjs', markRaw(IMdiLanguageJavascript) as Component],
  ['cjs', markRaw(IMdiLanguageJavascript) as Component],
  ['vue', markRaw(IMdiVuejs) as Component],
  ['go', markRaw(IMdiLanguageGo) as Component],
  ['py', markRaw(IMdiLanguagePython) as Component],
  ['java', markRaw(IMdiLanguageJava) as Component],
  ['php', markRaw(IMdiLanguagePhp) as Component],
  ['rb', markRaw(IMdiLanguageRuby) as Component],
  ['rs', markRaw(IMdiLanguageRust) as Component],
  ['c', markRaw(IMdiLanguageC) as Component],
  ['h', markRaw(IMdiLanguageC) as Component],
  ['cpp', markRaw(IMdiLanguageCpp) as Component],
  ['cc', markRaw(IMdiLanguageCpp) as Component],
  ['cxx', markRaw(IMdiLanguageCpp) as Component],
  ['hpp', markRaw(IMdiLanguageCpp) as Component],
  ['hh', markRaw(IMdiLanguageCpp) as Component],
  ['hxx', markRaw(IMdiLanguageCpp) as Component],
  ['cs', markRaw(IMdiLanguageCsharp) as Component],
  ['html', markRaw(IMdiLanguageHtml5) as Component],
  ['htm', markRaw(IMdiLanguageHtml5) as Component],
  ['css', markRaw(IMdiLanguageCss3) as Component],
  ['scss', markRaw(IMdiLanguageCss3) as Component],
  ['sass', markRaw(IMdiLanguageCss3) as Component],
  ['less', markRaw(IMdiLanguageCss3) as Component],
  ['styl', markRaw(IMdiLanguageCss3) as Component],
  ['json', jsonIcon],
  ['jsonc', jsonIcon],
  ['yaml', bracesIcon],
  ['yml', bracesIcon],
  ['toml', bracesIcon],
  ['xml', bracesIcon],
  ['ini', configIcon],
  ['conf', configIcon],
  ['cfg', configIcon],
  ['cnf', configIcon],
  ['properties', configIcon],
  ['prop', configIcon],
  ['md', documentIcon],
  ['markdown', documentIcon],
  ['txt', documentIcon],
  ['log', documentIcon],
  ['rtf', documentIcon],
  ['pdf', pdfIcon],
  ['doc', wordIcon],
  ['docx', wordIcon],
  ['odt', wordIcon],
  ['xls', excelIcon],
  ['xlsx', excelIcon],
  ['ods', excelIcon],
  ['csv', tableIcon],
  ['tsv', tableIcon],
  ['ppt', powerpointIcon],
  ['pptx', powerpointIcon],
  ['odp', powerpointIcon],
  ['png', imageIcon],
  ['jpg', imageIcon],
  ['jpeg', imageIcon],
  ['gif', imageIcon],
  ['webp', imageIcon],
  ['bmp', imageIcon],
  ['svg', imageIcon],
  ['ico', imageIcon],
  ['tif', imageIcon],
  ['tiff', imageIcon],
  ['avif', imageIcon],
  ['heic', imageIcon],
  ['zip', archiveIcon],
  ['rar', archiveIcon],
  ['7z', archiveIcon],
  ['tar', archiveIcon],
  ['gz', archiveIcon],
  ['tgz', archiveIcon],
  ['bz2', archiveIcon],
  ['xz', archiveIcon],
  ['mp3', audioIcon],
  ['wav', audioIcon],
  ['flac', audioIcon],
  ['aac', audioIcon],
  ['ogg', audioIcon],
  ['m4a', audioIcon],
  ['mp4', videoIcon],
  ['mkv', videoIcon],
  ['mov', videoIcon],
  ['avi', videoIcon],
  ['webm', videoIcon],
  ['m4v', videoIcon],
  ['flv', videoIcon],
  ['sh', scriptIcon],
  ['bash', scriptIcon],
  ['zsh', scriptIcon],
  ['ps1', scriptIcon],
  ['bat', scriptIcon],
  ['cmd', scriptIcon],
  ['sql', databaseIcon],
  ['sqlite', databaseIcon],
  ['sqlite3', databaseIcon],
  ['db', databaseIcon],
  ['db3', databaseIcon],
  ['pem', certificateIcon],
  ['crt', certificateIcon],
  ['cer', certificateIcon],
  ['der', certificateIcon],
  ['p12', certificateIcon],
  ['pfx', certificateIcon],
  ['key', keyIcon],
  ['pub', keyIcon],
  ['lock', codeIcon],
])

const exactNameIcons = new Map<string, Component>([
  ['dockerfile', configIcon],
  ['makefile', codeIcon],
  ['license', documentIcon],
  ['readme', documentIcon],
  ['readme.md', documentIcon],
  ['package.json', jsonIcon],
  ['package-lock.json', jsonIcon],
  ['pnpm-lock.yaml', configIcon],
  ['yarn.lock', codeIcon],
  ['go.mod', markRaw(IMdiLanguageGo) as Component],
  ['go.sum', markRaw(IMdiLanguageGo) as Component],
  ['cargo.toml', markRaw(IMdiLanguageRust) as Component],
  ['cargo.lock', markRaw(IMdiLanguageRust) as Component],
  ['composer.json', jsonIcon],
  ['composer.lock', codeIcon],
])

function normalizeName(name: string) {
  return name.trim().toLowerCase()
}

function extractName(target?: FileIconTarget | string | null) {
  if (!target) return ''
  if (typeof target === 'string') return normalizeName(target)
  return normalizeName(target.name || target.path || '')
}

export function extractFileExtension(target?: FileIconTarget | string | null) {
  const name = extractName(target)
  const lastDot = name.lastIndexOf('.')
  if (lastDot <= 0 || lastDot === name.length - 1) {
    return ''
  }
  return name.slice(lastDot + 1)
}

function resolveSpecialNameIcon(name: string) {
  if (!name) return null
  if (name === '.env' || name.startsWith('.env.')) return configIcon
  if (name === '.gitignore' || name === '.gitattributes' || name === '.editorconfig') return configIcon
  if (name === 'docker-compose.yml' || name === 'docker-compose.yaml') return configIcon
  if (name === 'compose.yml' || name === 'compose.yaml') return configIcon
  return exactNameIcons.get(name) ?? null
}

export function resolveFileIcon(target?: FileIconTarget | string | null, options: ResolveFileIconOptions = {}) {
  if (target && typeof target !== 'string' && target.type === DIRECTORY_ENTRY_TYPE) {
    return options.opened ? folderOpenIcon : folderIcon
  }

  const name = extractName(target)
  const specialIcon = resolveSpecialNameIcon(name)
  if (specialIcon) {
    return specialIcon
  }

  const extension = extractFileExtension(name)
  return extensionIcons.get(extension) ?? defaultFileIcon
}

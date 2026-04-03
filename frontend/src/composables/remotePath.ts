export function normalizeRemotePath(value: string | null | undefined): string {
  const trimmed = typeof value === 'string' ? value.trim() : ''
  if (!trimmed || trimmed === '.' || trimmed === '/') {
    return ''
  }

  const cleaned = trimmed
    .replace(/\\/g, '/')
    .split('/')
    .filter(Boolean)
    .join('/')

  return cleaned === '.' ? '' : cleaned
}

export function joinRemotePath(parentDir: string, name: string): string {
  const parent = normalizeRemotePath(parentDir)
  const child = normalizeRemotePath(name)
  if (!parent) return child
  if (!child) return parent
  return `${parent}/${child}`
}

export function parentRemotePath(value: string): string {
  const cleanPath = normalizeRemotePath(value)
  if (!cleanPath) return ''
  const index = cleanPath.lastIndexOf('/')
  return index === -1 ? '' : cleanPath.slice(0, index)
}

export function baseRemotePath(value: string): string {
  const cleanPath = normalizeRemotePath(value)
  if (!cleanPath) return ''
  const index = cleanPath.lastIndexOf('/')
  return index === -1 ? cleanPath : cleanPath.slice(index + 1)
}

export function isRemotePathWithin(candidate: string, parent: string): boolean {
  const cleanCandidate = normalizeRemotePath(candidate)
  const cleanParent = normalizeRemotePath(parent)
  if (!cleanCandidate || !cleanParent) return false
  return cleanCandidate === cleanParent || cleanCandidate.startsWith(`${cleanParent}/`)
}

export function remotePathKey(connectionId: string, path = ''): string {
  return `${connectionId.trim()}::${normalizeRemotePath(path)}`
}

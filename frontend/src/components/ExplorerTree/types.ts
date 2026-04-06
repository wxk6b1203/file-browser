import type { folder } from '../../../wailsjs/go/models'

export interface ExplorerNode {
  key: string
  kind: 'connection' | 'directory' | 'file'
  connectionId: string
  label: string
  path: string
  level: number
  driver?: string
  connected?: boolean
  entry?: folder.FileInfo
  children: ExplorerNode[]
}

export interface ExplorerNode {
  key: string
  kind: 'connection' | 'directory'
  connectionId: string
  label: string
  path: string
  level: number
  driver?: string
  connected?: boolean
  children: ExplorerNode[]
}

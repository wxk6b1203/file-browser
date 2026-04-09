export const CONNECTION_DIRECTORY_REFRESH_EVENT = 'workspace:connection-directory-refresh'

export interface ConnectionDirectoryRefreshDetail {
  connectionId: string
  path: string
  source: 'transfer' | 'mutation'
  taskId: string
  origin?: string
  mutation?: 'create' | 'delete' | 'rename'
  paths?: string[]
}

export function emitConnectionDirectoryRefresh(detail: ConnectionDirectoryRefreshDetail) {
  window.dispatchEvent(new CustomEvent<ConnectionDirectoryRefreshDetail>(CONNECTION_DIRECTORY_REFRESH_EVENT, {
    detail,
  }))
}

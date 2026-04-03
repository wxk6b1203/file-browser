export const CONNECTION_CONFIG_REFRESH_EVENT = 'workspace:connection-config-refresh'

export interface ConnectionConfigRefreshDetail {
  connectionId: string
  resetToRoot: boolean
}

export function emitConnectionConfigRefresh(detail: ConnectionConfigRefreshDetail) {
  window.dispatchEvent(new CustomEvent<ConnectionConfigRefreshDetail>(CONNECTION_CONFIG_REFRESH_EVENT, {
    detail,
  }))
}

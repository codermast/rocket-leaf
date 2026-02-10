const CONNECTIONS_CHANGED_EVENT = 'rocket-leaf:connections-changed'

export const emitConnectionsChanged = () => {
  if (typeof window === 'undefined') return
  window.dispatchEvent(new CustomEvent(CONNECTIONS_CHANGED_EVENT))
}

export const onConnectionsChanged = (listener: () => void): (() => void) => {
  if (typeof window === 'undefined') {
    return () => {}
  }

  const handler = () => listener()
  window.addEventListener(CONNECTIONS_CHANGED_EVENT, handler)

  return () => {
    window.removeEventListener(CONNECTIONS_CHANGED_EVENT, handler)
  }
}

import { useCallback, useEffect, useState } from 'react'
import { fetchDemo } from './api'

type State =
  | { status: 'loading'; values: null; error: null }
  | { status: 'ready'; values: number[]; error: null }
  | { status: 'error'; values: null; error: string }

const LOADING: State = { status: 'loading', values: null, error: null }

export function useDemoData() {
  const [state, setState] = useState<State>(LOADING)
  const [reloadToken, setReloadToken] = useState(0)

  useEffect(() => {
    const controller = new AbortController()

    fetchDemo(controller.signal)
      .then((data) => setState({ status: 'ready', values: data.values ?? [], error: null }))
      .catch((error: unknown) => {
        if (controller.signal.aborted) return
        setState({
          status: 'error',
          values: null,
          error: error instanceof Error ? error.message : 'Unknown error',
        })
      })

    return () => controller.abort()
  }, [reloadToken])

  // Reset to loading here rather than in the effect body, so a refetch does not
  // trigger a cascading render.
  const reload = useCallback(() => {
    setState(LOADING)
    setReloadToken((n) => n + 1)
  }, [])

  return { ...state, reload }
}

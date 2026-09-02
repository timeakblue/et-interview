/**
 * Backend client. Paths are relative so the Vite dev proxy (see vite.config.ts)
 * forwards `/_/*` and `/api/*` to the Go server on :8080.
 */

export class ApiError extends Error {
  readonly status?: number

  constructor(message: string, status?: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

async function getJSON<T>(path: string, signal?: AbortSignal): Promise<T> {
  let response: Response
  try {
    response = await fetch(path, { signal, headers: { Accept: 'application/json' } })
  } catch (cause) {
    if (cause instanceof DOMException && cause.name === 'AbortError') throw cause
    throw new ApiError(`Could not reach the backend at ${path}`)
  }

  if (!response.ok) {
    throw new ApiError(`${path} responded ${response.status} ${response.statusText}`, response.status)
  }

  return (await response.json()) as T
}

export interface DemoResponse {
  values: number[]
}

export function fetchDemo(signal?: AbortSignal): Promise<DemoResponse> {
  return getJSON<DemoResponse>('/api/demo', signal)
}

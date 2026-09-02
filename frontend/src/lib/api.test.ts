import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiError, fetchDemo } from './api'

function stubFetch(response: (...args: Parameters<typeof fetch>) => Promise<Response>) {
  const mock = vi.fn(response)
  vi.stubGlobal('fetch', mock)
  return mock
}

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('fetchDemo', () => {
  it('requests /api/demo and returns the values', async () => {
    const fetchMock = stubFetch(async () => jsonResponse({ values: [1, 2, 3] }))

    const result = await fetchDemo()

    expect(fetchMock).toHaveBeenCalledExactlyOnceWith('/api/demo', expect.anything())
    expect(result.values).toEqual([1, 2, 3])
  })

  it('throws an ApiError carrying the status when the backend rejects', async () => {
    stubFetch(async () => jsonResponse({ error: 'boom' }, 500))

    const error = await fetchDemo().catch((e: unknown) => e)

    expect(error).toBeInstanceOf(ApiError)
    expect((error as ApiError).status).toBe(500)
  })

  // The most common failure in development is the Go server not running, so
  // that message has to say so rather than surfacing "Failed to fetch".
  it('throws an ApiError naming the backend when it is unreachable', async () => {
    stubFetch(async () => {
      throw new TypeError('Failed to fetch')
    })

    const error = await fetchDemo().catch((e: unknown) => e)

    expect(error).toBeInstanceOf(ApiError)
    expect((error as ApiError).message).toContain('/api/demo')
  })
})

import type { UseFetchOptions } from 'nuxt/app'
import type { FetchError } from 'ofetch'

export interface ApiError {
  error: string
  message?: string
  details?: Record<string, string>
}

/**
 * Custom useFetch composable for calling custom Go API endpoints
 * Uses the configured $api instance from the plugin which handles:
 * - Authentication headers
 * - Base URL configuration
 * - Error handling
 */
export function useApi<T>(
  url: string | (() => string),
  options?: UseFetchOptions<T> & {
    silent?: boolean // If true, don't show toast errors
  }
) {
  const { $api } = useNuxtApp()
  const toast = useToast()
  const { silent = false, onResponseError, ...fetchOptions } = options || {}

  // Ensure API paths start with '/' to be absolute from root
  const normalizeUrl = (urlPath: string | (() => string)): string | (() => string) => {
    if (typeof urlPath === 'function') {
      return () => {
        const resolved = urlPath()
        return resolved.startsWith('/') ? resolved : `/${resolved}`
      }
    }
    return urlPath.startsWith('/') ? urlPath : `/${urlPath}`
  }

  return useFetch<T, FetchError<ApiError>>(normalizeUrl(url), {
    ...fetchOptions,
    $fetch: $api,
    onResponseError(context: any) {
      // Call original error handler if provided
      if (onResponseError && typeof onResponseError === 'function') {
        onResponseError(context)
      }

      // Show toast error unless silent
      if (!silent && context?.response?._data) {
        const error = context.response._data as ApiError
        toast.add({
          title: 'Error',
          description:
            error.error || error.message || 'An unexpected error occurred',
          color: 'error'
        })
      }
    }
  } as any)
}

/**
 * Custom useAsyncData composable for calling custom Go API endpoints
 * Useful when you don't need the reactive features of useFetch
 */
export function useApiAsync<T>(
  key: string,
  url: string | (() => string),
  options?: UseFetchOptions<T> & {
    silent?: boolean
  }
) {
  const { $api } = useNuxtApp()
  const { silent = false, ...fetchOptions } = options || {}

  // Ensure API paths start with '/' to be absolute from root
  const normalizeUrl = (urlPath: string | (() => string)): string => {
    const resolved = typeof urlPath === 'function' ? urlPath() : urlPath
    return resolved.startsWith('/') ? resolved : `/${resolved}`
  }

  return useAsyncData<T, FetchError<ApiError>>(key, () => {
    return $api(normalizeUrl(url), fetchOptions as any)
  })
}

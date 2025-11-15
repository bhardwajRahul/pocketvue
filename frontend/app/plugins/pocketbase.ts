import type { TypedPocketBase } from '@@/types/pocketbase'
import PocketBase from 'pocketbase'

export default defineNuxtPlugin((nuxtApp) => {
  const pocketbaseUrl = import.meta.dev ? 'http://localhost:8090' : '/'
  const pb = new PocketBase(pocketbaseUrl as string) as TypedPocketBase

  // Helper to build API URLs correctly in both dev and prod
  const buildApiUrl = (path: string) => {
    const cleanPath = path.startsWith('/') ? path.slice(1) : path
    const base = pb.baseURL.endsWith('/') ? pb.baseURL.slice(0, -1) : pb.baseURL
    return `${base}/${cleanPath}`
  }

  // Create custom $fetch instance for API calls
  // In production, use empty string so relative paths resolve from root
  // In dev, use pb.baseURL which is 'http://localhost:8090'
  const apiBaseURL = import.meta.dev ? pb.baseURL : ''
  const api = $fetch.create({
    baseURL: apiBaseURL,
    onRequest({ request, options }) {
      // Get auth token from PocketBase auth store
      const token = pb.authStore.token
      if (token) {
        // Set Authorization header
        // Headers can be Headers object or plain object, handle both cases
        if (options.headers instanceof Headers) {
          options.headers.set('Authorization', token)
        } else {
          // Ensure headers is a plain object
          const existingHeaders = (options.headers as Record<string, string>) || {}
          options.headers = {
            ...existingHeaders,
            Authorization: token
          } as typeof options.headers
        }
      }
    },
    async onResponseError({ response }) {
      // Handle 401 errors - user is not authenticated
      if (response.status === 401) {
        await nuxtApp.runWithContext(() => {
          pb.authStore.clear()
          navigateTo('/')
        })
      }
    }
  })

  nuxtApp.provide('pb', pb)
  nuxtApp.provide('buildApiUrl', buildApiUrl)
  nuxtApp.provide('api', api)
})

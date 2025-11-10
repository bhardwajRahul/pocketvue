declare module '#app' {
  interface NuxtApp {
    $api: typeof $fetch
    $pb: any
    $buildApiUrl: (path: string) => string
  }
}

declare module 'nuxt/schema' {
  interface NuxtApp {
    $api: typeof $fetch
    $pb: any
    $buildApiUrl: (path: string) => string
  }
}

export {}


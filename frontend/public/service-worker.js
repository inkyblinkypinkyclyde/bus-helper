const CACHE_NAME = 'bus-panel-v1'
const APP_SHELL = ['/', '/manifest.webmanifest', '/pwa-192.png', '/pwa-512.png']

self.addEventListener('install', (event) => {
  event.waitUntil(
    (async () => {
      const cache = await caches.open(CACHE_NAME)
      const appResponse = await fetch('/')
      const appHtml = await appResponse.clone().text()
      const assetPaths = [...appHtml.matchAll(/(?:src|href)="([^"]+\.(?:js|css))"/g)]
        .map((match) => match[1])
      await cache.put('/', appResponse)
      await cache.addAll([...APP_SHELL.slice(1), ...assetPaths])
      await self.skipWaiting()
    })(),
  )
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((cacheNames) =>
      Promise.all(
        cacheNames
          .filter((cacheName) => cacheName.startsWith('bus-panel-') && cacheName !== CACHE_NAME)
          .map((cacheName) => caches.delete(cacheName)),
      ),
    ),
  )
  event.waitUntil(self.clients.claim())
})

self.addEventListener('fetch', (event) => {
  const { request } = event
  const url = new URL(request.url)

  if (
    request.method !== 'GET' ||
    url.origin !== self.location.origin ||
    (request.mode !== 'navigate' && request.destination !== 'script' && request.destination !== 'style')
  ) {
    return
  }

  event.respondWith(
    fetch(request)
      .then((response) => {
        if (response.ok) {
          const responseCopy = response.clone()
          void caches.open(CACHE_NAME).then((cache) => cache.put(request, responseCopy))
        }
        return response
      })
      .catch(async () => {
        const cachedResponse = await caches.match(request)
        if (cachedResponse) return cachedResponse
        if (request.mode === 'navigate') {
          const appShell = await caches.match('/')
          if (appShell) return appShell
        }
        return Response.error()
      }),
  )
})
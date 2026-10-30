# bus-panel frontend

A React + TypeScript + Vite app that shows live buses on a map (Leaflet). It gets its data from the `/RouteInfo` endpoint of [bods-helper](../bods-helper/README.md).

## Getting started

Start `bods-helper` first (it listens on http://127.0.0.1:5000), then pick one:

### With Docker (no Node needed)

```sh
docker compose up
```

### With Node

```sh
npm install
npm run dev
```

Open http://localhost:5173. To connect from another device on the same network, open `http://<machine-LAN-IP>:5173` instead. The dev server proxies `/RouteInfo` to `http://127.0.0.1:5000`; set `BODS_API_URL` to use a different backend.

## Scripts

- `npm run dev` - start the dev server
- `npm run build` - type-check and build to `dist/`
- `npm run lint` - lint with Oxlint
- `npm run test:e2e` - Playwright tests (requires Chrome)

## Production build

The `Dockerfile` builds a static bundle served by nginx. Set the `VITE_API_BASE_URL` build arg to the URL of the API, since nginx does not proxy it.

## Installable app

The production build includes a web app manifest and service worker. Install Bus Panel from the browser menu on a supported device; the app shell is cached for offline startup, while live bus data and map tiles still require a network connection. Service workers require HTTPS, except on localhost.

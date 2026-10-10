import { expect, test } from '@playwright/test'

test('fills the viewport and zooms with the mouse wheel', async ({ page }) => {
  const viewport = { width: 1280, height: 720 }
  await page.setViewportSize(viewport)
  await page.goto('/')

  const map = page.locator('.leaflet-container')
  await expect(page.getByRole('main', { name: 'Bus network map' })).toBeVisible()
  await expect(map).toBeVisible()
  await expect(map.locator('.leaflet-control-attribution')).toContainText(
    'OpenStreetMap',
  )

  const bounds = await map.boundingBox()
  expect(bounds).toMatchObject({ x: 0, y: 0, ...viewport })

  const tileContainer = map.locator('.leaflet-tile-container').first()
  const initialTransform = await tileContainer.getAttribute('style')
  await page.mouse.move(viewport.width / 2, viewport.height / 2)
  await page.mouse.wheel(0, -500)

  await expect
    .poll(() => tileContainer.getAttribute('style'))
    .not.toBe(initialTransform)
})

test('accepts an operator and route over the map on mobile', async ({ page }) => {
  const viewport = { width: 390, height: 844 }
  await page.setViewportSize(viewport)
  await page.goto('/')

  const map = page.locator('.leaflet-container')
  const menu = page.getByRole('region', { name: 'Map filters' })
  const operator = page.getByRole('textbox', { name: 'Operator' })
  const route = page.getByRole('textbox', { name: 'Route number' })

  await expect(map).toBeVisible()
  await expect(menu).toBeVisible()
  await expect(operator).toHaveValue('Metroline')
  await expect(route).toHaveValue('24X')
  await operator.fill('Metroline')
  await route.fill('24X')

  await expect(operator).toHaveValue('Metroline')
  await expect(route).toHaveValue('24X')

  const mapBounds = await map.boundingBox()
  const menuBounds = await menu.boundingBox()
  expect(mapBounds).toMatchObject({ x: 0, y: 0, ...viewport })
  expect(menuBounds?.y).toBeGreaterThanOrEqual(0)
  expect(menuBounds?.y + menuBounds!.height).toBeLessThan(viewport.height)
})

test('requests a route and renders returned bus locations', async ({ page }) => {
  await page.goto('/')
  let requestUrl: URL | undefined

  await page.route('**/bus-photos/BUS-1.jpg', async (route) => {
    await route.fulfill({
      contentType: 'image/svg+xml',
      body: '<svg xmlns="http://www.w3.org/2000/svg" width="1" height="1"></svg>',
    })
  })

  await page.route('**/RouteInfo?*', async (route) => {
    requestUrl = new URL(route.request().url())
    await route.fulfill({
      contentType: 'application/json',
      body: JSON.stringify({
        operator: 'AB',
        line: '12',
        buses: [
          {
            vehicle_ref: 'BUS-1',
            line: '12',
            destination: 'Central',
            latitude: 51.5,
            longitude: -0.1,
          },
        ],
      }),
    })
  })

  await page.getByRole('textbox', { name: 'Operator' }).fill('AB')
  await page.getByRole('textbox', { name: 'Route number' }).fill('12')
  await page.getByRole('button', { name: 'show' }).click()

  const marker = page.locator('.bus-marker')
  await expect(marker).toBeVisible()
  await marker.click()
  await expect(page.getByText('BUS-1')).toBeVisible()
  await expect(page.getByRole('img', { name: 'BUS-1 bus' })).toBeVisible()
  expect(page.getByRole('img', { name: 'BUS-1 bus' })).toHaveAttribute(
    'src',
    '/bus-photos/BUS-1.jpg',
  )
  expect(requestUrl?.pathname).toBe('/RouteInfo')
  expect(requestUrl?.searchParams.get('operator')).toBe('AB')
  expect(requestUrl?.searchParams.get('line')).toBe('12')
  expect(requestUrl?.searchParams.get('lat')).toBe('54')
  expect(requestUrl?.searchParams.get('lon')).toBe('-3')
})
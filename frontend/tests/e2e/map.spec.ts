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
import { useState, type FormEvent } from 'react'
import { CircleMarker, MapContainer, Popup, TileLayer, useMap } from 'react-leaflet'
import './App.css'

type RouteBus = {
  vehicle_ref: string
  line: string
  destination: string
  latitude: number
  longitude: number
}

type RouteInfoResponse = {
  buses?: RouteBus[]
  error?: string
}

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? ''

function MapFilters() {
  const map = useMap()
  const [buses, setBuses] = useState<RouteBus[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function showBuses(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const formData = new FormData(event.currentTarget)
    const center = map.getCenter()
    const params = new URLSearchParams({
      operator: String(formData.get('operator')).trim(),
      line: String(formData.get('route')).trim(),
      lat: String(center.lat),
      lon: String(center.lng),
    })

    setLoading(true)
    setError('')

    try {
      const response = await fetch(`${apiBaseUrl}/RouteInfo?${params}`)
      const result = (await response.json()) as RouteInfoResponse
      if (!response.ok) {
        throw new Error(result.error ?? `Request failed (${response.status})`)
      }

      const locations = (result.buses ?? []).filter(
        (bus) => Number.isFinite(bus.latitude) && Number.isFinite(bus.longitude),
      )
      setBuses(locations)
      if (locations.length > 0) {
        map.fitBounds(
          locations.map((bus) => [bus.latitude, bus.longitude]),
          { padding: [48, 48], maxZoom: 12 },
        )
      }
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'Request failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <section className="map-menu" aria-label="Map filters">
        <form className="map-menu__form" onSubmit={showBuses}>
          <label className="map-menu__field">
            Operator
            <input
              name="operator"
              type="text"
              autoComplete="off"
              placeholder="Enter operator"
              required
            />
          </label>
          <label className="map-menu__field">
            Route number
            <input
              name="route"
              type="text"
              autoComplete="off"
              placeholder="Enter route number"
              required
            />
          </label>
          <button className="map-menu__submit" type="submit" disabled={loading}>
            show
          </button>
        </form>
        {error && (
          <p className="map-menu__error" role="alert">
            {error}
          </p>
        )}
      </section>
      {buses.map((bus) => (
        <CircleMarker
          key={bus.vehicle_ref}
          center={[bus.latitude, bus.longitude]}
          radius={7}
          pathOptions={{
            className: 'bus-marker',
            color: '#fff',
            fillColor: '#176b57',
            fillOpacity: 1,
            weight: 2,
          }}
        >
          <Popup>
            <strong>{bus.vehicle_ref}</strong>
            <br />
            {bus.line} to {bus.destination}
          </Popup>
        </CircleMarker>
      ))}
    </>
  )
}

function App() {
  return (
    <main className="map-shell" aria-label="Bus network map">
      <MapContainer
        className="map"
        center={[54, -3]}
        zoom={6}
        zoomControl={false}
        scrollWheelZoom
        touchZoom
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        <MapFilters />
      </MapContainer>
    </main>
  )
}

export default App

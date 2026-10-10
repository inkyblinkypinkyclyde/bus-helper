import { MapContainer, TileLayer } from 'react-leaflet'
import './App.css'

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
      </MapContainer>
      <section className="map-menu" aria-label="Map filters">
        <label className="map-menu__field">
          Operator
          <input
            name="operator"
            type="text"
            autoComplete="off"
            placeholder="Enter operator"
          />
        </label>
        <label className="map-menu__field">
          Route number
          <input
            name="route"
            type="text"
            autoComplete="off"
            placeholder="Enter route number"
          />
        </label>
      </section>
    </main>
  )
}

export default App

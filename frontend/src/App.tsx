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
    </main>
  )
}

export default App

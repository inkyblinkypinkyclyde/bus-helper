#!/usr/bin/env python3
"""Look up a bus stop by location and show nearby live bus activity from BODS.

The Bus Open Data Service (BODS) doesn't expose a single "stop details"
endpoint via bods-client, so this queries the SIRI-VM real-time vehicle
feed for a small bounding box around the given stop coordinates and
reports the buses currently operating near that stop.

Usage:
    python stop_info.py --lat 51.5074 --lon -0.1278 [--radius 250] [--name "Stop name"]
"""

import argparse
import math
import os
import sys
from dataclasses import dataclass

from dotenv import load_dotenv

from bods_client.client import BODSClient
from bods_client.models import BoundingBox, Siri, SIRIVMParams
from bearing import initial_bearing_degrees

EARTH_RADIUS_M = 6_371_000


@dataclass
class NearbyBus:
    line: str
    destination: str
    origin: str
    operator: str
    vehicle_ref: str
    distance_m: float
    bearing: float
    bearing_from_user: float | None


def bounding_box_around(lat: float, lon: float, radius_m: float) -> BoundingBox:
    """Build a BoundingBox roughly `radius_m` metres around a lat/lon point."""
    lat_delta = radius_m / 111_320  # metres per degree of latitude
    lon_delta = radius_m / (111_320 * math.cos(math.radians(lat)) or 1)
    return BoundingBox(
        min_latitude=lat - lat_delta,
        max_latitude=lat + lat_delta,
        min_longitude=lon - lon_delta,
        max_longitude=lon + lon_delta,
    )


def haversine_m(lat1: float, lon1: float, lat2: float, lon2: float) -> float:
    phi1, phi2 = math.radians(lat1), math.radians(lat2)
    d_phi = math.radians(lat2 - lat1)
    d_lambda = math.radians(lon2 - lon1)
    a = math.sin(d_phi / 2) ** 2 + math.cos(phi1) * math.cos(phi2) * math.sin(d_lambda / 2) ** 2
    return 2 * EARTH_RADIUS_M * math.asin(math.sqrt(a))


def get_nearby_buses(client: BODSClient, lat: float, lon: float, radius_m: float) -> list[NearbyBus]:
    bbox = bounding_box_around(lat, lon, radius_m)
    params = SIRIVMParams(bounding_box=bbox)
    response = client.get_siri_vm_data_feed(params=params)

    if not isinstance(response, (bytes, bytearray)):
        raise RuntimeError(f"BODS request failed: {response.status_code} {response.reason}")

    siri = Siri.from_bytes(response)
    activities = siri.service_delivery.vehicle_monitoring_delivery.vehicle_activities or []

    buses = []
    for activity in activities:
        journey = activity.monitored_vehicle_journey
        vehicle_location = journey.vehicle_location
        distance = haversine_m(lat, lon, vehicle_location.latitude, vehicle_location.longitude)
        buses.append(
            NearbyBus(
                line=journey.published_line_name or "?",
                destination=journey.destination_name or "?",
                origin=journey.origin_name or "?",
                operator=journey.operator_ref or "?",
                vehicle_ref=journey.vehicle_ref or "?",
                distance_m=distance,
                bearing=journey.bearing if journey.bearing is not None else float("nan"),
                bearing_from_user=initial_bearing_degrees(
                    lat, lon, vehicle_location.latitude, vehicle_location.longitude
                ),
            )
        )

    return sorted(buses, key=lambda bus: bus.distance_m)


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--lat", type=float, required=True, help="Stop latitude")
    parser.add_argument("--lon", type=float, required=True, help="Stop longitude")
    parser.add_argument("--radius", type=float, default=250, help="Search radius in metres (default: 250)")
    parser.add_argument("--name", default=None, help="Optional stop name to display")
    args = parser.parse_args()

    load_dotenv()
    api_key = os.environ.get("API_KEY")
    if not api_key:
        sys.exit("API_KEY not set. Add it to your .env file.")

    client = BODSClient(api_key=api_key)

    label = args.name or f"({args.lat}, {args.lon})"
    print(f"Bus stop: {label}")
    print(f"Searching within {args.radius:.0f}m...\n")

    try:
        buses = get_nearby_buses(client, args.lat, args.lon, args.radius)
    except RuntimeError as exc:
        sys.exit(str(exc))

    if not buses:
        print("No live buses found near this stop right now.")
        return

    for bus in buses:
        print(
            f"Line {bus.line:<6} -> {bus.destination:<30} "
            f"(from {bus.origin}) | operator: {bus.operator} | "
            f"vehicle: {bus.vehicle_ref} | {bus.distance_m:.0f}m away"
        )


if __name__ == "__main__":
    main()

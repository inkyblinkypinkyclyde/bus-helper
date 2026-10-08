#!/usr/bin/env python3
"""Show how far away each live bus on a given route is from a location.

Usage:
    python route_distance.py --operator TFLO --line 29 --lat 51.5074 --lon -0.1278
    python route_distance.py --operator TFLO --line 29 --lat 51.5074 --lon -0.1278 --destination "Trafalgar Square,Aldwych"
"""

import argparse
import math
import os
import sys
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone

from dotenv import load_dotenv

from bods_client.client import BODSClient
from bods_client.models import Siri, SIRIVMParams

EARTH_RADIUS_M = 6_371_000
MAX_RECORD_AGE = timedelta(minutes=10)


@dataclass
class RouteBus:
    vehicle_ref: str
    destination: str
    origin: str
    distance_m: float
    recorded_at: str


def haversine_m(lat1: float, lon1: float, lat2: float, lon2: float) -> float:
    phi1, phi2 = math.radians(lat1), math.radians(lat2)
    d_phi = math.radians(lat2 - lat1)
    d_lambda = math.radians(lon2 - lon1)
    a = math.sin(d_phi / 2) ** 2 + math.cos(phi1) * math.cos(phi2) * math.sin(d_lambda / 2) ** 2
    return 2 * EARTH_RADIUS_M * math.asin(math.sqrt(a))


def get_route_buses(
    client: BODSClient,
    operator: str,
    line: str,
    lat: float,
    lon: float,
    destinations: list[str] | None = None,
) -> list[RouteBus]:
    params = SIRIVMParams(operator_refs=[operator], line_ref=line)
    response = client.get_siri_vm_data_feed(params=params)

    if not isinstance(response, (bytes, bytearray)):
        raise RuntimeError(f"BODS request failed: {response.status_code} {response.reason}")

    siri = Siri.from_bytes(response)
    activities = siri.service_delivery.vehicle_monitoring_delivery.vehicle_activities or []

    buses = []
    cutoff = datetime.now(timezone.utc) - MAX_RECORD_AGE
    for activity in activities:
        recorded_at = activity.recorded_at_time
        if recorded_at is None:
            continue
        if recorded_at.tzinfo is None:
            recorded_at = recorded_at.replace(tzinfo=timezone.utc)
        if recorded_at <= cutoff:
            continue

        journey = activity.monitored_vehicle_journey
        destination_name = journey.destination_name or "?"
        if destinations and not any(dest.lower() in destination_name.lower() for dest in destinations):
            continue
        vehicle_location = journey.vehicle_location
        distance = haversine_m(lat, lon, vehicle_location.latitude, vehicle_location.longitude)
        buses.append(
            RouteBus(
                vehicle_ref=journey.vehicle_ref or "?",
                destination=destination_name,
                origin=journey.origin_name or "?",
                distance_m=distance,
                recorded_at=str(recorded_at),
            )
        )

    return sorted(buses, key=lambda bus: bus.distance_m)


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--operator", required=True, help="National Operator Code, e.g. TFLO")
    parser.add_argument("--line", required=True, help="Published line name/number, e.g. 29")
    parser.add_argument("--lat", type=float, required=True, help="Reference latitude")
    parser.add_argument("--lon", type=float, required=True, help="Reference longitude")
    parser.add_argument(
        "--destination",
        default=None,
        help="Comma-separated list of final destination stop names to filter to, matched as substrings "
        "(useful when an operator reuses a line number for multiple routes with alternate end stops)",
    )
    args = parser.parse_args()

    destinations = None
    if args.destination:
        destinations = [dest.strip() for dest in args.destination.split(",") if dest.strip()]

    load_dotenv()
    api_key = os.environ.get("API_KEY")
    if not api_key:
        sys.exit("API_KEY not set. Add it to your .env file.")

    client = BODSClient(api_key=api_key)

    print(f"Line {args.line} ({args.operator}) buses relative to ({args.lat}, {args.lon}):\n")

    try:
        buses = get_route_buses(client, args.operator, args.line, args.lat, args.lon, destinations)
    except RuntimeError as exc:
        sys.exit(str(exc))

    if not buses:
        print("No live buses found for this operator/line/destination right now.")
        return

    for bus in buses:
        print(
            f"Vehicle {bus.vehicle_ref:<10} -> {bus.destination:<30} "
            f"(from {bus.origin}) | {bus.distance_m:.0f}m away | recorded: {bus.recorded_at}"
        )


if __name__ == "__main__":
    main()

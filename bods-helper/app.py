import math
import os
from dataclasses import asdict

from dotenv import load_dotenv
from flask import Flask, jsonify, request

from bods_client.client import BODSClient
from route_distance import get_route_buses
from stop_info import get_nearby_buses

load_dotenv()

app = Flask(__name__)


def get_client() -> BODSClient:
    api_key = os.environ.get("API_KEY")
    if not api_key:
        raise RuntimeError("API_KEY not set. Add it to your .env file.")
    return BODSClient(api_key=api_key)


def required_float(name: str) -> float:
    value = request.args.get(name)
    if value is None:
        raise ValueError(f"Missing required parameter: {name}")
    try:
        return float(value)
    except ValueError:
        raise ValueError(f"Parameter {name} must be a number")


def bus_to_dict(bus) -> dict:
    # NaN isn't valid JSON, so strict parsers (e.g. Go) reject it
    return {k: (None if isinstance(v, float) and math.isnan(v) else v) for k, v in asdict(bus).items()}


@app.errorhandler(ValueError)
def handle_bad_request(exc):
    return jsonify(error=str(exc)), 400


@app.errorhandler(RuntimeError)
def handle_upstream_error(exc):
    return jsonify(error=str(exc)), 502


@app.get("/LocationInfo")
def location_info():
    lat = required_float("lat")
    lon = required_float("lon")
    radius = float(request.args.get("radius", 250))
    buses = get_nearby_buses(get_client(), lat, lon, radius)
    return jsonify(
        name=request.args.get("name"),
        lat=lat,
        lon=lon,
        radius_m=radius,
        buses=[bus_to_dict(bus) for bus in buses],
    )


@app.get("/RouteInfo")
def route_info():
    operator = request.args.get("operator")
    line = request.args.get("line")
    if not operator or not line:
        raise ValueError("Missing required parameters: operator, line")
    lat = required_float("lat")
    lon = required_float("lon")
    destination = request.args.get("destination")
    destinations = [d.strip() for d in destination.split(",") if d.strip()] if destination else None
    buses = get_route_buses(get_client(), operator, line, lat, lon, destinations)
    return jsonify(operator=operator, line=line, buses=[bus_to_dict(bus) for bus in buses])


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=int(os.environ.get("PORT", 5000)))

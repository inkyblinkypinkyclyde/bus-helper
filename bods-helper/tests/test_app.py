import math
import unittest
from dataclasses import dataclass
from unittest.mock import patch

import app
from bearing import initial_bearing_degrees
from route_distance import RouteBus


@dataclass
class Bus:
    line: str
    bearing: float
    bearing_from_user: float | None


class BearingTestCase(unittest.TestCase):
    def test_initial_bearing_due_east_is_90_degrees(self):
        self.assertAlmostEqual(initial_bearing_degrees(0, 0, 0, 1), 90)

    def test_initial_bearing_rounds_to_nearest_degree(self):
        self.assertEqual(initial_bearing_degrees(0, 0, 1, 2), 63)

    def test_initial_bearing_is_undefined_at_same_location(self):
        self.assertIsNone(initial_bearing_degrees(51.5, -0.1, 51.5, -0.1))


class AppTestCase(unittest.TestCase):
    def setUp(self):
        self.client = app.app.test_client()

    @patch("app.get_nearby_buses")
    @patch("app.get_client")
    def test_location_info_uses_default_radius_and_serializes_nan(self, get_client, get_nearby_buses):
        get_nearby_buses.return_value = [Bus(line="12", bearing=math.nan, bearing_from_user=90)]

        response = self.client.get("/LocationInfo?lat=51.5&lon=-0.1&name=Central")

        self.assertEqual(response.status_code, 200)
        self.assertEqual(
            response.get_json(),
            {
                "name": "Central",
                "lat": 51.5,
                "lon": -0.1,
                "radius_m": 250.0,
                "buses": [{"line": "12", "bearing": None, "bearing_from_user": 90}],
            },
        )
        get_nearby_buses.assert_called_once_with(get_client.return_value, 51.5, -0.1, 250.0)

    @patch("app.get_route_buses")
    @patch("app.get_client")
    def test_route_info_splits_and_trims_destinations(self, get_client, get_route_buses):
        get_route_buses.return_value = [
            RouteBus(
                vehicle_ref="BUS-1",
                destination="Central",
                origin="North Road",
                latitude=51.5,
                longitude=-0.1,
                distance_m=25,
                recorded_at="2026-10-10T10:00:00+00:00",
                bearing_from_user=90,
            )
        ]

        response = self.client.get(
            "/RouteInfo?operator=AB&line=12&lat=51.5&lon=-0.1&destination=North%20Road,%20Central"
        )

        self.assertEqual(response.status_code, 200)
        self.assertEqual(
            response.get_json(),
            {
                "operator": "AB",
                "line": "12",
                "buses": [
                    {
                        "vehicle_ref": "BUS-1",
                        "destination": "Central",
                        "origin": "North Road",
                        "latitude": 51.5,
                        "longitude": -0.1,
                        "distance_m": 25,
                        "recorded_at": "2026-10-10T10:00:00+00:00",
                        "bearing_from_user": 90,
                    }
                ],
            },
        )
        get_route_buses.assert_called_once_with(
            get_client.return_value, "AB", "12", 51.5, -0.1, ["North Road", "Central"]
        )

    def test_location_info_rejects_missing_coordinates(self):
        response = self.client.get("/LocationInfo?lat=51.5")

        self.assertEqual(response.status_code, 400)
        self.assertEqual(response.get_json(), {"error": "Missing required parameter: lon"})

    def test_route_info_rejects_missing_route_parameters(self):
        response = self.client.get("/RouteInfo?lat=51.5&lon=-0.1")

        self.assertEqual(response.status_code, 400)
        self.assertEqual(response.get_json(), {"error": "Missing required parameters: operator, line"})

    def test_location_info_rejects_non_numeric_coordinates(self):
        response = self.client.get("/LocationInfo?lat=north&lon=-0.1")

        self.assertEqual(response.status_code, 400)
        self.assertEqual(response.get_json(), {"error": "Parameter lat must be a number"})


if __name__ == "__main__":
    unittest.main()
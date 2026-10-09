import math
import unittest
from dataclasses import dataclass
from unittest.mock import patch

import app


@dataclass
class Bus:
    line: str
    bearing: float


class AppTestCase(unittest.TestCase):
    def setUp(self):
        self.client = app.app.test_client()

    @patch("app.get_nearby_buses")
    @patch("app.get_client")
    def test_location_info_uses_default_radius_and_serializes_nan(self, get_client, get_nearby_buses):
        get_nearby_buses.return_value = [Bus(line="12", bearing=math.nan)]

        response = self.client.get("/LocationInfo?lat=51.5&lon=-0.1&name=Central")

        self.assertEqual(response.status_code, 200)
        self.assertEqual(
            response.get_json(),
            {
                "name": "Central",
                "lat": 51.5,
                "lon": -0.1,
                "radius_m": 250.0,
                "buses": [{"line": "12", "bearing": None}],
            },
        )
        get_nearby_buses.assert_called_once_with(get_client.return_value, 51.5, -0.1, 250.0)

    @patch("app.get_route_buses")
    @patch("app.get_client")
    def test_route_info_splits_and_trims_destinations(self, get_client, get_route_buses):
        get_route_buses.return_value = []

        response = self.client.get(
            "/RouteInfo?operator=AB&line=12&lat=51.5&lon=-0.1&destination=North%20Road,%20Central"
        )

        self.assertEqual(response.status_code, 200)
        self.assertEqual(response.get_json(), {"operator": "AB", "line": "12", "buses": []})
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
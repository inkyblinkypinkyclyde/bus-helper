import math


def initial_bearing_degrees(
    from_latitude: float,
    from_longitude: float,
    to_latitude: float,
    to_longitude: float,
) -> int | None:
    if from_latitude == to_latitude and from_longitude == to_longitude:
        return None

    from_latitude_radians = math.radians(from_latitude)
    to_latitude_radians = math.radians(to_latitude)
    longitude_delta_radians = math.radians(to_longitude - from_longitude)

    east_component = math.sin(longitude_delta_radians) * math.cos(to_latitude_radians)
    north_component = (
        math.cos(from_latitude_radians) * math.sin(to_latitude_radians)
        - math.sin(from_latitude_radians)
        * math.cos(to_latitude_radians)
        * math.cos(longitude_delta_radians)
    )
    bearing = (math.degrees(math.atan2(east_component, north_component)) + 360) % 360
    return math.floor(bearing + 0.5) % 360
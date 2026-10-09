# bus-panel

Live bus info from the Bus Open Data Service (BODS).

## Setup

```sh
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

## Start the server

```sh
python app.py
```

The server listens on http://127.0.0.1:5000.

- `GET /LocationInfo?lat=51.5077&lon=-0.1297&radius=50&name=Trafalgar%20Square`
  - `lat`, `lon` required; `radius` (metres, default 250) and `name` optional.
- `GET /RouteInfo?operator=ARBB&line=4&lat=51.5074&lon=-0.1278&destination=Church_Street,Central_Railway_Station`
  - `operator`, `line`, `lat`, `lon` required; `destination` (comma-separated substrings) optional.

Each bus response includes `bearing`, the bus's heading, and `bearing_from_user`, the initial compass bearing from the requested coordinates to the bus in degrees clockwise from north. `bearing_from_user` is null when the coordinates are identical.

```sh
curl 'http://127.0.0.1:5000/LocationInfo?lat=51.5077&lon=-0.1297&radius=50'
curl 'http://127.0.0.1:5000/RouteInfo?operator=ARBB&line=4&lat=51.5074&lon=-0.1278'
```

## Run tests and build executables

Run the tests with `make test` after installing the project requirements.

`make build` installs the build requirements into the active Python environment and creates `dist/bods-helper` for the host operating system and architecture. Run the build on each target OS and architecture; PyInstaller does not cross-compile.

## Run the helpers from the command line

Stop info (buses near a location):

```sh
python stop_info.py --lat 51.50769337362035 --lon -0.12967084691419908 --radius 50 --name "Trafalgar Square"
```

Route distance (distance of each live bus on a route from a location):

```sh
python route_distance.py --operator ARBB --line 4 --lat 51.5074 --lon -0.1278
python route_distance.py --operator ARBB --line 4 --lat 52.0 --lon -0.75 --destination "Church_Street,Central_Railway_Station"
```

Repeat a helper at a fixed interval (default 10 seconds):

```sh
python repeat.py stop_info.py --lat 52.0 --lon -0.75 --radius 500
python repeat.py --interval 10 route_distance.py --operator ARBB --line 4 --lat 52.0 --lon -0.75 --destination "Church_Street,Central_Railway_Station"
```

Pipe route output into `distances_csv.sh` (needs `jq` and an Ollama server; edit the URL in the script):

```sh
python route_distance.py --operator ARBB --line 4 --lat 51.5074 --lon -0.1278 | ./distances_csv.sh
```
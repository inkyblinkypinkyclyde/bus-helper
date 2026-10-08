#!/usr/bin/env python3
"""Run a Python script repeatedly at a fixed interval.

Usage:
    python repeat.py stop_info.py --lat 52.0 --lon -0.75 --radius 500
    python repeat.py --interval 10 route_distance.py --operator ARBB --line 4 --lat 52.0 --lon -0.75
"""

import argparse
import subprocess
import sys
import time
from datetime import datetime


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--interval",
        type=float,
        default=10,
        help="Seconds between runs (default: 10)",
    )
    parser.add_argument("script", help="Python script to run repeatedly")
    parser.add_argument("script_args", nargs=argparse.REMAINDER, help="Arguments passed to the child script")
    args = parser.parse_args()

    if args.interval <= 0:
        parser.error("--interval must be greater than zero")

    command = [sys.executable, args.script, *args.script_args]
    next_run = time.monotonic()
    run_number = 0

    try:
        while True:
            delay = next_run - time.monotonic()
            if delay > 0:
                time.sleep(delay)

            run_number += 1
            print(f"\n--- Run {run_number} at {datetime.now().astimezone().isoformat(timespec='seconds')} ---", flush=True)
            try:
                result = subprocess.run(command, check=False)
            except OSError as exc:
                sys.exit(f"Could not run {args.script}: {exc}")

            if result.returncode:
                print(f"Child script exited with status {result.returncode}; continuing.", flush=True)

            next_run += args.interval
            now = time.monotonic()
            if next_run <= now:
                missed_intervals = int((now - next_run) // args.interval) + 1
                next_run += missed_intervals * args.interval
    except KeyboardInterrupt:
        print("\nStopped.", flush=True)


if __name__ == "__main__":
    main()

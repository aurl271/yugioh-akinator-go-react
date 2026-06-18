from __future__ import annotations

import argparse
import subprocess
import sys
from pathlib import Path


BASE_DIR = Path(__file__).resolve().parents[1]
SCRIPT_DIR = Path(__file__).resolve().parent
STEPS_DIR = SCRIPT_DIR / "steps"
PYTHON = sys.executable


def run_step(name: str, command: list[str]) -> None:
    print(f"\n== {name} ==", flush=True)
    subprocess.run(command, cwd=BASE_DIR.parent, check=True)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Prepare PostgreSQL import files and optionally import them.")
    parser.add_argument("--skip-import", action="store_true")
    return parser.parse_args()


def main() -> None:
    args = parse_args()

    run_step("export csv from sqlite", [PYTHON, str(STEPS_DIR / "export_csv_from_sqlite.py")])
    if not args.skip_import:
        run_step("import csv to postgres", [PYTHON, str(STEPS_DIR / "import_csv_to_postgres.py")])
    print("\ndone")


if __name__ == "__main__":
    main()

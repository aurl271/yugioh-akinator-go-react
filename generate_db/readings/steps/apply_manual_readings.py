from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any


BASE_DIR = Path(__file__).resolve().parents[2]
DEFAULT_INPUT_PATH = BASE_DIR / "source_data" / "json" / "missing_readings.json"
DEFAULT_MANUAL_PATH = BASE_DIR / "source_data" / "json" / "manual_readings.json"
DEFAULT_OUTPUT_PATH = BASE_DIR / "source_data" / "json" / "missing_readings.json"


def load_json(path: Path, default: Any) -> Any:
    if not path.exists():
        return default
    with path.open("r", encoding="utf-8") as file:
        return json.load(file)


def save_json(path: Path, data: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def main() -> None:
    parser = argparse.ArgumentParser(description="Apply manual readings to missing_readings.json.")
    parser.add_argument("--input", type=Path, default=DEFAULT_INPUT_PATH)
    parser.add_argument("--manual", type=Path, default=DEFAULT_MANUAL_PATH)
    parser.add_argument("--output", type=Path, default=DEFAULT_OUTPUT_PATH)
    args = parser.parse_args()

    missing_cards = load_json(args.input, [])
    manual_readings = load_json(args.manual, {})

    updated = 0
    for card in missing_cards:
        card_id = str(card["card_id"])
        reading = manual_readings.get(card_id)
        if reading:
            card["reading"] = reading
            updated += 1

    save_json(args.output, missing_cards)
    print(f"manual readings: {len(manual_readings)}")
    print(f"updated missing readings: {updated}")
    print(f"output: {args.output}")


if __name__ == "__main__":
    main()

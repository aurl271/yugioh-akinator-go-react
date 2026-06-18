from __future__ import annotations

import argparse
import json
import sqlite3
from pathlib import Path
from typing import Any


BASE_DIR = Path(__file__).resolve().parents[2]
DEFAULT_DB_PATH = BASE_DIR / "output" / "CQA.db"
DEFAULT_MANUAL_PATH = BASE_DIR / "source_data" / "json" / "manual_readings.json"
DEFAULT_CACHE_PATH = BASE_DIR / "source_data" / "json" / "official_readings_cache.json"
DEFAULT_MISSING_PATH = BASE_DIR / "source_data" / "json" / "missing_readings.json"


def load_json(path: Path, default: Any) -> Any:
    if not path.exists():
        return default
    with path.open("r", encoding="utf-8") as file:
        return json.load(file)


def build_reading_map(manual_path: Path, cache_path: Path, missing_path: Path) -> dict[str, str]:
    readings: dict[str, str] = {}

    cache_readings = load_json(cache_path, {})
    for card_id, reading in cache_readings.items():
        if reading:
            readings[str(card_id)] = reading

    missing_cards = load_json(missing_path, [])
    for card in missing_cards:
        card_id = card.get("card_id")
        reading = card.get("reading")
        if card_id is not None and reading:
            readings[str(card_id)] = reading

    manual_readings = load_json(manual_path, {})
    for card_id, reading in manual_readings.items():
        if reading:
            readings[str(card_id)] = reading

    return readings


def apply_readings_to_cqa(db_path: Path, readings: dict[str, str]) -> int:
    updated = 0
    with sqlite3.connect(db_path) as conn:
        cursor = conn.cursor()
        for card_id, reading in readings.items():
            cursor.execute(
                """
                UPDATE cards
                SET reading = ?
                WHERE card_id = ?
                  AND (reading IS NULL OR reading != ?)
                """,
                (reading, int(card_id), reading),
            )
            updated += cursor.rowcount
        conn.commit()
    return updated


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Apply collected readings to output/CQA.db.")
    parser.add_argument("--db", type=Path, default=DEFAULT_DB_PATH)
    parser.add_argument("--manual", type=Path, default=DEFAULT_MANUAL_PATH)
    parser.add_argument("--cache", type=Path, default=DEFAULT_CACHE_PATH)
    parser.add_argument("--missing", type=Path, default=DEFAULT_MISSING_PATH)
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    if not args.db.exists():
        raise FileNotFoundError(f"CQA database not found: {args.db}")

    readings = build_reading_map(args.manual, args.cache, args.missing)
    updated = apply_readings_to_cqa(args.db, readings)
    print(f"readings to apply: {len(readings)}")
    print(f"updated CQA cards: {updated}")
    print(f"database: {args.db}")


if __name__ == "__main__":
    main()

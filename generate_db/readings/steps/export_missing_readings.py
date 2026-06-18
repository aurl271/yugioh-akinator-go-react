from __future__ import annotations

import argparse
import json
import sqlite3
from pathlib import Path


BASE_DIR = Path(__file__).resolve().parents[2]
DEFAULT_DB_PATH = BASE_DIR / "output" / "CQA.db"
DEFAULT_OUTPUT_PATH = BASE_DIR / "source_data" / "json" / "missing_readings.json"
DEFAULT_NOT_FOUND_PATH = BASE_DIR / "source_data" / "json" / "not_found_readings.json"


def load_json(path: Path, default):
    if not path.exists():
        return default
    with path.open("r", encoding="utf-8") as file:
        return json.load(file)


def fetch_missing_readings(db_path: Path, excluded_card_ids: set[str]) -> list[dict[str, int | str]]:
    with sqlite3.connect(db_path) as conn:
        conn.row_factory = sqlite3.Row
        cursor = conn.cursor()
        cursor.execute(
            """
            SELECT card_id, name
            FROM cards
            WHERE reading IS NULL
              AND name NOT LIKE '%トークン%'
            ORDER BY card_id
            """
        )
        cards = [
            {
                "card_id": row["card_id"],
                "name": row["name"],
            }
            for row in cursor.fetchall()
        ]
        return [card for card in cards if str(card["card_id"]) not in excluded_card_ids]


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Export cards whose reading is missing.")
    parser.add_argument("--db", type=Path, default=DEFAULT_DB_PATH)
    parser.add_argument("--output", type=Path, default=DEFAULT_OUTPUT_PATH)
    parser.add_argument("--not-found", type=Path, default=DEFAULT_NOT_FOUND_PATH)
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    if not args.db.exists():
        raise FileNotFoundError(f"CQA database not found: {args.db}")

    not_found_cards = load_json(args.not_found, [])
    excluded_card_ids = {str(card.get("card_id")) for card in not_found_cards}
    missing_cards = fetch_missing_readings(args.db, excluded_card_ids)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(missing_cards, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    print(f"missing readings: {len(missing_cards)}")
    print(f"output: {args.output}")
    for card in missing_cards[:20]:
        print(f'{card["card_id"]}: {card["name"]}')
    if len(missing_cards) > 20:
        print(f"... and {len(missing_cards) - 20} more")


if __name__ == "__main__":
    main()

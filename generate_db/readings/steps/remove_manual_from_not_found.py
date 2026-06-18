from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any


BASE_DIR = Path(__file__).resolve().parents[2]
DEFAULT_MANUAL_PATH = BASE_DIR / "source_data" / "json" / "manual_readings.json"
DEFAULT_NOT_FOUND_PATH = BASE_DIR / "source_data" / "json" / "not_found_readings.json"


def load_json(path: Path, default: Any) -> Any:
    if not path.exists():
        return default
    with path.open("r", encoding="utf-8") as file:
        return json.load(file)


def save_json(path: Path, data: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def collect_manual_card_ids(manual_readings: Any) -> set[str]:
    if not isinstance(manual_readings, dict):
        raise TypeError("manual_readings.json must be an object: {card_id: reading}")
    return {str(card_id) for card_id, reading in manual_readings.items() if reading}


def filter_not_found_cards(not_found_cards: Any, manual_ids: set[str]) -> tuple[list[dict[str, Any]], list[dict[str, Any]]]:
    if not isinstance(not_found_cards, list):
        raise TypeError("not_found_readings.json must be an array of card objects")

    kept_cards: list[dict[str, Any]] = []
    removed_cards: list[dict[str, Any]] = []
    for card in not_found_cards:
        card_id = str(card.get("card_id"))
        if card_id in manual_ids:
            removed_cards.append(card)
        else:
            kept_cards.append(card)
    return kept_cards, removed_cards


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Remove manually resolved cards from not_found_readings.json.")
    parser.add_argument("--manual", type=Path, default=DEFAULT_MANUAL_PATH)
    parser.add_argument("--not-found", type=Path, default=DEFAULT_NOT_FOUND_PATH)
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    manual_readings = load_json(args.manual, {})
    not_found_cards = load_json(args.not_found, [])

    manual_ids = collect_manual_card_ids(manual_readings)
    filtered_cards, removed_cards = filter_not_found_cards(not_found_cards, manual_ids)

    if removed_cards:
        save_json(args.not_found, filtered_cards)
    print(f"manual readings: {len(manual_ids)}")
    print(f"removed from not_found: {len(removed_cards)}")
    if removed_cards:
        print("removed card ids: " + ", ".join(str(card.get("card_id")) for card in removed_cards))
    print(f"output: {args.not_found}")


if __name__ == "__main__":
    main()

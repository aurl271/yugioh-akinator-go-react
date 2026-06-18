from __future__ import annotations

import argparse
import json
import unicodedata
from pathlib import Path
from typing import Any


BASE_DIR = Path(__file__).resolve().parents[2]
DEFAULT_INPUT_PATH = BASE_DIR / "source_data" / "json" / "missing_readings.json"
DEFAULT_OUTPUT_PATH = DEFAULT_INPUT_PATH


def is_fullwidth_katakana(char: str) -> bool:
    return "\u30A0" <= char <= "\u30FF"


def convert_except_katakana(text: str) -> str:
    converted = []
    for char in text:
        if is_fullwidth_katakana(char):
            converted.append(char)
        else:
            converted.append(unicodedata.normalize("NFKC", char))
    return "".join(converted)


def normalize_text_fields(value: Any) -> tuple[Any, int]:
    if isinstance(value, str):
        normalized = convert_except_katakana(value)
        return normalized, int(normalized != value)

    if isinstance(value, list):
        normalized_items = []
        changed_count = 0
        for item in value:
            normalized_item, item_changed_count = normalize_text_fields(item)
            normalized_items.append(normalized_item)
            changed_count += item_changed_count
        return normalized_items, changed_count

    if isinstance(value, dict):
        normalized_dict = {}
        changed_count = 0
        for key, item in value.items():
            normalized_item, item_changed_count = normalize_text_fields(item)
            normalized_dict[key] = normalized_item
            changed_count += item_changed_count
        return normalized_dict, changed_count

    return value, 0


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Normalize full-width characters in a readings JSON file.")
    parser.add_argument("--input", type=Path, default=DEFAULT_INPUT_PATH)
    parser.add_argument("--output", type=Path, default=DEFAULT_OUTPUT_PATH)
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    if not args.input.exists():
        print(f"skip normalize: file not found: {args.input}")
        return

    with args.input.open("r", encoding="utf-8") as file:
        readings_json = json.load(file)

    normalized, changed_count = normalize_text_fields(readings_json)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(normalized, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    print(f"normalized text fields: {changed_count}")
    print(f"output: {args.output}")


if __name__ == "__main__":
    main()

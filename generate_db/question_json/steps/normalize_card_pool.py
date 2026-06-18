from __future__ import annotations

import argparse
import json
import unicodedata
from pathlib import Path
from typing import Any


def is_fullwidth_katakana(char: str) -> bool:
    return "\u30A0" <= char <= "\u30FF"


def convert_except_katakana(text: str) -> str:
    """全角カタカナは残し、それ以外は NFKC で半角寄りに正規化する。"""
    converted = []
    for char in text:
        if is_fullwidth_katakana(char):
            converted.append(char)
        else:
            converted.append(unicodedata.normalize("NFKC", char))
    return "".join(converted)


def normalize_value(value: Any) -> Any:
    if isinstance(value, str):
        return convert_except_katakana(value)
    if isinstance(value, list):
        return [normalize_value(item) for item in value]
    if isinstance(value, dict):
        return {key: normalize_value(item) for key, item in value.items()}
    return value


def normalize_card_pool(card_pool_json_path: Path) -> dict[str, Any]:
    with card_pool_json_path.open("r", encoding="utf-8") as file:
        card_pool = json.load(file)
    return normalize_value(card_pool)


def default_input_path() -> Path:
    return Path(__file__).resolve().parents[2] / "source_data" / "json" / "CardPool.json"


def default_output_path() -> Path:
    return default_input_path()


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Normalize CardPool.json strings.")
    parser.add_argument("--input", type=Path, default=default_input_path())
    parser.add_argument("--output", type=Path, default=default_output_path())
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    normalized = normalize_card_pool(args.input)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(normalized, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"normalized CardPool in place: {args.output}")


if __name__ == "__main__":
    main()

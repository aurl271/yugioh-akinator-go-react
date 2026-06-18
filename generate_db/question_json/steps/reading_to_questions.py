from __future__ import annotations

import json
from pathlib import Path
from typing import Any


QuestionMap = dict[str, dict[str, Any]]


def _field_condition(field: str, op: str, value: str) -> dict[str, Any]:
    return {
        "field": field,
        "op": op,
        "value": value,
    }


def add_reading_chars(reading: str, first_chars: set[str], last_chars: set[str]) -> None:
    if not reading:
        return
    first_chars.add(reading[0])
    last_chars.add(reading[-1])


def default_official_cache_path(card_pool_json_path: Path) -> Path:
    return card_pool_json_path.parent / "official_readings_cache.json"


def load_official_readings_cache(official_cache_path: Path) -> dict[str, str]:
    if not official_cache_path.exists():
        return {}
    with official_cache_path.open("r", encoding="utf-8") as file:
        return json.load(file)


def build_reading_questions(card_pool_json_path: Path, official_cache_path: Path | None = None) -> QuestionMap:
    """CardPool.json と official_readings_cache.json の読みから質問を作る。"""
    with card_pool_json_path.open("r", encoding="utf-8") as file:
        card_pool = json.load(file)
    if official_cache_path is None:
        official_cache_path = default_official_cache_path(card_pool_json_path)
    official_readings_cache = load_official_readings_cache(official_cache_path)

    first_chars: set[str] = set()
    last_chars: set[str] = set()

    for card in card_pool["cards"]:
        add_reading_chars(card.get("ruby", ""), first_chars, last_chars)

    for reading in official_readings_cache.values():
        add_reading_chars(reading, first_chars, last_chars)

    questions: QuestionMap = {}
    note = "読み(公式データベースに準拠 大文字、小文字、ひらがな、カタカナは区別するが全角、半角は区別しない)"

    for char in sorted(first_chars):
        questions[f"カード名の読みが「{char}」で始まるカードですか？{note}"] = {
            "query": f"reading LIKE '{char}%'",
            "condition": _field_condition("reading", "starts_with", char),
            "unset_bit": 131072,
            "new_state": 131072,
        }

    for char in sorted(last_chars):
        questions[f"カード名の読みが「{char}」で終わるカードですか？{note}"] = {
            "query": f"reading LIKE '%{char}'",
            "condition": _field_condition("reading", "ends_with", char),
            "unset_bit": 131072,
            "new_state": 131072,
        }

    return questions

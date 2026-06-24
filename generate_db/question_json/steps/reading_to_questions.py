from __future__ import annotations

import json
from pathlib import Path
from typing import Any


QuestionMap = dict[str, dict[str, Any]]


def _condition(logic: str, conditions: list[dict[str, Any]]) -> dict[str, Any]:
    return {
        "logic": logic,
        "conditions": conditions,
    }


def _field_condition(field: str, op: str, text: str) -> dict[str, Any]:
    return {
        "field": field,
        "op": op,
        "text": text,
    }


def add_reading_chars(reading: str, first_chars: set[str], last_chars: set[str]) -> None:
    if not reading:
        return
    first_chars.add(reading[0])
    last_chars.add(reading[-1])


def default_official_cache_path(card_pool_json_path: Path) -> Path:
    return card_pool_json_path.parent / "official_readings_cache.json"


def default_manual_readings_path(card_pool_json_path: Path) -> Path:
    return card_pool_json_path.parent / "manual_readings.json"


def load_readings(path: Path) -> dict[str, str]:
    if not path.exists():
        return {}
    with path.open("r", encoding="utf-8") as file:
        return json.load(file)


def build_reading_questions(
    card_pool_json_path: Path,
    official_cache_path: Path | None = None,
    manual_readings_path: Path | None = None,
) -> QuestionMap:
    """CardPool.json、official_readings_cache.json、manual_readings.json の読みから質問を作る。"""
    with card_pool_json_path.open("r", encoding="utf-8") as file:
        card_pool = json.load(file)
    if official_cache_path is None:
        official_cache_path = default_official_cache_path(card_pool_json_path)
    if manual_readings_path is None:
        manual_readings_path = default_manual_readings_path(card_pool_json_path)
    official_readings_cache = load_readings(official_cache_path)
    manual_readings = load_readings(manual_readings_path)

    first_chars: set[str] = set()
    last_chars: set[str] = set()

    for card in card_pool["cards"]:
        add_reading_chars(card.get("ruby", ""), first_chars, last_chars)

    for reading in official_readings_cache.values():
        add_reading_chars(reading, first_chars, last_chars)

    for reading in manual_readings.values():
        add_reading_chars(reading, first_chars, last_chars)

    questions: QuestionMap = {}
    note = "読み(公式データベースに準拠 大文字、小文字、ひらがな、カタカナは区別するが全角、半角は区別しない)"

    for char in sorted(first_chars):
        questions[f"カード名の読みが「{char}」で始まるカードですか？{note}"] = {
            "query": f"reading LIKE '{char}%'",
            "condition": _condition(
                "and",
                [_field_condition("reading", "starts_with", char)],
            ),
            "unset_bit": 131072,
            "new_state": 131072,
        }

    for char in sorted(last_chars):
        questions[f"カード名の読みが「{char}」で終わるカードですか？{note}"] = {
            "query": f"reading LIKE '%{char}'",
            "condition": _condition(
                "and",
                [_field_condition("reading", "ends_with", char)],
            ),
            "unset_bit": 131072,
            "new_state": 131072,
        }

    return questions

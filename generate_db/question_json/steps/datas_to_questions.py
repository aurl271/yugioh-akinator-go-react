from __future__ import annotations

import difflib
import json
import sqlite3
from pathlib import Path
from typing import Any


QuestionMap = dict[str, dict[str, Any]]


FIXED_THEME_READINGS = {
    12296: "エレメンタルヒーロー",
    4373: "せんとうき",
    4288: "ひょういそうちゃく",
    4109: "エックス-セイバー",
    468: "エニアクラフト",
    442: "メタルか",
    432: "デモンスミス",
    290: "ワルキューレ",
    276: "くうがだん",
    271: "ヴァレル",
    192: "ひょうい",
    191: "れいつかい",
    115: "エクシーズ",
    124: "えんぶ",
    154: "ちょうじゅうむしゃ",
    164: "クリボー",
    188: "じんぞうにんげん",
    189: "あんこくきしガイア",
    239: "だてんし",
    43: "にんじゃ",
    52: "ほうぎょく",
    86: "インゼクター",
    97: "にんぽう",
    20602: "えんせいきし",
}

FIXED_MATCHED_SETCODES = {
    8: {
        8,
        12296,
        20488,
        24584,
        40968,
        49160,
        614408,
        1454088,
        13578248,
        19070984,
        19095560,
        25440264,
        38668283912,
        1666447912968,
        1666473799688,
    },
    9: {9, 614408, 4587529, 38668283912, 1666447912968},
    19: {19, 12307, 20499, 24595, 1519635},
    155: {155, 4251},
    158: {158, 12845214},
    165: {
        165,
        1507493,
        9568421,
        10813442,
        10813465,
        10813507,
        10813525,
        10813554,
        10813555,
        10813599,
        10817667,
        269942949,
        708946100309,
    },
    180: {180, 12845236, 55847420084},
    211: {211},
    227: {227},
    238: {238, 4334, 8430},
    239: {239},
    265: {265},
    383: {383},
    427: {427},
    425: {425},
    424: {424},
    304: {304, 4400},
    310: {310, 4587830},
}

SKIPPED_SETCODES = {
    35,
    5439644,
    602120,
    281018559,
    276758600,
    276762696,
}

LEADING_THEME_CHARS = {"・"}
TRAILING_THEME_CHARS = {"・", " ", "-"}

THEME_READING_REPLACEMENTS = {
    "霞の谷の": "霞の谷",
    "しんらの": "しんら",
    "のこわくま": "こわくま",
    "れいじゅうの": "れいじゅう",
    "サブテラーの": "サブテラー",
    "はかもりの": "はかもり",
    "ミスト・バレーの": "ミスト・バレー",
    "まかいだいほん「": "まかいだいほん",
}


def _condition(logic: str, conditions: list[dict[str, Any]]) -> dict[str, Any]:
    return {
        "logic": logic,
        "conditions": conditions,
    }


def _field_condition(field: str, op: str, **values: Any) -> dict[str, Any]:
    return {
        "field": field,
        "op": op,
        **values,
    }


def _fetch_values(cards_db_path: Path, sql: str) -> list[int]:
    with sqlite3.connect(cards_db_path) as conn:
        cursor = conn.cursor()
        cursor.execute(sql)
        return [row[0] for row in cursor.fetchall()]


def build_datas_questions(cards_db_path: Path) -> QuestionMap:
    """cards.cdb の datas テーブルから、数値で判定できる質問を作る。"""
    questions: QuestionMap = {}

    atk_values = _fetch_values(cards_db_path, "SELECT DISTINCT atk FROM datas ORDER BY atk")
    def_values = _fetch_values(
        cards_db_path,
        "SELECT DISTINCT def FROM datas WHERE type & 67108864 = 0 ORDER BY def",
    )

    for atk in atk_values:
        label = "?" if atk == -2 else str(atk)
        questions[f"攻撃力が{label}のモンスターですか？(魔法・罠ならいいえを選択)"] = {
            "query": f"type & 1 != 0 AND atk = {atk}",
            "condition": _condition(
                "and",
                [
                    _field_condition("type", "bit_on", value=1),
                    _field_condition("atk", "eq", value=atk),
                ],
            ),
            "unset_bit": 710,
            "new_state": 769,
        }

    for lower in range(0, 4501, 500):
        upper = lower + 500
        questions[f"攻撃力が{lower}以上{upper}以下のモンスターですか？(魔法・罠ならいいえを選択)"] = {
            "query": f"type & 1 != 0 AND {lower} <= atk AND atk <= {upper}",
            "condition": _condition(
                "and",
                [
                    _field_condition("type", "bit_on", value=1),
                    _field_condition("atk", "between", min=lower, max=upper),
                ],
            ),
            "unset_bit": 966,
            "new_state": 257,
        }

    for defence in def_values:
        label = "?" if defence == -2 else str(defence)
        questions[f"守備力が{label}のモンスターですか？(リンクモンスター・魔法・罠ならいいえを選択)"] = {
            "query": f"type & 67108864 = 0 AND type & 1 != 0 AND def = {defence}",
            "condition": _condition(
                "and",
                [
                    _field_condition("type", "bit_off", value=67108864),
                    _field_condition("type", "bit_on", value=1),
                    _field_condition("def", "eq", value=defence),
                ],
            ),
            "unset_bit": 2278,
            "new_state": 3073,
        }

    for lower in range(0, 4501, 500):
        upper = lower + 500
        questions[
            f"守備力が{lower}以上{upper}以下のモンスターですか？(リンクモンスター・魔法・罠ならいいえを選択)"
        ] = {
            "query": f"type & 67108864 = 0 AND type & 1 != 0 AND {lower} <= def AND def <= {upper}",
            "condition": _condition(
                "and",
                [
                    _field_condition("type", "bit_off", value=67108864),
                    _field_condition("type", "bit_on", value=1),
                    _field_condition("def", "between", min=lower, max=upper),
                ],
            ),
            "unset_bit": 3302,
            "new_state": 1025,
        }

    for level in range(1, 14):
        questions[
            f"レベル、ランク、リンクマーカーの数(未来龍王などはテキストに書いてあるレベル)が{level}のモンスターですか？(魔法・罠ならいいえを選択)"
        ] = {
            "query": f"type & 1 != 0 AND level & 0xff = {level}",
            "condition": _condition(
                "and",
                [
                    _field_condition("type", "bit_on", value=1),
                    _field_condition("level", "bit_mask_eq", mask=0xFF, value=level),
                ],
            ),
            "unset_bit": 198,
            "new_state": 4097,
        }

    for scale in range(14):
        questions[f"ペンデュラムスケールが{scale}のモンスターですか？(魔法・罠ならいいえを選択)"] = {
            "query": f"type & 16777216 != 0 AND (level>>16)&0xff = {scale}",
            "condition": _condition(
                "and",
                [
                    _field_condition("type", "bit_on", value=16777216),
                    _field_condition("level", "shift_mask_eq", shift=16, mask=0xFF, value=scale),
                ],
            ),
            "unset_bit": 198,
            "new_state": 24577,
        }

    return questions


def _load_json(path: Path) -> Any:
    if not path.exists():
        return {}
    with path.open("r", encoding="utf-8") as file:
        return json.load(file)


def _load_readings(card_pool_json_path: Path) -> tuple[dict[int, str], dict[str, str]]:
    card_id_to_reading: dict[int, str] = {}
    name_to_reading: dict[str, str] = {}

    card_pool = _load_json(card_pool_json_path)
    for card in card_pool.get("cards", []):
        reading = card.get("ruby", "")
        if not reading:
            continue
        card_id_to_reading[int(card["id"])] = reading
        name_to_reading[card["name"]] = reading

    json_dir = card_pool_json_path.parent
    for filename in ("official_readings_cache.json", "manual_readings.json"):
        readings = _load_json(json_dir / filename)
        for card_id, reading in readings.items():
            if reading:
                card_id_to_reading[int(card_id)] = reading

    return card_id_to_reading, name_to_reading


def _longest_common_part(readings: list[str]) -> str:
    if not readings:
        return ""

    common = readings[0]
    for reading in readings[1:]:
        matcher = difflib.SequenceMatcher(None, common, reading)
        match = matcher.find_longest_match(0, len(common), 0, len(reading))
        common = common[match.a : match.a + match.size]
        if not common:
            return ""

    return common


def _normalize_theme_reading(reading: str) -> str:
    normalized = reading.strip()
    while normalized and normalized[0] in LEADING_THEME_CHARS:
        normalized = normalized[1:].strip()
    while normalized and normalized[-1] in TRAILING_THEME_CHARS:
        normalized = normalized[:-1].strip()
    return THEME_READING_REPLACEMENTS.get(normalized, normalized)


def build_legacy_setcode_questions(
    cards_db_path: Path,
    card_pool_json_path: Path,
) -> QuestionMap:
    """setcode ごとの reading 共通部分から、setcode 質問を作る。"""
    card_id_to_reading, name_to_reading = _load_readings(card_pool_json_path)

    with sqlite3.connect(cards_db_path) as conn:
        cursor = conn.cursor()
        cursor.execute(
            """
            SELECT datas.id, datas.setcode, texts.name
            FROM datas
            JOIN texts ON texts.id = datas.id
            WHERE datas.setcode != 0
            ORDER BY datas.setcode, datas.id
            """
        )
        card_rows = cursor.fetchall()

    cards: list[dict[str, Any]] = []
    readings_by_setcode: dict[int, list[str]] = {}

    for card_id, setcode, name in card_rows:
        reading = card_id_to_reading.get(card_id) or name_to_reading.get(name, "")
        if not reading:
            continue

        card = {
            "card_id": card_id,
            "setcode": setcode,
            "name": name,
            "reading": reading,
        }
        cards.append(card)
        readings_by_setcode.setdefault(setcode, []).append(reading)

    questions: QuestionMap = {}
    used_theme_readings: set[str] = set()

    for setcode in sorted(readings_by_setcode):
        if setcode in SKIPPED_SETCODES:
            continue

        readings = readings_by_setcode[setcode]
        if len(readings) < 5:
            continue

        raw_theme_reading = FIXED_THEME_READINGS.get(setcode) or _longest_common_part(readings)
        theme_reading = _normalize_theme_reading(raw_theme_reading)
        if not theme_reading or theme_reading in used_theme_readings:
            continue

        used_theme_readings.add(theme_reading)
        if setcode in FIXED_MATCHED_SETCODES:
            matched_setcodes = FIXED_MATCHED_SETCODES[setcode]
        else:
            matched_setcodes = {
                card["setcode"]
                for card in cards
                if card["setcode"] >= setcode
                and card["setcode"] not in SKIPPED_SETCODES
                and theme_reading in card["reading"]
            }

        query = " OR ".join(f"setcode = {matched_setcode}" for matched_setcode in sorted(matched_setcodes))
        questions[f"「{theme_reading}」カードですか？"] = {
            "query": query,
            "condition": _condition(
                "or",
                [_field_condition("setcode", "eq", value=matched_setcode) for matched_setcode in sorted(matched_setcodes)],
            ),
            "unset_bit": 0,
            "new_state": 0,
        }

    return questions

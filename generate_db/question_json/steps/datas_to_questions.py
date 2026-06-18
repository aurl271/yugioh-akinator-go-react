from __future__ import annotations

import difflib
import sqlite3
from pathlib import Path
from typing import Any


QuestionMap = dict[str, dict[str, Any]]


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


USED_LEGACY_SETCODES = {
    2,
    7,
    11,
    12,
    23,
    25,
    27,
    30,
    31,
    35,
    37,
    38,
    43,
    52,
    54,
    56,
    58,
    59,
    68,
    69,
    70,
    82,
    83,
    85,
    86,
    97,
    111,
    113,
    115,
    123,
    126,
    127,
    137,
    147,
    149,
    156,
    157,
    163,
    164,
    165,
    170,
    172,
    173,
    186,
    191,
    196,
    198,
    207,
    219,
    220,
    221,
    223,
    226,
    229,
    234,
    242,
    243,
    268,
    273,
    274,
    277,
    281,
    283,
    284,
    290,
    291,
    298,
    301,
    308,
    317,
    320,
    325,
    336,
    340,
    345,
    347,
    352,
    356,
    378,
    381,
    385,
    390,
    392,
    393,
    395,
    400,
    401,
    406,
    407,
    411,
    412,
    418,
    419,
    430,
    432,
    442,
    444,
    448,
    453,
    717,
    4288,
    4316,
    4373,
    20602,
    4260113,
    281018559,
}


LEGACY_CATEGORY_NAMES = [
    [2, "ジェネクス"],
    [11, "インフェルニティ"],
    [12, "エーリアン"],
    [23, "シンクロ"],
    [29, "コアキメイル"],
    [31, "ネオスペーシアン", "Ｎ"],
    [35, "Ｓｉｎ"],
    [43, "忍者"],
    [52, "宝玉"],
    [54, "マシンナーズ"],
    [56, "ライトロード"],
    [58, "リチュア"],
    [59, "レッドアイズ", "真紅眼"],
    [68, "代行者"],
    [69, "デーモン"],
    [70, "融合", "フュージョン"],
    [82, "ガーディアン"],
    [83, "セイクリッド"],
    [85, "フォトン"],
    [86, "甲虫装機"],
    [97, "忍法"],
    [111, "ヒロイック", "Ｈ－Ｃ"],
    [113, "マドルチェ"],
    [115, "エクシーズ", "ＣＸ", "レイ・ピアース"],
    [123, "ギャラクシー", "銀河"],
    [126, "炎舞"],
    [127, "ホープ"],
    [147, "サイバー"],
    [149, "ＲＵＭ"],
    [156, "テラナイト", "星因子", "星輝士"],
    [157, "シャドール", "影依", "神の写し身との接触", "魂写しの同化"],
    [163, "スターダスト"],
    [164, "クリボー"],
    [165, "チェンジ", "紋章変換"],
    [170, "クリフォート"],
    [172, "ゴブリン", "百鬼羅刹"],
    [173, "デストーイ", "魔玩具"],
    [186, "ＲＲ", "レイド・ラプターズ", "Ｒ・Ｒ・Ｒ", "レイダーズ・アンブレイカブル・マインド"],
    [191, "霊使い"],
    [196, "セフィラ"],
    [198, "Ｅｍ"],
    [207, "カオス", "混沌", "", "ヌメロニアス・ヌメロニア", "ＣＨＡＯＳ", "ＣＸ", "ＣＮ"],
    [219, "ファントム", "幻影"],
    [220, "超量"],
    [221, "ブルーアイズ", "青眼"],
    [223, "月光"],
    [226, "トラミッド"],
    [229, "サイファー", "光波"],
    [234, "クリストロン", "水晶機巧"],
    [242, "ペンデュラム", "軌跡の魔術師", "奇跡の魔導剣士", "ドラゴニックＰ", "竜剣士マスター", "竜剣士ラスター", "竜魔王ベクター", "竜魔王レクター"],
    [243, "プレデター", "捕食"],
    [268, "ジャックナイツ", "機界騎士", "宵星の騎士", "明星の機械騎士", "双穹の騎士アストラム"],
    [273, "アームド・ドラゴン", "武装竜"],
    [274, "トロイメア", "夢幻転星イドリース", "夢幻崩界イヴリース"],
    [277, "閃刀", "未来の柱－キアノス", "智の賢者－ヒンメル", "閃術兵器－.", "慈愛の賢者－シエラ", "エルロン", "武の賢者－アーカス"],
    [281, "サラマングレイト", "転生炎獣", "フュージョン・オブ・ファイア", "フューリー・オブ・ファイア", "ライジング・オブ・ファイア"],
    [283, "オルフェゴール", "宵星の機神", "宵星の騎士"],
    [284, "サンダー・ドラゴン", "雷龍"],
    [290, "ワルキューレ", "戦乙女の戦車", "運命の戦車", "Ｗａｌｋｕｒｅｎ"],
    [291, "ローズ"],
    [298, "エンディミオン", "魔法都市の実験施設"],
    [301, "シムルグ"],
    [320, "アダマシア", "魔救"],
    [325, "ドラグマ", "凶導", "教導", "白の枢機竜", "烙印の命数", "導きの聖女クエム"],
    [336, "マギストス", "聖月の魔導士エンディミオン", "聖魔の大賢者エンディミオン"],
    [340, "ドライトロン", "輝巧", "竜儀巧"],
    [356, "デスピア", "導きの聖女クエム", "導きの聖女クエム"],
    [378, "スケアクロー", "肆世壊"],
    [381, "ヴァリアンツ"],
    [382, "ラビュリンス", "白銀の城"],
    [385, "ティアラメンツ", "壱世壊"],
    [393, "クシャトリラ", "六世壊"],
    [400, "マナドゥム", "伍世壊"],
    [407, "レシピ"],
    [411, "ディアベル", "蛇眼の大炎魔"],
    [412, "スネークアイ", "蛇眼"],
    [430, "千年", "ミレニアム"],
    [442, "メタル化"],
    [444, "アザミナ"],
    [453, "リゼェネシス", "再世"],
    [717, "アルトメギア", "神芸"],
    [4316, "超量士"],
    [4373, "閃刀姫"],
    [20602, "焔聖騎士"],
]


def build_legacy_setcode_questions(cards_db_path: Path) -> QuestionMap:
    """旧 datas_to_json.py のコメントアウト部分を再現する。

    元コードでは JSON として完成した形ではなく、テーマ名と query だけを出力していた。
    そのため、この関数も通常の生成には混ぜず、setcode 質問を作り直すための材料として残す。
    """
    categories = [list(category) for category in LEGACY_CATEGORY_NAMES]

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

    names_by_setcode: dict[int, list[str]] = {}
    setcode_by_id: dict[int, int] = {}
    all_names: list[tuple[int, str]] = []

    for card_id, setcode, name in card_rows:
        names_by_setcode.setdefault(setcode, []).append(name)
        setcode_by_id[card_id] = setcode
        all_names.append((card_id, name))

    for setcode in sorted(names_by_setcode):
        if setcode in USED_LEGACY_SETCODES:
            continue

        card_names = names_by_setcode[setcode]
        if len(card_names) < 5:
            continue

        theme_name = card_names[0]
        for card_name in card_names:
            matcher = difflib.SequenceMatcher(None, theme_name, card_name)
            match = matcher.find_longest_match(0, len(theme_name), 0, len(card_name))
            theme_name = theme_name[match.a : match.a + match.size]

        if theme_name:
            categories.append([setcode, theme_name])

    questions: QuestionMap = {}
    for category in categories:
        setcodes = {category[0]}
        for category_name in category[1:]:
            if not category_name:
                continue

            for card_id, card_name in all_names:
                if category_name in card_name:
                    setcodes.add(setcode_by_id[card_id])

        query = " OR ".join(f"setcode = {setcode}" for setcode in sorted(setcodes))
        questions[f"「{category[1]}」カードですか？(テキストにルール上「～」カードとして扱う場合も含む)"] = {
            "query": query,
            "condition": _condition(
                "or",
                [_field_condition("setcode", "eq", value=setcode) for setcode in sorted(setcodes)],
            ),
            "unset_bit": 0,
            "new_state": 0,
        }

    return questions

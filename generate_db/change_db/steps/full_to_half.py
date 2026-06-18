import sqlite3
import unicodedata


# カード名の比較を安定させるため、全角カタカナ以外を NFKC で正規化する。
def is_fullwidth_katakana(char):
    return "\u30A0" <= char <= "\u30FF"


def convert_except_katakana(text):
    result = []
    for char in text:
        if is_fullwidth_katakana(char):
            # カタカナはカード名としての見た目を保つ。
            result.append(char)
        else:
            result.append(unicodedata.normalize("NFKC", char))
    return "".join(result)


def normalize_card_names(cards_db):
    conn = sqlite3.connect(cards_db)
    cursor = conn.cursor()

    cursor.execute("SELECT id, name FROM texts")
    rows = cursor.fetchall()

    changed = 0
    for card_id, name in rows:
        normalized = convert_except_katakana(name)
        if normalized == name:
            continue

        # texts.name のみを正規化する。カード効果テキストやID系の値はここでは触らない。
        cursor.execute("UPDATE texts SET name = ? WHERE id = ?", (normalized, card_id))
        changed += 1

    conn.commit()
    conn.close()

    print(f"normalized card names: {changed}")

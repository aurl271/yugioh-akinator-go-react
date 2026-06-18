import sqlite3


# ot=2 は海外先行カード。読みデータの扱いが面倒だったため生成対象から外す。
def delete_ot2(cards_db):
    conn = sqlite3.connect(cards_db)
    cursor = conn.cursor()

    cursor.execute("SELECT id FROM datas WHERE ot = 2")
    card_ids = [row[0] for row in cursor.fetchall()]

    for card_id in card_ids:
        # texts と datas は同じ id で1枚のカードを表すので、必ず両方から消す。
        cursor.execute("DELETE FROM texts WHERE id = ?", (card_id,))
        cursor.execute("DELETE FROM datas WHERE id = ?", (card_id,))

    conn.commit()
    conn.close()

    print(f"deleted ot=2 cards: {len(card_ids)}")

import sqlite3


# 同名カードのうち、alias が同名グループ内の別IDを指しているものを絵違い扱いで削除する。
def delete_same_name_aliases(cards_db):
    conn = sqlite3.connect(cards_db)
    cursor = conn.cursor()

    # まず同名カードが複数ある名前だけを対象に絞る。
    cursor.execute("SELECT name FROM texts GROUP BY name HAVING COUNT(*) > 1")
    duplicate_names = [row[0] for row in cursor.fetchall()]

    deleted_ids = set()
    for name in duplicate_names:
        cursor.execute("SELECT id FROM texts WHERE name = ?", (name,))
        card_ids = [row[0] for row in cursor.fetchall()]
        card_id_set = set(card_ids)

        for card_id in card_ids:
            cursor.execute("SELECT alias FROM datas WHERE id = ?", (card_id,))
            row = cursor.fetchone()
            if row is None:
                continue

            alias = row[0]
            if alias in card_id_set:
                # texts と datas の片方だけ残ると後続生成で壊れるため、両方から削除する。
                cursor.execute("DELETE FROM texts WHERE id = ?", (card_id,))
                cursor.execute("DELETE FROM datas WHERE id = ?", (card_id,))
                deleted_ids.add(card_id)

    conn.commit()
    conn.close()

    print(f"deleted duplicate same-name cards: {len(deleted_ids)}")

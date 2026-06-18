import sqlite3


# pre-release の cards DB を本体 cards.cdb に取り込む。
# datas/texts は同じ id を主キーに持つため、重複は INSERT OR IGNORE で残す。
def table_count(cursor, table):
    cursor.execute(f"SELECT COUNT(*) FROM {table}")
    return cursor.fetchone()[0]


def merge_pre_release(cards_db, pre_release_db):
    if not pre_release_db.exists():
        print(f"skip merge: pre-release database not found: {pre_release_db}")
        return

    conn = sqlite3.connect(cards_db)
    cursor = conn.cursor()
    attached = False

    try:
        before_datas = table_count(cursor, "datas")
        before_texts = table_count(cursor, "texts")

        # ATTACH で別DBを同じ接続から参照し、テーブル単位でコピーする。
        cursor.execute("ATTACH DATABASE ? AS src", (str(pre_release_db),))
        attached = True
        for table in ("datas", "texts"):
            cursor.execute(f"""
                INSERT OR IGNORE INTO {table}
                SELECT * FROM src.{table}
            """)

        # DETACH は未完了のトランザクションがあると database is locked になる。
        conn.commit()
        cursor.execute("DETACH DATABASE src")
        attached = False

        after_datas = table_count(cursor, "datas")
        after_texts = table_count(cursor, "texts")
    finally:
        if attached:
            conn.rollback()
            cursor.execute("DETACH DATABASE src")
        conn.close()

    print(f"merged pre-release: datas +{after_datas - before_datas}, texts +{after_texts - before_texts}")

#prereleaseのcdbとcards.cdbを結合
import sqlite3
import os

def merge_databases(target_db, source_db, tables):
    #target_dbにsource_dbを追加
    conn = sqlite3.connect(target_db)
    cursor = conn.cursor()

    #アタッチ
    cursor.execute(f"ATTACH DATABASE '{source_db}' AS src")

    for table in tables:
        # 重複対策で INSERT OR IGNORE にしている
        cursor.execute(f"""
            INSERT OR IGNORE INTO {table}
            SELECT * FROM src.{table}
        """)

    #保存して終了
    conn.commit()
    cursor.execute("DETACH DATABASE src")
    cursor.close()
    conn.close()

#データベースへのパス
current_dir = os.path.dirname(os.path.abspath(__file__))
db_path = os.path.dirname(current_dir)
db_path = os.path.join(db_path, "cards.cdb")
pre_db_path = os.path.dirname(current_dir)
pre_db_path = os.path.join(pre_db_path, "pre_release.cdb")
merge_databases(db_path, pre_db_path, ["datas", "texts"])

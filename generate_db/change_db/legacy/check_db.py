#texts,datasの片方にしかないカードの削除
import sqlite3
import os

def check_db(db_path,table1,table2):
    #table1,table2の片方にしかないカードの削除
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    #table1sテーブルとtable2テーブルのcardidを取得
    cursor.execute(f"SELECT id FROM {table1}")
    table1_ids_values = [row[0] for row in cursor.fetchall()]
    cursor.execute(f"SELECT id FROM {table2}")
    table2_ids_values = [row[0] for row in cursor.fetchall()]

    #片方にしかないのを削除
    for table1_id in table1_ids_values:
        if not table1_id in table2_ids_values:
            print(f"table1_id:{table1_id}")
            cursor.execute(f"DELETE FROM {table1} WHERE id = ?",(table1_id,))
            
    for table2_id in table2_ids_values:
        if not table2_id in table1_ids_values:
            print(f"table2_id:{table2_id}")
            cursor.execute(f"DELETE FROM {table2} WHERE id = ?",(table2_id,))
    
    #保存して終了
    conn.commit()
    cursor.close()
    conn.close()

#データべースへのパス
current_dir = os.path.dirname(os.path.abspath(__file__))
db_path = os.path.dirname(current_dir)
db_path = os.path.join(db_path, "cards.cdb")
#関数呼び出し
check_db(db_path,'datas','texts')
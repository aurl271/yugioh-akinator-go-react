#ot==2となっている海外先行カードの削除(読みが面倒から)
import sqlite3
import os


def delete_databases(db_path, text_table, data_table):
    #ot==2となっている海外先行カードの削除(読みが面倒から)
    #データベースを開く
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()

    #ot==2となっているidを取り出す
    cursor.execute(f'SELECT id FROM {data_table} WHERE ot = 2')
    ids_values = [id[0] for id in cursor.fetchall()]
    
    #text_table,data_tableのot==2となっているidの行を削除
    for id in ids_values:
        cursor.execute(f'DELETE FROM {text_table} WHERE id = ?;',(id,))
        cursor.execute(f'DELETE FROM {data_table} WHERE id = ?;',(id,))
    
    #保存して終了    
    conn.commit()
    cursor.close()
    conn.close()

#データべースへのパス
current_dir = os.path.dirname(os.path.abspath(__file__))
db_path = os.path.dirname(current_dir)
db_path = os.path.join(db_path, "cards.cdb")
#関数呼び出し
delete_databases(db_path, "texts","datas")

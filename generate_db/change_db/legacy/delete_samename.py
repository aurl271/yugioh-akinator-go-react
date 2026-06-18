#同名カードの削除
import sqlite3
import os

def delete_databases(db_path, text_table, data_table):
    #データベースを開く
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()

    #同じ名前で2つ以上行を持っているnameを取り出す
    cursor.execute(f'SELECT name FROM {text_table} GROUP BY name HAVING COUNT(*) > 1;')
    names_values = [id[0] for id in cursor.fetchall()]
    
    for name in names_values:
        #名前からidを取り出す
        cursor.execute(f'SELECT id FROM {text_table} WHERE name = ?;',(name,))
        ids_values = [id[0] for id in cursor.fetchall()]
        for id in ids_values:
            #aliasを取り出す
            cursor.execute(f'SELECT alias FROM {data_table} WHERE id = ?;',(id,))
            alias = cursor.fetchone()[0]
            if alias in ids_values:
                #aliasがids_valuesにあれば、絵違いの同名カードだから削除
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

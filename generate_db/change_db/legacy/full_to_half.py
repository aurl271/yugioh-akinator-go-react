#カタカナ以外の全角文字を半角にする
import sqlite3
import os
import unicodedata

def is_fullwidth_katakana(char):
    # 全角カタカナの範囲：U+30A0〜U+30FF
    return '\u30A0' <= char <= '\u30FF'

def convert_except_katakana(text):
    #カタカナ以外を半角に
    result = []
    for char in text:
        if is_fullwidth_katakana(char):
            result.append(char)
        else:
            result.append(unicodedata.normalize('NFKC', char))
    return ''.join(result)


def full_to_half(db_path, text_table):
    #データベースを開く
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()

    #id,name,descの取り出す
    cursor.execute(f"SELECT id,name FROM {text_table}")
    for id, name in cursor.fetchall():
        #カタカナ以外を半角にして上書き
        name = convert_except_katakana(name)
        cursor.execute(f"UPDATE {text_table} SET name = ? WHERE id = ?;",(name,id))

    #保存して終了
    conn.commit()
    cursor.close()
    conn.close()
    
#データべースへのパス
current_dir = os.path.dirname(os.path.abspath(__file__))
db_path = os.path.dirname(current_dir)
db_path = os.path.join(db_path, "cards.cdb")
#関数呼び出し
full_to_half(db_path,"texts")
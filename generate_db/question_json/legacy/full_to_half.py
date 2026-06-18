import sys
import io
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

#utf-8で出力
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

#CardPool.jsonを飛来て全角を半角にする
current_dir = os.path.dirname(os.path.abspath(__file__))
db_path = os.path.join(current_dir, "CardPool.json")
with open(db_path, 'r', encoding='utf-8') as f:
    content = f.read()
    print(convert_except_katakana(content))
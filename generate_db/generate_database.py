"""
generate_database.py
カード、質問、回答のデータベースを作成するプログラム
以下の3つのテーブルを作成し、カード、質問、回答を追加、削除できるようにする
cardsテーブル
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    card_id INTEGER NOT NULL UNIQUE, -- card_id
    name TEXT NOT NULL  -- カード名
    reading TEXT -- カードの読み方
    desc TEXT NOT NULL -- テキスト
    setcode INTEGER NOT NULL -- テーマ指定
    type INTEGER NOT NULL -- カード種類
    atk INTEGER NOT NULL -- 攻撃力
    def INTEGER NOT NULL -- 守備力
    level INTEGER NOT NULL -- レベル
    race INTEGER NOT NULL -- 種族
    attribute INTEGER NOT NULL -- 属性
    
questionsテーブル
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    text TEXT NOT NULL UNIQUE,   -- 質問文
    category INTEGER NOT NULL   -- scriptかcardsテーブルからなのか
    query TEXT --  cardsテーブルから回答を判断するためのクエリ(scriptから判断する場合はNULL)
    condition_json TEXT -- query と同じ意味の構造化された判定条件
    unset_bit INTEGER NOT NULL DEFAULT 0,   -- ビットが立っていればこの質問はしない
    new_state INTEGER NOT NULL DEFAULT 0    -- 新しい状態にするためのビット
    
answersテーブル
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    card_id INTEGER NOT NULL,   -- card_id
    question_id INTEGER NOT NULL,   -- 質問id
    answer INTEGER CHECK(answer IN (1, -1)) NOT NULL, -- 回答
    FOREIGN KEY (card_id) REFERENCES cards(card_id), -- cardsテーブルにあるcard_idしか使わない
    FOREIGN KEY (question_id) REFERENCES questions(id), -- questionsテーブルにあるidしか使わない
    UNIQUE(card_id, question_id)  -- 同じカードに同じ質問を複数登録しないよう制限
"""

#sqlを使いたいのでインポート
import sqlite3
#正規表現を使いたいのでインポート
import re
import sys
import time
#enumを使いたいのでインポート
from enum import Enum
#dbを同じディレクトリにしたいのでインポート
import os
#jsonを扱うのでインポート
import json
from pathlib import Path

class QuestionCategory(Enum):
    #何から回答を判断するかどうか
    #スクリプトから(例:破壊効果を持ちますか？)
    SCRIPT = 0
    #cardsテーブルから判断(例:モンスターですか？)
    CARDS = 1

    
class AnswerValue(Enum):
    #回答がYES,NOのときの値
    YES = 1
    NO = -1
    PROBABLY = 0.5
    PROBABLY_NO = -0.5
    UNKNOWN = 0


#データベースの名前
CARD_QUESTION_ANSWER_DB_NAME = "output/CQA.db"
#元からあるカードデータベース
CARD_DATABASE = "source_data/cards.cdb"
#質問と判断材料のjson
SCRIPT_TO_QUESTION_JSON = "source_data/json/script_to_question.json"
CARDS_TO_QUESTION_JSON = "source_data/json/generated_cards_to_question.json"
NAME_READING_JSON = "source_data/json/CardPool.json"


def log(message):
    # 生成ログはstdoutのJSON等と混ざらないようstderrへ出す。
    print(message, file=sys.stderr)


def format_elapsed(start_time):
    # perf_counterの開始時刻から経過秒数の表示文字列を作る。
    return f"{time.perf_counter() - start_time:.2f}s"


class CardDb:
    def __init__(self,
                db_name = CARD_QUESTION_ANSWER_DB_NAME,
                carddb_name = CARD_DATABASE,
                script_json_name = SCRIPT_TO_QUESTION_JSON, 
                cards_json_name = CARDS_TO_QUESTION_JSON,
                name_reading_json_name = NAME_READING_JSON):
        # generate_db を基準に、入力ファイルと出力DBのpathを設定する。
        base_dir = Path(__file__).resolve().parent
        db_path = base_dir / db_name
        carddb_path = base_dir / carddb_name
        script_json_path = base_dir / script_json_name
        cards_json_path = base_dir / cards_json_name
        name_reading_json_path = base_dir / name_reading_json_name
        db_path.parent.mkdir(parents=True, exist_ok=True)
        self.db_path = db_path
        self.carddb_path = carddb_path
        self.script_json_path = script_json_path
        self.cards_json_path = cards_json_path
        self.name_reading_json_path = name_reading_json_path
        
        #質問とluascriptのセットの読み込み
        self.script_json = self.read_json(script_json_path)
        
        #質問とdatasテーブルのセットの読み込み
        self.cards_json = self.read_json(cards_json_path)
        
        #カード名の読みの読み込み
        self.name_reading_json = self.read_json(name_reading_json_path)
        
        #生成するsqliteの初期設定
        self.conn = sqlite3.connect(db_path)
        self.conn.execute("PRAGMA foreign_keys = ON")
        self.cursor = self.conn.cursor()
        self.create_tables()
        
        #元からあるデータベースの読み込みsqliteの初期設定
        self.cdbconn = sqlite3.connect(carddb_path)
        self.cdbcursor = self.cdbconn.cursor()
        
    def read_json(self,filepath):
        # 質問定義や読み仮名JSONをPythonオブジェクトとして読み込む。
        try:
            #jsonの読み込み
            with open(filepath, 'r', encoding='utf-8') as f:
                return json.load(f)
        except Exception as e:
            #エラー発生時
            print(f"json読み込みエラー: {e}")

    def build_name_to_reading(self):
        # CardPool.jsonのカード名と読み仮名を、名前から引ける辞書に変換する。
        name_to_reading = {}
        for card in self.name_reading_json["cards"]:
            if "name" not in card or "ruby" not in card:
                continue
            name_to_reading[card["name"]] = card["ruby"]
        return name_to_reading

    def create_tables(self):
        try:
            #cards,questions,answersテーブル作成
            self.cursor.execute("""
                CREATE TABLE IF NOT EXISTS cards (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    card_id INTEGER NOT NULL UNIQUE,
                    name TEXT NOT NULL,
                    reading TEXT,
                    desc TEXT,
                    setcode INTEGER NOT NULL,
                    type INTEGER NOT NULL,
                    atk INTEGER NOT NULL,
                    def INTEGER NOT NULL,
                    level INTEGER NOT NULL,
                    race INTEGER NOT NULL,
                    attribute INTEGER NOT NULL
                )
            """)
            self.cursor.execute(f"""
                CREATE TABLE IF NOT EXISTS questions (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    question_text TEXT NOT NULL UNIQUE,
                    category INTEGER CHECK(category IN ({QuestionCategory.SCRIPT.value}, {QuestionCategory.CARDS.value})) NOT NULL,
                    query TEXT,
                    condition_json TEXT,
                    unset_bit INTEGER NOT NULL DEFAULT 0,
                    new_state INTEGER NOT NULL DEFAULT 0
                )
            """)
            self.cursor.execute(f"""
                CREATE TABLE IF NOT EXISTS answers (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    card_id INTEGER NOT NULL,
                    question_id INTEGER NOT NULL,
                    answer INTEGER CHECK(answer IN ({AnswerValue.YES.value}, {AnswerValue.NO.value})) NOT NULL,
                    FOREIGN KEY (card_id) REFERENCES cards(card_id),
                    FOREIGN KEY (question_id) REFERENCES questions(id),
                    UNIQUE(card_id, question_id)
                )
            """)
            self.add_column_if_missing("questions", "condition_json", "TEXT")
            self.conn.commit()
        except Exception as e:
            #エラー発生時はロールバック
            self.conn.rollback()
            print(f"データベース生成時にエラーが発生しました: {e}")

    def add_column_if_missing(self, table_name, column_name, column_definition):
        # 既存DBを再利用する場合でも、足りない列だけを安全に追加する。
        self.cursor.execute(f"PRAGMA table_info({table_name})")
        columns = [row[1] for row in self.cursor.fetchall()]
        if column_name not in columns:
            self.cursor.execute(f"ALTER TABLE {table_name} ADD COLUMN {column_name} {column_definition}")

    def table_count(self, table_name):
        # 生成ログで差分件数を出すため、指定テーブルの行数を数える。
        self.cursor.execute(f"SELECT COUNT(*) FROM {table_name}")
        return self.cursor.fetchone()[0]

    def add_card(self,card_id,card_name,card_reading,card_desc,card_setcode,card_type,card_atk,card_def,card_level,card_race,card_attribute):
        try:
            #cardsテーブルにカードの追加、answersテーブルにカード、質問、回答を追加
            self.cursor.execute("SELECT 1 FROM cards WHERE card_id = ?;", (card_id,))
            result = self.cursor.fetchone()
            if not result:
                #cardsテーブルになければ追加
                self.cursor.execute("INSERT INTO cards (card_id,name,reading,desc,setcode,type,atk,def,level,race,attribute) VALUES (?,?,?,?,?,?,?,?,?,?,?);", (card_id,card_name,card_reading,card_desc,card_setcode,card_type,card_atk,card_def,card_level,card_race,card_attribute))        
            elif card_reading is not None:
                self.cursor.execute("UPDATE cards SET reading = ? WHERE card_id = ?;", (card_reading, card_id))
            
            for question_id,question_text in self.get_all_question_script():
                #questionsテーブルからquestionを取り出し、answersテーブルにカード、質問、回答を追加
                self.cursor.execute("SELECT 1 FROM answers WHERE card_id = ? AND question_id = ?;", (card_id,question_id))
                result = self.cursor.fetchone()
                if not result and self.get_script_answer(card_id,question_text) == AnswerValue.YES.value:
                    #answersテーブルにカード、質問、回答のセットがなければ追加
                    self.add_answer(card_id,question_id,AnswerValue.YES.value)
            
            self.conn.commit()
            
        except Exception as e:
            #エラー発生時はロールバック
            self.conn.rollback()
            print(f"カード追加時にエラーが発生しました: {e}")
            print(card_id,card_name)
            
    def add_question(self,question_text,category,query = None,unset_bit = 0,new_state = 0,condition = None):
        question_id = None
        try:
            condition_json = None
            if condition is not None:
                condition_json = json.dumps(condition, ensure_ascii=False)

            #cardsテーブルにカードの追加、answersテーブルにカード、質問、回答を追加
            self.cursor.execute("SELECT 1 FROM questions WHERE question_text = ?;", (question_text,))
            result = self.cursor.fetchone()
            if not result:
                #questionsテーブルになければ追加
                self.cursor.execute("INSERT INTO questions (question_text,category,query,condition_json,unset_bit,new_state) VALUES (?,?,?,?,?,?);", (question_text,category,query,condition_json,unset_bit,new_state))        
            else:
                self.cursor.execute(
                    "UPDATE questions SET category = ?, query = ?, condition_json = ?, unset_bit = ?, new_state = ? WHERE question_text = ?;",
                    (category, query, condition_json, unset_bit, new_state, question_text),
                )
            
            self.cursor.execute("SELECT id FROM questions WHERE question_text = ?;", (question_text,))
            result = self.cursor.fetchone()
            if result:
                #question_idの取得
                question_id = result[0]
            else:
                raise ValueError("question_idが取得できませんでした")
            
            if category == QuestionCategory.SCRIPT.value:   
                for card_id in self.get_all_cards_id():
                    #questionsテーブルからquestionを取り出し、answersテーブルにカード、質問、回答を追加
                    self.cursor.execute("SELECT 1 FROM answers WHERE card_id = ? AND question_id = ?;", (card_id,question_id))
                    result = self.cursor.fetchone()
                    if not result and self.get_script_answer(card_id,question_text) == AnswerValue.YES.value:
                        #answersテーブルにカード、質問、回答のセットがなければ追加
                        self.add_answer(card_id,question_id,AnswerValue.YES.value)
            
            self.conn.commit()
            
        except Exception as e:
            #エラー発生時はロールバック
            self.conn.rollback()
            print(f"質問追加時にエラーが発生しました: {e}")
            print(question_id,question_text,category,unset_bit,new_state)
    
    def add_answer(self,card_id,question_id,answer):
        try:
            #answersテーブルにカード、質問、回答を追加
            self.cursor.execute("INSERT INTO answers (card_id,question_id,answer) VALUES (?,?,?);", (card_id,question_id,answer))
        except Exception as e:
            #エラー発生時はロールバック
            self.conn.rollback()
            print(f"回答追加時にエラーが発生しました: {e}")
            print(card_id,question_id,answer)
        
    def get_script_answer(self,card_id,question_text):
        try:
            #scriptのパスの生成
            current_dir = Path(__file__).resolve().parent
            script_path = current_dir / "source_data" / "script" / f"c{card_id}.lua"
            if not os.path.isfile(script_path):
                #通常モンスターとかでスクリプトがない場合
                return AnswerValue.NO.value
            #scriptから質問の回答を判断
            with open(script_path, "r", encoding="utf-8") as file:
                lua_text = file.read()
                for script_text in self.script_json[question_text]:
                    #質問文から特定のscriptの取り出し
                    if re.search(script_text,lua_text):
                        #scriptに特定の文字列が含まれているかどうか
                        return AnswerValue.YES.value
                #含まれていないのでFalse
                return AnswerValue.NO.value
            
        except Exception as e:
            print(f"get_answer関数 エラー: {e}")
            print(card_id,question_text)
            return None
            
    def get_all_question_script(self):
        try:
            #quesionsテーブルにあるSCRIPTの質問を返す
            self.cursor.execute("SELECT id,question_text FROM questions WHERE category = ?;",(QuestionCategory.SCRIPT.value,))
            return self.cursor.fetchall()
        except Exception as e:
            print(f"{CARD_QUESTION_ANSWER_DB_NAME} 読み込みエラー: {e}")
            
    def get_all_cards_id(self):
        try:
            #cardsテーブルにあるcard_id,nameを返す
            self.cursor.execute("SELECT card_id FROM cards;")
            return [id[0] for id in self.cursor.fetchall()]
        except Exception as e:
            print(f"{CARD_QUESTION_ANSWER_DB_NAME} 読み込みエラー: {e}")
                
    def delete_card(self,card_id):
        try:
            #card_idの削除
            self.cursor.execute(f'DELETE FROM cards WHERE card_id = ?;',(card_id,))
            self.cursor.execute(f'DELETE FROM answers WHERE card_id = ?;',(card_id,))
        except Exception as e:
            print(f"delete_cardエラー: {e}")
            print(card_id)
            
    def delete_question(self,question_id):
        try:
            #question_idの削除
            self.cursor.execute(f'DELETE FROM questions WHERE card_id = ?;',(question_id,))
            self.cursor.execute(f'DELETE FROM answers WHERE question_id = ?;',(question_id,))
        except Exception as e:
            print(f"delete_cardエラー: {e}")
            print(question_id)
                
    def populate_from_sources(self):
        try:
            total_start = time.perf_counter()
            log(f"output database: {self.db_path}")
            log(f"cards database: {self.carddb_path}")
            log(f"script json: {self.script_json_path}")
            log(f"cards question json: {self.cards_json_path}")
            log(f"name reading json: {self.name_reading_json_path}")

            #名前と読みの辞書を作成
            step_start = time.perf_counter()
            log("[start] build name reading map")
            name_to_reading = self.build_name_to_reading()
            log(f"[done]  build name reading map: {len(name_to_reading)} names ({format_elapsed(step_start)})")
            
            # cards.cdb からすべてのカードを読み込む処理
            step_start = time.perf_counter()
            before_cards = self.table_count("cards")
            before_answers = self.table_count("answers")
            log("[start] load cards")
            self.cdbcursor.execute("SELECT id, name,desc FROM texts;")
            card_rows = self.cdbcursor.fetchall()
            processed_cards = 0
            for card_id, card_name, card_desc in card_rows:
                self.cdbcursor.execute("SELECT setcode, type, atk, def, level, race, attribute FROM datas WHERE id = ?;",(card_id,))
                card_data = self.cdbcursor.fetchone()
                #読みを取得
                if card_name in name_to_reading:
                    reading = name_to_reading[card_name]
                else:
                    reading = None
                self.add_card(card_id, card_name,reading,card_desc,card_data[0],card_data[1],card_data[2],card_data[3],card_data[4],card_data[5],card_data[6])
                processed_cards += 1
            after_cards = self.table_count("cards")
            after_answers = self.table_count("answers")
            log(
                f"[done]  load cards: processed {processed_cards}, cards +{after_cards - before_cards}, "
                f"script answers +{after_answers - before_answers} ({format_elapsed(step_start)})"
            )

            # script_to_question.json から質問を追加
            step_start = time.perf_counter()
            before_questions = self.table_count("questions")
            before_answers = self.table_count("answers")
            log("[start] load script questions")
            script_question_count = 0
            for question_text in self.script_json:
                self.add_question(question_text, QuestionCategory.SCRIPT.value)
                script_question_count += 1
            after_questions = self.table_count("questions")
            after_answers = self.table_count("answers")
            log(
                f"[done]  load script questions: processed {script_question_count}, "
                f"questions +{after_questions - before_questions}, answers +{after_answers - before_answers} "
                f"({format_elapsed(step_start)})"
            )
            
            # database_to_question.json から質問を追加
            step_start = time.perf_counter()
            before_questions = self.table_count("questions")
            log("[start] load cards questions")
            cards_question_count = 0
            for question_text in self.cards_json:
                if isinstance(self.cards_json[question_text],dict):
                    question_data = self.cards_json[question_text]
                    self.add_question(
                        question_text,
                        QuestionCategory.CARDS.value,
                        question_data.get("query"),
                        question_data.get("unset_bit", 0),
                        question_data.get("new_state", 0),
                        question_data.get("condition"),
                    )
                else:
                    self.add_question(question_text, QuestionCategory.CARDS.value,self.cards_json[question_text])
                cards_question_count += 1
            after_questions = self.table_count("questions")
            log(
                f"[done]  load cards questions: processed {cards_question_count}, "
                f"questions +{after_questions - before_questions} ({format_elapsed(step_start)})"
            )

            log(
                f"total: cards {self.table_count('cards')}, questions {self.table_count('questions')}, "
                f"answers {self.table_count('answers')} ({format_elapsed(total_start)})"
            )

        except Exception as e:
            print(f"cards.cdb 読み込みエラー: {e}")

    def close(self):
        # SQLite接続を閉じる前に残りの変更をcommitする。
        self.conn.commit()
        self.cdbconn.commit()
        #閉じる
        self.cursor.close()
        self.conn.close()
        self.cdbcursor.close()
        self.cdbconn.close()


if __name__ == "__main__":
    start_time = time.perf_counter()
    #データベースの生成
    card_db = CardDb()
    #cards.cdbからデータを読み込む
    card_db.populate_from_sources()
    #閉じる
    card_db.close()
    log(f"done ({format_elapsed(start_time)})")

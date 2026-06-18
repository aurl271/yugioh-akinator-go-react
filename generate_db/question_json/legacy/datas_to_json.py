#cards.cdbのdatasからの質問を生成
import sqlite3
import os
import difflib
import sys
import io

#utf-8で出力
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

#データベースを開く
current_dir = os.path.dirname(os.path.abspath(__file__))
db_path = os.path.dirname(current_dir)
db_path = os.path.join(db_path, "cards.cdb")
conn = sqlite3.connect(db_path)
cursor = conn.cursor()

#攻撃力を取得
cursor.execute("SELECT DISTINCT atk FROM datas ORDER BY atk")
atk_values = [row[0] for row in cursor.fetchall()]

#守備力を取得
cursor.execute("SELECT DISTINCT def FROM datas WHERE type & 67108864 = 0 ORDER BY def")
def_values = [row[0] for row in cursor.fetchall()]

#攻撃力が特定の値かどうかの質問の生成
for atk in atk_values:
    if atk == -2:
        print(f'    "攻撃力が?のモンスターですか？(魔法・罠ならいいえを選択)":{{')
        print(f'        "query":"type & 1 != 0 AND atk = {atk}",')
        print(f'        "unset_bit":710,')
        print(f'        "new_state":769')
        print(f'    }},')
    else:
        print(f'    "攻撃力が{atk}のモンスターですか？(魔法・罠ならいいえを選択)":{{')
        print(f'        "query":"type & 1 != 0 AND atk = {atk}",')
        print(f'        "unset_bit":710,')
        print(f'        "new_state":769')
        print(f'    }},')

#範囲指定の質問   
for i in range(0, 4501, 500):
    print(f'    "攻撃力が{i}以上{i+500}以下のモンスターですか？(魔法・罠ならいいえを選択)":{{')
    print(f'        "query":"type & 1 != 0 AND {i} <= atk AND atk <= {i+500}",')
    print(f'        "unset_bit":966,')
    print(f'        "new_state":257')
    print(f'    }},')



#守備力が特定の値かどうかの質問の生成
for defence in def_values:
    if defence == -2:
        print(f'    "守備力が?のモンスターですか？(リンクモンスター・魔法・罠ならいいえを選択)":{{')
        print(f'        "query":"type & 67108864 = 0 AND type & 1 != 0 AND def = {defence}",')
        print(f'        "unset_bit":2278,')
        print(f'        "new_state":3073')
        print(f'    }},')
    else:
        print(f'    "守備力が{defence}のモンスターですか？(リンクモンスター・魔法・罠ならいいえを選択)":{{')
        print(f'        "query":"type & 67108864 = 0 AND type & 1 != 0 AND def = {defence}",')
        print(f'        "unset_bit":2278,')
        print(f'        "new_state":3073')
        print(f'    }},')

#範囲指定の質問  
for i in range(0, 4501, 500):
    print(f'    "守備力が{i}以上{i+500}以下のモンスターですか？(リンクモンスター・魔法・罠ならいいえを選択)":{{')
    print(f'        "query":"type & 67108864 = 0 AND type & 1 != 0 AND {i} <= def AND def <= {i+500}",')
    print(f'        "unset_bit":3302,')
    print(f'        "new_state":1025')
    print(f'    }},')


#レベルが特定の値かどうかの質問の生成
for level in range(1,14):
    print(f'    "レベル、ランク、リンクマーカーの数(未来龍王などはテキストに書いてあるレベル)が{level}のモンスターですか？(魔法・罠ならいいえを選択)":{{')
    print(f'        "query":"type & 1 != 0 AND level & 0xff = {level}",')
    print(f'        "unset_bit":198,')
    print(f'        "new_state":4097')
    print(f'    }},')

#スケールの質問
for scale in range(14):
    print(f'    "ペンデュラムスケールが{scale}のモンスターですか？(魔法・罠ならいいえを選択)":{{')
    print(f'        "query":"type & 16777216 != 0 AND (level>>16)&0xff  = {scale}",')
    print(f'        "unset_bit":198,')
    print(f'        "new_state":24577')
    print(f'    }},')
"""
#setcodeの取得
cursor.execute("SELECT DISTINCT setcode FROM datas WHERE setcode != 0 ORDER BY setcode")
setcode_values = [row[0] for row in cursor.fetchall()]

#手動で作ったsetcode
used_setcode = [2,7,11,12,23,25,27,30,31,35,37,38,43,52,54,56,58,59,68,69,70,82,83,85,86,97,111,113,115,123,126,127,137,147,149,156,
                157,163,164,165,170,172,173,186,191,196,198,207,219,220,221,223,226,229,234,242,243,268,273,274,277,281,283,284,290,
                291,298,301,308,317,320,325,336,340,345,347.352,356,378,381,385,390,392,393,395,400,401,406,407,411,412,418,419,
                430,432,442,444,448,453,717,4288,4316,4373,20602,4260113,281018559]

#手動で作ったsetcodeとテーマ名
category_names = [[2,'ジェネクス'],[11,'インフェルニティ'],[12,'エーリアン'],[23,'シンクロ'],[29,'コアキメイル'],[31,'ネオスペーシアン','Ｎ'],
                  [35,'Ｓｉｎ'],[43,'忍者'],[52,'宝玉'],[54,'マシンナーズ'],[56,'ライトロード'],[58,'リチュア'],
                  [59,'レッドアイズ','真紅眼'],[68,'代行者'],[69,'デーモン'],[70,'融合','フュージョン'],[82,'ガーディアン'],
                  [83,'セイクリッド'],[85,'フォトン'],[86,'甲虫装機'],[97,'忍法'],[111,'ヒロイック','Ｈ－Ｃ'],[113,'マドルチェ'],
                  [115,'エクシーズ','ＣＸ','レイ・ピアース'],[123,'ギャラクシー','銀河'],[126,'炎舞'],[127,'ホープ'],[147,'サイバー'],[149,'ＲＵＭ'],[156,'テラナイト','星因子','星輝士'],
                  [157,'シャドール','影依','神の写し身との接触','魂写しの同化'],[163,'スターダスト'],[164,'クリボー'],[165,'チェンジ','紋章変換'],
                  [170,'クリフォート'],[172,'ゴブリン','百鬼羅刹'],[173,'デストーイ','魔玩具'],[186,'ＲＲ','レイド・ラプターズ','Ｒ・Ｒ・Ｒ','レイダーズ・アンブレイカブル・マインド'],
                  [191,'霊使い'],[196,'セフィラ'],[198,'Ｅｍ'],[207,'カオス','混沌','','ヌメロニアス・ヌメロニア','ＣＨＡＯＳ','ＣＸ','ＣＮ'],
                  [219,'ファントム','幻影'],[220,'超量'],[221,'ブルーアイズ','青眼'],[223,'月光'],[226,'トラミッド'],[229,'サイファー','光波'],
                  [234,'クリストロン','水晶機巧'],[242,'ペンデュラム','軌跡の魔術師','奇跡の魔導剣士','ドラゴニックＰ','竜剣士マスター','竜剣士ラスター','竜魔王ベクター','竜魔王レクター'],
                  [243,'プレデター','捕食'],[268,'ジャックナイツ','機界騎士','宵星の騎士','明星の機械騎士','双穹の騎士アストラム'],
                  [273,'アームド・ドラゴン','武装竜'],[274,'トロイメア','夢幻転星イドリース','夢幻崩界イヴリース'],
                  [277,'閃刀','未来の柱－キアノス','智の賢者－ヒンメル','閃術兵器－.','慈愛の賢者－シエラ','エルロン','武の賢者－アーカス'],
                  [281,'サラマングレイト','転生炎獣','フュージョン・オブ・ファイア','フューリー・オブ・ファイア','ライジング・オブ・ファイア'],
                  [283,'オルフェゴール','宵星の機神','宵星の騎士'],[284,'サンダー・ドラゴン','雷龍'],[290,'ワルキューレ','戦乙女の戦車','運命の戦車','Ｗａｌｋｕｒｅｎ'],
                  [291,'ローズ'],[298,'エンディミオン','魔法都市の実験施設'],[301,'シムルグ'],[320,'アダマシア','魔救'],
                  [325,'ドラグマ','凶導','教導','白の枢機竜','烙印の命数','導きの聖女クエム'],[336,'マギストス','聖月の魔導士エンディミオン','聖魔の大賢者エンディミオン'],
                  [340,'ドライトロン','輝巧','竜儀巧'],[356,'デスピア','導きの聖女クエム','導きの聖女クエム'],[378,'スケアクロー','肆世壊'],[381,'ヴァリアンツ'],
                  [382,'ラビュリンス','白銀の城'],[385,'ティアラメンツ','壱世壊'],[393,'クシャトリラ','六世壊'],[400,'マナドゥム','伍世壊'],[407,'レシピ'],
                  [411,'ディアベル','蛇眼の大炎魔'],[412,'スネークアイ','蛇眼'],[430,'千年','ミレニアム'],[442,'メタル化'],[444,'アザミナ'],
                  [453,'リゼェネシス','再世'],[717,'アルトメギア','神芸'],[4316,'超量士'],[4373,'閃刀姫'],
                  [20602,'焔聖騎士']]

for setcode in setcode_values:
    #手動で作ったのはパス
    if setcode in used_setcode:
        continue
    
    #setcodeに該当するidを取得
    cursor.execute("SELECT DISTINCT id FROM datas WHERE setcode = ? ORDER BY id",(setcode,))
    id_values = [row[0] for row in cursor.fetchall()]
    
    cards_names = []
    for id in id_values:
        #setcodeに該当するカードを追加
        cursor.execute("SELECT DISTINCT name FROM texts WHERE id = ?",(id,))
        card_name = cursor.fetchone()
        if not card_name:
            continue
        cards_names.append(card_name[0])
    if len(cards_names) < 5:
        continue
    
    theme_name = cards_names[0]
    for name in cards_names:
        #名前の共通部分を取り出す
        matcher = difflib.SequenceMatcher(None, theme_name, name)
        match = matcher.find_longest_match(0, len(theme_name), 0, len(name))
        theme_name = theme_name[match.a: match.a + match.size]
    if len(theme_name) == 0:
        #共通部分がないならパス
        continue
    #テーマ名(共通部分)を追加
    category_names.append([setcode,theme_name])
    
for category in category_names:
    #テーマのsetcodeを追加
    setcodes = {category[0]}
    for category_name in category[1:]:
        #テーマ名を持っているカードを取り出す
        cursor.execute("SELECT id FROM texts WHERE name LIKE ?",('%'+category_name+'%',))
        id_values = [row[0] for row in cursor.fetchall()]
        for id in id_values:
            #setcodeを追加
            cursor.execute("SELECT setcode FROM datas WHERE id = ?",(id,))
            setcode = cursor.fetchone()[0]
            setcodes.add(setcode)
            
    #ソート
    sorted_setcodes = sorted(setcodes)
    query = f'"'
    for i in range(len(sorted_setcodes)):
        #queryを作成
        query = query + f'setcode = {sorted_setcodes[i]}'
        if i == len(sorted_setcodes) - 1:
            query = query + f'"'
        else:
            query = query + f' OR '
    
    query = f"    「{category[1]}」カード？(テキストにルール上「～」カードとして扱う場合も含む):" + query
    print(query)
"""
#閉じる
cursor.close()
conn.close()
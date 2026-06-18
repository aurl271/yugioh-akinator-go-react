# YugiohAkinator リファクタリング設計メモ

## 目的

現在の YugiohAkinator は Flask を使ったテンプレート一体型のWebアプリです。
今後は、エンジニアのアルバイトやインターン応募時に見せられる実績として、backend と frontend を明確に分けた構成へ作り直します。

このリファクタリングで見せたい技術要素は次の通りです。

- backend と frontend の責務分離
- REST API backend の実装
- React + TypeScript による型付き frontend
- SQL / PostgreSQL のテーブル設計
- repository 層を使ったDBアクセス設計
- Render と Cloudflare Pages を使った分離デプロイ
- 計算処理の高速化
- README や docs で設計意図を説明できる状態

## 現状

現在の実体は主に以下にあります。

- Flask起動: `YugiohAkinator/YugiohAkinator/run.py`
- ルーティング: `YugiohAkinator/YugiohAkinator/app/routes.py`
- 推測ロジック: `YugiohAkinator/YugiohAkinator/app/akinator.py`
- DB生成: `YugiohAkinator/YugiohAkinator/app/database/generate_database.py`
- 現在のDB: `YugiohAkinator/YugiohAkinator/app/database/CQA.db`
- 現在の画面: `YugiohAkinator/YugiohAkinator/app/templates/*.html`

現在の実装では、次の責務が同じアプリ内に混ざっています。

- HTTPルーティング
- HTMLテンプレート描画
- session管理
- SQLiteアクセス
- ゲーム進行管理
- 確率計算
- 次の質問の選択

今後はこれらを分けます。

## 採用する技術スタック

### Backend

Go を使います。

理由:

- Cに近い感覚がある
- コンパイル言語で、構造体・ポインタ・明示的なエラー処理がある
- Rustより最初の大きな作り直しを完走しやすい
- Web API、Docker、デプロイ、サービス設計を見せやすい
- 計算処理をメモリ上のデータ構造で高速化しやすい
- ポートフォリオとして説明しやすい

候補ライブラリ:

- まずは標準の `net/http`
- ルーティングを楽にしたい場合は `chi`
- PostgreSQL接続は `pgx`
- ローカル開発用に必要なら `godotenv`

### Frontend

React + TypeScript + Vite を使います。

理由:

- TypeScript は frontend 開発でよく使われる
- React は採用側にも伝わりやすい
- Vite は開発環境を作りやすく、ビルドも速い
- Cloudflare Pages にデプロイしやすい

### Database

最初から Supabase Postgres を使います。

理由:

- PostgreSQL は SQLite だけよりも実績として見せやすい
- Supabase の管理画面でデータを確認しやすい
- 将来的にカード追加申請、正解ログ、フィードバック保存、管理画面などを作りやすい
- Render のローカルファイル永続化に依存しなくてよい

重要な方針:

PostgreSQL は永続データ置き場として使い、プレイ中の計算で毎回SQLを投げる設計にはしません。

backend 起動時に PostgreSQL からカード・質問・回答データを読み込み、Go のメモリ上に推測用データ構造を作ります。

## デプロイ方針

デプロイ先は以下にします。

- Backend: Render Web Service
- Frontend: Cloudflare Pages
- Database: Supabase Postgres

Render 側の注意:

- backend は `0.0.0.0` でlistenする
- 無料プランでは一定時間アクセスがないとスリープする可能性がある
- 起動直後のAPIレスポンスが遅くなることがある
- 実行時に書き込んだファイルへ依存しない

Cloudflare Pages 側の注意:

- React アプリは静的ファイルとしてビルドする
- backend のURLは環境変数で管理する

例:

```env
VITE_API_BASE_URL=https://your-backend.onrender.com
```

frontend と backend のドメインが分かれるため、Go backend 側で CORS 設定が必要です。

許可するoriginの例:

- `http://localhost:5173`
- Cloudflare Pages のURL
- 将来使う独自ドメイン

## 目標ディレクトリ構成

最終的にはルート直下をこのような構成にします。

```text
YugiohAkinator/
  backend/
    cmd/
      server/
        main.go
    internal/
      api/
        game_handler.go
        card_handler.go
        health_handler.go
      config/
        config.go
      engine/
        engine.go
        scorer.go
        question_selector.go
      game/
        service.go
        session.go
      model/
        answer.go
        card.go
        question.go
      repository/
        postgres_repository.go
    migrations/
      001_create_cards.sql
      002_create_questions.sql
      003_create_answers.sql
    go.mod
    README.md

  frontend/
    src/
      api/
        client.ts
        game.ts
      components/
        AnswerButtons.tsx
        CandidateTable.tsx
        HistoryTable.tsx
        QuestionCard.tsx
      pages/
        GamePage.tsx
        ResultPage.tsx
      types/
        game.ts
      main.tsx
    package.json
    README.md

  docs/
    refactor-plan.md
    api-design.md
    data-update-guide.md
```

## Backend の責務

backend は以下を担当します。

- カード・質問・回答データの読み込み
- ゲーム状態の管理
- 候補カードのスコア計算
- 次の質問の選択
- JSON API の提供
- CORS設定
- health check

backend はHTMLを描画しません。

## Frontend の責務

frontend は以下を担当します。

- 現在の質問を表示する
- 回答ボタンを表示する
- 計算中のローディングを表示する
- 候補カード一覧を表示する
- 回答履歴を表示する
- 結果画面を表示する
- backend API を呼び出す

frontend は推測ロジックを持ちません。

## API設計

最初に作るAPIは以下です。

```text
GET  /api/health
POST /api/game/start
POST /api/game/answer
POST /api/game/revert
POST /api/game/reset
GET  /api/cards/search?q=...
```

ゲーム中のレスポンス例:

```json
{
  "question": "このカードはモンスターですか？",
  "candidates": [
    {
      "rank": 1,
      "name": "ブラック・マジシャン",
      "probability": 0.42
    }
  ],
  "history": [
    {
      "question": "魔法カードですか？",
      "answer": "no"
    }
  ],
  "isAnswer": false,
  "answerCard": null
}
```

回答値:

```text
yes
probably
unknown
probably_no
no
```

内部では数値に変換します。

```text
yes         -> 1.0
probably    -> 0.5
unknown     -> 0.0
probably_no -> -0.5
no          -> -1.0
```

## DB設計

最初は以下のテーブルを中心にします。

```sql
cards (
  card_id bigint primary key,
  name text not null,
  reading text,
  description text,
  setcode bigint not null,
  type bigint not null,
  atk integer not null,
  def integer not null,
  level integer not null,
  race bigint not null,
  attribute bigint not null
);

questions (
  id bigserial primary key,
  question_text text not null unique,
  category text not null,
  query_key text,
  unset_bit bigint not null default 0,
  new_state bigint not null default 0
);

answers (
  card_id bigint not null references cards(card_id),
  question_id bigint not null references questions(id),
  answer integer not null,
  primary key (card_id, question_id)
);
```

検討するインデックス:

```sql
create index idx_answers_question_id on answers(question_id);
create index idx_answers_card_id on answers(card_id);
create index idx_cards_name on cards(name);
create index idx_cards_reading on cards(reading);
create index idx_questions_category on questions(category);
```

現在の SQLite 版では `questions.query` にSQL条件のような文字列を保存しています。
PostgreSQL版では、この文字列をそのまま実行する設計はなるべく避けます。

将来的には以下のように、より安全な形へ寄せます。

- `query_key`
- `condition_type`
- `condition_value`
- Go側での条件判定

ただし、最初の移行では動かすことを優先し、段階的に改善します。

## 高速化方針

現在遅くなりやすい部分は以下です。

- リクエストごとにDBを読む
- カードIDと配列indexの対応を毎回作る
- 質問候補ごとにDBへ問い合わせる
- YESになるカード集合を毎回作る
- 過去回答から Akinator を毎回作り直す

新しいbackendでは以下の流れにします。

起動時:

1. PostgreSQL から全カードを読む
2. PostgreSQL から全質問を読む
3. PostgreSQL から全回答を読む
4. `cardID -> index` と `index -> cardID` を作る
5. `questionID -> yesCardIndexes` を作る
6. 推測計算用の配列を作る

プレイ中:

1. ユーザーの回答済み質問IDと回答値を保持する
2. メモリ上の配列でスコアを更新する
3. キャッシュ済み質問データから候補質問を選ぶ
4. エントロピーを計算する
5. JSONで結果を返す

説明文としては以下のように言えます。

> PostgreSQL は永続データの保存先として使い、Go backend 起動時にカード・質問データをメモリへ読み込みます。ゲーム中はDBへ繰り返し問い合わせず、キャッシュ済み配列を使って高速に推測計算を行います。

## 移行手順

一気に全部を書き換えず、小さく進めます。

### Phase 1: 設計と雛形

- この設計メモを置く
- Go backend の雛形を作る
- `/api/health` を作る
- React + TypeScript frontend の雛形を作る
- READMEに方針を書く

### Phase 2: 最小backend API

- Go backend から Supabase Postgres に接続する
- migration SQL を作る
- 現在のカード・質問・回答データを移行する
- `/api/game/start` を実装する

### Phase 3: 最小frontend

- `/api/game/start` を呼ぶ画面を作る
- 最初の質問を表示する
- frontend を Cloudflare Pages にデプロイする
- backend を Render にデプロイする

### Phase 4: ゲーム進行

- `/api/game/answer` を実装する
- 候補カードを返す
- 回答履歴を返す
- 戻る・リセットを実装する
- 既存Flaskテンプレートの役割を置き換える

### Phase 5: 高速化

- 起動時キャッシュを作る
- プレイ中のDB問い合わせを減らす
- 旧Flask版と新Go版の速度を比較する
- READMEに結果を書く

### Phase 6: データ管理

- `docs/data-update-guide.md` を作る
- データ移行コマンドを作る
- 必要になったら管理画面やCLIを作る

## READMEに書くと良いこと

最終的なREADMEには以下を書きます。

- アプリの概要
- なぜ作り直したか
- アーキテクチャ図
- 技術スタック
- API例
- DBスキーマ
- デプロイURL
- 高速化の工夫
- 苦労した点
- 今後の改善予定

良い説明文の例:

> Flask のテンプレート一体型アプリを、Go の REST API backend と React/TypeScript frontend に再設計しました。Supabase PostgreSQL を永続データベースとして使い、推測計算はGo backend起動時に構築したメモリキャッシュ上で行うことで高速化を目指しています。

## 最初の小さなゴール

最初の実装目標は小さくします。

1. Go backend を作る
2. `/api/health` で `ok` を返す
3. React frontend を作る
4. frontend から `/api/health` を呼んで表示する
5. backend を Render にデプロイする
6. frontend を Cloudflare Pages にデプロイする

ここまでできれば、backend/frontend分離と分離デプロイが成立します。


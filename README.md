# Yugioh Akinator

遊戯王カードを対象にした、アキネーター風の推理Webアプリです。

ユーザーが質問に回答していくと、バックエンドが候補カードの確率を更新し、一定以上の自信度になったカードを回答として提示します。Python/Flaskで作っていた旧版を、ポートフォリオとして説明しやすい構成にするため、GoバックエンドとReact/TypeScriptフロントエンドに分けて作り直しています。
参考にさせていただいた記事:https://qiita.com/tsukemono/items/2a18e5d307a978e8ab09

## 技術スタック

- Backend: Go
- Frontend: React + TypeScript + Vite
- UI: MUI
- Database: PostgreSQL
- DB生成補助: Python
- 想定デプロイ:
  - Backend: Render
  - Frontend: Cloudflare Pages
  - Database: Supabase Postgres

## 主な機能

- 回答履歴をもとにカード候補の確率を計算
- `yes`, `probably`, `unknown`, `probably_no`, `no` の5段階回答
- ミスを許容しやすい/しない推理設定

## ディレクトリ構成

```text
backend/       Goバックエンド
frontend/      React/TypeScriptフロントエンド
generate_db/   データベースの生成やPostgreSQL投入用データの生成をするPythonコード
```

## Backend

### 起動

`backend` に移動してから実行します。

```powershell
cd backend
go run ./cmd/server
```

`DATABASE_URL` が必要です。

```powershell
$env:DATABASE_URL="postgresql://USER:PASSWORD@HOST:PORT/DBNAME"
go run ./cmd/server
```

サーバーはデフォルトで `http://localhost:8080` で起動します。

### API

主なエンドポイントです。

```text
POST /api/game/start
POST /api/game/answer
POST /api/game/confirm
GET  /api/health
```

バックエンドは基本的にステートレスです。回答履歴はフロントエンドが保持し、リクエストごとにGo側が履歴を再適用して状態を復元します。

## Frontend

### 起動

```powershell
cd frontend
npm install
npm run dev
```

Viteの開発サーバーは通常 `http://localhost:5173` で起動します。

BackendのURLを変えたい場合は、環境変数 `VITE_API_BASE_URL` を使います。

```powershell
$env:VITE_API_BASE_URL="http://localhost:8080"
npm run dev
```

### ビルド

```powershell
npm run build
```

## 推理エンジンの概要

各カードは `score` を持ちます。`score` は、ユーザーの回答と「そのカードが正解だった場合の期待回答」のズレの累積です。

- `score` が小さいほど正解候補として強い
- 確率は `exp(-beta * score)` を正規化して計算
- 次の質問は、回答が分かれやすい質問をエントロピーで選択

高速化のため、次質問選択では全カードを毎回見るのではなく、カードを500枚に絞って評価しています。

- 初回: ランダムに500枚
- 2問目以降: `score` が低い上位500枚
//go:build integration

package repository

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	integrationCardAID       = 1001
	integrationCardBID       = 1002
	integrationQuestionAID   = 1
	integrationQuestionBID   = 2
	integrationExpectedYes   = 1
	integrationDatabaseName  = "akinator_test"
	integrationDatabaseUser  = "test"
	integrationDatabasePass  = "test"
	integrationPostgresImage = "postgres:16-alpine"
)

// TestRepositoriesWithPostgres は一時PostgreSQLを起動し、RepositoryのDB処理を独立したサブテストで確認する。
func TestRepositoriesWithPostgres(t *testing.T) {
	ctx, db := startIntegrationPostgres(t)
	tests := []struct {
		name string
		run  func(t *testing.T, ctx context.Context, db *sql.DB)
	}{
		{name: "connection", run: testPostgresConnection},
		{name: "load cards", run: testLoadCardsFromPostgres},
		{name: "load questions", run: testLoadQuestionsFromPostgres},
		{name: "reject invalid condition json", run: testLoadQuestionsRejectsInvalidConditionJSON},
		{name: "load answers", run: testLoadAnswersFromPostgres},
		{name: "load game data", run: testLoadGameDataFromPostgres},
		{name: "save game result", run: testSaveGameResultToPostgres},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetIntegrationDatabase(t, ctx, db)
			tt.run(t, ctx, db)
		})
	}
}

// startIntegrationPostgres はschemaとmigrationを適用した一時PostgreSQLを起動し、接続済みのsql.DBを返す。
func startIntegrationPostgres(t *testing.T) (context.Context, *sql.DB) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	t.Cleanup(cancel)
	schemaPath, migrationPath := integrationSQLPaths(t)

	container, err := postgres.Run(
		ctx,
		integrationPostgresImage,
		postgres.WithDatabase(integrationDatabaseName),
		postgres.WithUsername(integrationDatabaseUser),
		postgres.WithPassword(integrationDatabasePass),
		postgres.BasicWaitStrategies(),
		postgres.WithOrderedInitScripts(schemaPath, migrationPath),
	)
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("start PostgreSQL container: %v", err)
	}

	databaseURL, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("get PostgreSQL connection string: %v", err)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open PostgreSQL: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close PostgreSQL: %v", err)
		}
	})

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}

	return ctx, db
}

// integrationSQLPaths はこのファイルの場所を基準にschemaとmigrationの絶対パスを求める。
func integrationSQLPaths(t *testing.T) (string, string) {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("get integration test file path")
	}
	projectRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
	schemaPath := filepath.Join(projectRoot, "generate_db", "postgres", "schema.sql")
	migrationPath := filepath.Join(projectRoot, "backend", "migrations", "create_game_results.sql")

	for _, path := range []string{schemaPath, migrationPath} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("find integration SQL file %q: %v", path, err)
		}
	}

	return schemaPath, migrationPath
}

// resetIntegrationDatabase はサブテスト間でデータが混ざらないように全テーブルを空にする。
func resetIntegrationDatabase(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	_, err := db.ExecContext(
		ctx,
		`TRUNCATE answers, questions, cards, game_results RESTART IDENTITY CASCADE`,
	)
	if err != nil {
		t.Fatalf("reset integration database: %v", err)
	}
}

// insertIntegrationGameData はLoad系テストで使うカード2枚、質問2件、回答1件をDBへ登録する。
func insertIntegrationGameData(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	_, err := db.ExecContext(ctx, `
		INSERT INTO cards (
			id, card_id, name, reading, "desc", setcode, type,
			atk, def, level, race, attribute
		)
		VALUES
			(1, 1001, 'カードA', 'カードエー', 'このカードは特殊召喚できる。', 52, 10, 2500, 2100, 8, 1, 2),
			(2, 1002, 'カードB', NULL, NULL, 0, 2, 1000, 1000, 4, 1, 2)
	`)
	if err != nil {
		t.Fatalf("insert integration cards: %v", err)
	}

	conditionJSON := `{"logic":"and","conditions":[{"field":"atk","op":"between","min":2000,"max":3000}]}`
	_, err = db.ExecContext(ctx, `
		INSERT INTO questions (
			id, question_text, category, query, condition_json, unset_bit, new_state
		)
		VALUES
			(1, 'スクリプト質問', 0, 'legacy query', NULL, 0, 0),
			(2, '攻撃力が2000以上3000以下ですか？', 1, NULL, $1, 4, 8)
	`, conditionJSON)
	if err != nil {
		t.Fatalf("insert integration questions: %v", err)
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO answers (id, card_id, question_id, answer)
		VALUES (1, 1001, 1, 1)
	`)
	if err != nil {
		t.Fatalf("insert integration answers: %v", err)
	}
}

// testPostgresConnection はTestcontainersで起動したPostgreSQLへPingできることを確認する。
func testPostgresConnection(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}
}

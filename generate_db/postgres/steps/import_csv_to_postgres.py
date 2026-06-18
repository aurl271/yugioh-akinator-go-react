from __future__ import annotations

import argparse
import os
from pathlib import Path


BASE_DIR = Path(__file__).resolve().parents[2]
DEFAULT_SCHEMA_PATH = BASE_DIR / "postgres" / "schema.sql"
DEFAULT_CSV_DIR = BASE_DIR / "postgres" / "csv"

TABLES = {
    "cards": [
        "id",
        "card_id",
        "name",
        "reading",
        "desc",
        "setcode",
        "type",
        "atk",
        "def",
        "level",
        "race",
        "attribute",
    ],
    "questions": [
        "id",
        "question_text",
        "category",
        "query",
        "condition_json",
        "unset_bit",
        "new_state",
    ],
    "answers": [
        "id",
        "card_id",
        "question_id",
        "answer",
    ],
}


def quote_ident(name: str) -> str:
    return '"' + name.replace('"', '""') + '"'


def copy_csv(cursor, table_name: str, columns: list[str], csv_dir: Path) -> None:
    csv_path = csv_dir / f"{table_name}.csv"
    if not csv_path.exists():
        raise FileNotFoundError(f"CSV file not found: {csv_path}")

    column_sql = ", ".join(quote_ident(column) for column in columns)
    copy_sql = f"COPY {quote_ident(table_name)} ({column_sql}) FROM STDIN WITH (FORMAT csv, HEADER true)"

    with csv_path.open("r", encoding="utf-8", newline="") as file:
        with cursor.copy(copy_sql) as copy:
            while chunk := file.read(1024 * 1024):
                copy.write(chunk)


def sync_id_sequence(cursor, table_name: str) -> None:
    cursor.execute(
        f"""
        SELECT setval(
            pg_get_serial_sequence('{table_name}', 'id'),
            COALESCE((SELECT MAX(id) FROM {quote_ident(table_name)}), 1),
            (SELECT COUNT(*) > 0 FROM {quote_ident(table_name)})
        )
        """
    )


def print_row_counts(cursor) -> None:
    for table_name in TABLES:
        cursor.execute(f"SELECT COUNT(*) FROM {quote_ident(table_name)}")
        row_count = cursor.fetchone()[0]
        print(f"{table_name}: {row_count} rows")


def import_csv_to_postgres(database_url: str, schema_path: Path, csv_dir: Path) -> None:
    if not schema_path.exists():
        raise FileNotFoundError(f"schema.sql not found: {schema_path}")

    import psycopg

    schema_sql = schema_path.read_text(encoding="utf-8")

    with psycopg.connect(database_url) as conn:
        with conn.cursor() as cursor:
            cursor.execute(schema_sql)
            for table_name, columns in TABLES.items():
                copy_csv(cursor, table_name, columns, csv_dir)
                sync_id_sequence(cursor, table_name)
        conn.commit()

        with conn.cursor() as cursor:
            print_row_counts(cursor)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Import PostgreSQL CSV files generated from CQA.db.")
    parser.add_argument("--schema", type=Path, default=DEFAULT_SCHEMA_PATH)
    parser.add_argument("--csv-dir", type=Path, default=DEFAULT_CSV_DIR)
    parser.add_argument("--database-url")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    database_url = args.database_url or os.environ.get("DATABASE_URL")
    if not database_url:
        raise RuntimeError("DATABASE_URL is not set. Set it in your environment or pass --database-url.")
    import_csv_to_postgres(database_url, args.schema, args.csv_dir)


if __name__ == "__main__":
    main()

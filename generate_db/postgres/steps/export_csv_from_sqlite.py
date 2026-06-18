from __future__ import annotations

import argparse
import csv
import sqlite3
from pathlib import Path


BASE_DIR = Path(__file__).resolve().parents[2]
DEFAULT_CQA_DB = BASE_DIR / "output" / "CQA.db"
DEFAULT_OUTPUT_DIR = BASE_DIR / "postgres" / "csv"

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


def export_table(conn: sqlite3.Connection, table_name: str, columns: list[str], output_dir: Path) -> int:
    output_path = output_dir / f"{table_name}.csv"
    column_sql = ", ".join(columns)
    query = f"SELECT {column_sql} FROM {table_name} ORDER BY id"

    with output_path.open("w", encoding="utf-8", newline="") as file:
        writer = csv.writer(file)
        writer.writerow(columns)

        row_count = 0
        for row in conn.execute(query):
            writer.writerow(row)
            row_count += 1

    return row_count


def export_csv_from_sqlite(cqa_db: Path, output_dir: Path) -> None:
    if not cqa_db.exists():
        raise FileNotFoundError(f"CQA database not found: {cqa_db}")

    output_dir.mkdir(parents=True, exist_ok=True)

    with sqlite3.connect(cqa_db) as conn:
        for table_name, columns in TABLES.items():
            row_count = export_table(conn, table_name, columns, output_dir)
            print(f"{table_name}: {row_count} rows -> {output_dir / f'{table_name}.csv'}")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Export CQA.db tables to CSV files for PostgreSQL import.")
    parser.add_argument("--cqa-db", type=Path, default=DEFAULT_CQA_DB)
    parser.add_argument("--output-dir", type=Path, default=DEFAULT_OUTPUT_DIR)
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    export_csv_from_sqlite(args.cqa_db, args.output_dir)


if __name__ == "__main__":
    main()

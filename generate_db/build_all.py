from __future__ import annotations

import argparse
import subprocess
import sys
from pathlib import Path


BASE_DIR = Path(__file__).resolve().parent
PYTHON = sys.executable


def run_step(name: str, command: list[str]) -> None:
    """生成処理の各ステップを表示付きで実行する。"""
    print(f"\n== {name} ==", flush=True)
    subprocess.run(command, cwd=BASE_DIR.parent, check=True)


def generate_question_json() -> None:
    """cards.cdbやCardPool.jsonから、cards由来質問JSONを生成する。"""
    run_step(
        "generate question json",
        [
            PYTHON,
            str(BASE_DIR / "question_json" / "generate_question_json.py"),
            "--output",
            str(BASE_DIR / "source_data" / "json" / "generated_cards_to_question.json"),
        ],
    )


def generate_cqa() -> None:
    """カード・質問・回答をまとめたSQLiteのCQA.dbを生成する。"""
    run_step(
        "generate CQA.db",
        [PYTHON, str(BASE_DIR / "generate_database.py")],
    )


def prepare_readings(skip_scrape: bool, skip_cqa: bool, args: argparse.Namespace) -> None:
    """読み仮名の不足分を整理し、必要なら公式サイト取得用ステップへ渡す。"""
    readings_command = [
        PYTHON,
        str(BASE_DIR / "readings" / "prepare_readings.py"),
        "--limit",
        str(args.reading_limit),
        "--delay",
        str(args.reading_delay),
        "--delay-jitter",
        str(args.reading_delay_jitter),
    ]
    if skip_scrape:
        readings_command.append("--skip-scrape")
    if skip_cqa:
        readings_command.append("--skip-cqa")
    run_step("prepare readings", readings_command)


def apply_readings_to_cqa() -> None:
    """収集済みの読み仮名を、生成済みCQA.dbのcards.readingへ反映する。"""
    run_step(
        "apply readings to CQA.db",
        [PYTHON, str(BASE_DIR / "readings" / "steps" / "apply_readings_to_cqa.py")],
    )


def main() -> None:
    """generate_db全体の生成パイプラインを順番に実行する入口。"""
    parser = argparse.ArgumentParser(description="Build all generated database files.")
    parser.add_argument("--skip-readings", action="store_true")
    parser.add_argument("--scrape-readings", action="store_true")
    parser.add_argument("--reading-limit", type=int, default=10)
    parser.add_argument("--reading-delay", type=float, default=2.0)
    parser.add_argument("--reading-delay-jitter", type=float, default=0.5)
    args = parser.parse_args()

    run_step(
        "prepare cards.cdb",
        [PYTHON, str(BASE_DIR / "change_db" / "prepare_cards_cdb.py")],
    )
    run_step(
        "normalize CardPool.json",
        [PYTHON, str(BASE_DIR / "question_json" / "steps" / "normalize_card_pool.py")],
    )
    if not args.skip_readings:
        prepare_readings(skip_scrape=not args.scrape_readings, skip_cqa=True, args=args)
    generate_question_json()
    generate_cqa()
    if not args.skip_readings:
        apply_readings_to_cqa()
    print("\ndone")


if __name__ == "__main__":
    main()

from __future__ import annotations

import argparse
import subprocess
import sys
from pathlib import Path


BASE_DIR = Path(__file__).resolve().parent
PYTHON = sys.executable


def run_step(name: str, command: list[str]) -> None:
    print(f"\n== {name} ==", flush=True)
    subprocess.run(command, cwd=BASE_DIR.parent, check=True)


def generate_question_json() -> None:
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
    run_step(
        "generate CQA.db",
        [PYTHON, str(BASE_DIR / "generate_database.py")],
    )


def prepare_readings(skip_scrape: bool, args: argparse.Namespace) -> None:
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
    run_step("prepare readings", readings_command)


def main() -> None:
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
    generate_question_json()
    generate_cqa()
    if not args.skip_readings:
        prepare_readings(skip_scrape=not args.scrape_readings, args=args)
        if args.scrape_readings:
            # Scraping can add new reading initials/finals to official_readings_cache.json.
            # Regenerate questions and CQA once, then reapply collected readings.
            generate_question_json()
            generate_cqa()
            prepare_readings(skip_scrape=True, args=args)
    print("\ndone")


if __name__ == "__main__":
    main()

from __future__ import annotations

import argparse
import json
import sys
import time
from pathlib import Path
from typing import Any

from steps.datas_to_questions import build_datas_questions, build_legacy_setcode_questions
from steps.reading_to_questions import build_reading_questions


QuestionMap = dict[str, dict[str, Any]]


def log(message: str) -> None:
    print(message, file=sys.stderr)


def default_cards_db_path() -> Path:
    return Path(__file__).resolve().parents[1] / "source_data" / "cards.cdb"


def default_card_pool_json_path() -> Path:
    return Path(__file__).resolve().parents[1] / "source_data" / "json" / "CardPool.json"


def run_generation_step(name: str, questions: QuestionMap, builder) -> None:
    before_count = len(questions)
    start_time = time.perf_counter()
    log(f"[start] {name}")

    generated = builder()
    questions.update(generated)

    elapsed = time.perf_counter() - start_time
    added_count = len(questions) - before_count
    log(f"[done]  {name}: +{added_count} questions ({elapsed:.2f}s)")


def build_generated_questions(cards_db_path: Path, card_pool_json_path: Path) -> QuestionMap:
    questions: QuestionMap = {}
    run_generation_step("datas questions", questions, lambda: build_datas_questions(cards_db_path))
    run_generation_step("setcode questions", questions, lambda: build_legacy_setcode_questions(cards_db_path))
    run_generation_step("reading questions", questions, lambda: build_reading_questions(card_pool_json_path))
    return questions


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Generate cards_to_question JSON entries.")
    parser.add_argument("--cards-db", type=Path, default=default_cards_db_path())
    parser.add_argument("--card-pool-json", type=Path, default=default_card_pool_json_path())
    parser.add_argument("--output", type=Path, help="Output file path. Prints to stdout when omitted.")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    start_time = time.perf_counter()

    log(f"cards database: {args.cards_db}")
    log(f"card pool json: {args.card_pool_json}")

    questions = build_generated_questions(args.cards_db, args.card_pool_json)
    output = json.dumps(questions, ensure_ascii=False, indent=2)
    elapsed = time.perf_counter() - start_time

    if args.output:
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(output + "\n", encoding="utf-8")
        log(f"output: {args.output}")
        log(f"total: {len(questions)} questions ({elapsed:.2f}s)")
        return

    log(f"total: {len(questions)} questions ({elapsed:.2f}s)")
    print(output)


if __name__ == "__main__":
    main()

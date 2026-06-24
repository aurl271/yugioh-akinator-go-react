from __future__ import annotations

import argparse
import subprocess
import sys
from pathlib import Path


BASE_DIR = Path(__file__).resolve().parents[1]
SCRIPT_DIR = Path(__file__).resolve().parent
STEPS_DIR = SCRIPT_DIR / "steps"
PYTHON = sys.executable


def run_step(name: str, command: list[str]) -> None:
    """読み仮名準備の各ステップを表示付きで実行する。"""
    print(f"\n== {name} ==", flush=True)
    subprocess.run(command, cwd=BASE_DIR.parent, check=True)


def normalize_readings_files() -> None:
    """読み仮名関連JSONの表記をそろえ、差分が追いやすい状態にする。"""
    normalize_script = STEPS_DIR / "normalize_missing_readings.py"
    json_dir = BASE_DIR / "source_data" / "json"
    for path in (
        json_dir / "missing_readings.json",
        json_dir / "official_readings_cache.json",
        json_dir / "not_found_readings.json",
    ):
        run_step(
            f"normalize {path.name}",
            [PYTHON, str(normalize_script), "--input", str(path), "--output", str(path)],
        )


def apply_readings_to_cqa() -> None:
    """読み仮名JSONの内容をCQA.dbへ反映する。"""
    run_step("apply readings to CQA.db", [PYTHON, str(STEPS_DIR / "apply_readings_to_cqa.py")])


def main() -> None:
    """不足読み仮名の抽出・手動反映・公式取得を順番に実行する入口。"""
    parser = argparse.ArgumentParser(description="Prepare missing readings JSON.")
    parser.add_argument("--skip-scrape", action="store_true")
    parser.add_argument("--skip-cqa", action="store_true")
    parser.add_argument("--dry-run-scrape", action="store_true")
    parser.add_argument("--limit", type=int, default=10)
    parser.add_argument("--delay", type=float, default=2.0)
    parser.add_argument("--delay-jitter", type=float, default=0.5)
    args = parser.parse_args()

    run_step("remove manual readings from not_found", [PYTHON, str(STEPS_DIR / "remove_manual_from_not_found.py")])
    if not args.skip_cqa:
        run_step("export missing readings", [PYTHON, str(STEPS_DIR / "export_missing_readings.py")])
        run_step("apply manual readings", [PYTHON, str(STEPS_DIR / "apply_manual_readings.py")])
    normalize_readings_files()
    if not args.skip_cqa:
        apply_readings_to_cqa()

    if args.skip_scrape:
        print("\n== scrape official readings ==", flush=True)
        print("skip scrape: --skip-scrape is set", flush=True)
        return

    scrape_command = [
        PYTHON,
        str(STEPS_DIR / "scrape_official_readings.py"),
        "--limit",
        str(args.limit),
        "--delay",
        str(args.delay),
        "--delay-jitter",
        str(args.delay_jitter),
    ]
    if args.dry_run_scrape:
        scrape_command.append("--dry-run")
    run_step("scrape official readings", scrape_command)
    normalize_readings_files()
    if not args.dry_run_scrape and not args.skip_cqa:
        apply_readings_to_cqa()


if __name__ == "__main__":
    main()

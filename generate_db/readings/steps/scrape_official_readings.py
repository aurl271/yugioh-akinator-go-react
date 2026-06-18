from __future__ import annotations

import argparse
import html.parser
import json
import random
import sys
import time
import urllib.parse
import urllib.request
from pathlib import Path
from typing import Any


BASE_DIR = Path(__file__).resolve().parents[2]
DEFAULT_INPUT_PATH = BASE_DIR / "source_data" / "json" / "missing_readings.json"
DEFAULT_CACHE_PATH = BASE_DIR / "source_data" / "json" / "official_readings_cache.json"
DEFAULT_NOT_FOUND_PATH = BASE_DIR / "source_data" / "json" / "not_found_readings.json"
DEFAULT_OUTPUT_PATH = BASE_DIR / "source_data" / "json" / "missing_readings.json"
OFFICIAL_SEARCH_URL = "https://www.db.yugioh-card.com/yugiohdb/card_search.action"


def safe_print(message: str) -> None:
    encoding = sys.stdout.encoding or "utf-8"
    print(message.encode(encoding, errors="replace").decode(encoding, errors="replace"))


class CardRubyParser(html.parser.HTMLParser):
    def __init__(self) -> None:
        super().__init__()
        self.in_card_ruby = False
        self.in_rt = False
        self.card_ruby_values: list[str] = []
        self.rt_values: list[str] = []

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        tag_name = tag.lower()
        attrs_dict = dict(attrs)
        class_names = (attrs_dict.get("class") or "").split()

        if tag_name == "span" and "card_ruby" in class_names:
            self.in_card_ruby = True
        if tag_name == "rt":
            self.in_rt = True

    def handle_endtag(self, tag: str) -> None:
        tag_name = tag.lower()
        if tag_name == "span" and self.in_card_ruby:
            self.in_card_ruby = False
        if tag_name == "rt":
            self.in_rt = False

    def handle_data(self, data: str) -> None:
        value = data.strip()
        if not value:
            return
        if self.in_card_ruby:
            self.card_ruby_values.append(value)
        if self.in_rt:
            self.rt_values.append(value)


def load_json(path: Path, default: Any) -> Any:
    if not path.exists():
        return default
    with path.open("r", encoding="utf-8") as file:
        return json.load(file)


def save_json(path: Path, data: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def build_search_url(card_name: str) -> str:
    params = {
        "ope": "1",
        "rp": "10",
        "sort": "1",
        "keyword": card_name,
        "stype": "1",
        "othercon": "2",
        "releaseDStart": "1",
        "releaseMStart": "1",
        "releaseYStart": "1999",
    }
    return f"{OFFICIAL_SEARCH_URL}?{urllib.parse.urlencode(params)}"


def fetch_html(url: str, user_agent: str, timeout: int) -> str:
    request = urllib.request.Request(
        url,
        headers={
            "User-Agent": user_agent,
            "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
            "Accept-Language": "ja,en-US;q=0.8,en;q=0.6",
        },
    )
    with urllib.request.urlopen(request, timeout=timeout) as response:
        charset = response.headers.get_content_charset() or "utf-8"
        return response.read().decode(charset, errors="replace")


def extract_reading_from_html(html: str) -> str | None:
    parser = CardRubyParser()
    parser.feed(html)
    if parser.card_ruby_values:
        return parser.card_ruby_values[0]
    if parser.rt_values:
        return "".join(parser.rt_values)
    return None


def scrape_reading(card: dict[str, Any], user_agent: str, timeout: int) -> str | None:
    url = build_search_url(card["name"])
    html = fetch_html(url, user_agent, timeout)
    return extract_reading_from_html(html)


def apply_cache_to_missing_cards(missing_cards: list[dict[str, Any]], cache: dict[str, str]) -> int:
    updated = 0
    for card in missing_cards:
        card_id = str(card["card_id"])
        if card.get("reading"):
            continue
        reading = cache.get(card_id)
        if reading:
            card["reading"] = reading
            updated += 1
    return updated


def add_not_found_card(not_found_cards: list[dict[str, Any]], card: dict[str, Any]) -> None:
    card_id = str(card["card_id"])
    if any(str(item.get("card_id")) == card_id for item in not_found_cards):
        return
    not_found_cards.append(
        {
            "card_id": card["card_id"],
            "name": card["name"],
        }
    )


def remove_cards_by_id(cards: list[dict[str, Any]], card_ids: set[str]) -> list[dict[str, Any]]:
    return [card for card in cards if str(card["card_id"]) not in card_ids]


def wait_between_requests(delay: float, jitter: float) -> None:
    wait_seconds = delay
    if jitter > 0:
        wait_seconds = random.uniform(max(0, delay - jitter), delay + jitter)
    safe_print(f"wait: {wait_seconds:.2f}s")
    time.sleep(wait_seconds)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Fetch missing card readings from the official database.")
    parser.add_argument("--input", type=Path, default=DEFAULT_INPUT_PATH)
    parser.add_argument("--cache", type=Path, default=DEFAULT_CACHE_PATH)
    parser.add_argument("--not-found", type=Path, default=DEFAULT_NOT_FOUND_PATH)
    parser.add_argument("--output", type=Path, default=DEFAULT_OUTPUT_PATH)
    parser.add_argument("--limit", type=int, default=10, help="Maximum number of cards to fetch in one run.")
    parser.add_argument("--delay", type=float, default=2.0, help="Seconds to wait between requests.")
    parser.add_argument("--delay-jitter", type=float, default=0.5, help="Random seconds added/subtracted from delay.")
    parser.add_argument("--timeout", type=int, default=20)
    parser.add_argument(
        "--user-agent",
        default="YugiohAkinatorReadingCollector/0.1 (+local development; contact: set-your-contact)",
    )
    parser.add_argument("--dry-run", action="store_true", help="Print targets without network access.")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    missing_cards = load_json(args.input, [])
    cache = load_json(args.cache, {})
    not_found_cards = load_json(args.not_found, [])
    not_found_ids = {str(card.get("card_id")) for card in not_found_cards}
    current_not_found_ids: set[str] = set()

    fetched = 0
    for card in missing_cards:
        card_id = str(card["card_id"])
        if card.get("reading") or cache.get(card_id) or card_id in not_found_ids:
            continue

        safe_print(f"target: {card_id} {card['name']}")
        if args.dry_run:
            fetched += 1
            if fetched >= args.limit:
                break
            continue

        try:
            reading = scrape_reading(card, args.user_agent, args.timeout)
        except Exception as error:
            safe_print(f"failed: {card_id} {card['name']} ({error})")
            break

        if reading:
            cache[card_id] = reading
            safe_print(f"found: {card_id} {reading}")
        else:
            safe_print(f"not found: {card_id} {card['name']}")
            add_not_found_card(not_found_cards, card)
            current_not_found_ids.add(card_id)

        fetched += 1
        if not args.dry_run:
            save_json(args.cache, cache)
            save_json(args.not_found, not_found_cards)
        if fetched >= args.limit:
            break
        wait_between_requests(args.delay, args.delay_jitter)

    updated = apply_cache_to_missing_cards(missing_cards, cache)
    missing_cards = remove_cards_by_id(missing_cards, not_found_ids | current_not_found_ids)
    if not args.dry_run:
        save_json(args.cache, cache)
        save_json(args.not_found, not_found_cards)
        save_json(args.output, missing_cards)
    safe_print(f"cache entries: {len(cache)}")
    safe_print(f"not found entries: {len(not_found_cards)}")
    safe_print(f"updated missing readings: {updated}")
    safe_print(f"output: {args.output}")
    if args.dry_run:
        safe_print("dry run: files were not updated")


if __name__ == "__main__":
    main()

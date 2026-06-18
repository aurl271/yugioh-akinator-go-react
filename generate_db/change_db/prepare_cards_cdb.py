import argparse
from pathlib import Path

from steps.delete_ot2 import delete_ot2
from steps.delete_samename import delete_same_name_aliases
from steps.full_to_half import normalize_card_names
from steps.merge_database import merge_pre_release


# cards.cdb の前処理を実行する入口。
# 個別処理の中身は steps/ に分け、このファイルでは実行順だけを管理する。
def default_source_dir():
    return Path(__file__).resolve().parents[1] / "source_data"


def default_pre_release_db(source_dir):
    # 過去のファイル名ゆれに対応するため、両方の候補を見る。
    candidates = [
        source_dir / "pre-release.cdb",
        source_dir / "pre_release.cdb",
    ]
    for path in candidates:
        if path.exists():
            return path
    return candidates[0]


def main():
    source_dir = default_source_dir()

    parser = argparse.ArgumentParser(description="Prepare source_data/cards.cdb.")
    parser.add_argument("--cards-db", type=Path, default=source_dir / "cards.cdb")
    parser.add_argument("--pre-release-db", type=Path, default=default_pre_release_db(source_dir))
    parser.add_argument("--skip-merge", action="store_true")
    args = parser.parse_args()

    if not args.cards_db.exists():
        raise FileNotFoundError(f"cards database not found: {args.cards_db}")

    print(f"cards database: {args.cards_db}")
    print(f"pre-release database: {args.pre_release_db}")

    # 先に追加カードを取り込み、その後に不要データ削除と表記正規化を行う。
    if args.skip_merge:
        print("skip merge: --skip-merge is set")
    else:
        merge_pre_release(args.cards_db, args.pre_release_db)

    delete_ot2(args.cards_db)
    delete_same_name_aliases(args.cards_db)
    normalize_card_names(args.cards_db)
    print("done")


if __name__ == "__main__":
    main()

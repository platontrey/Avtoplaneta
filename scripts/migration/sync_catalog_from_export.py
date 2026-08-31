#!/usr/bin/env python3

"""Add missing legacy JoomShopping part name/category pairs to catalog.json."""

from __future__ import annotations

import argparse
import csv
import gzip
import json
import os
import re
import tempfile
import unicodedata
from pathlib import Path


def normalized(value: str) -> str:
    value = unicodedata.normalize("NFKC", value)
    return " ".join(value.split()).casefold()


def open_csv(path: Path):
    if path.suffix == ".gz":
        return gzip.open(path, "rt", encoding="utf-8", newline="")
    return path.open("r", encoding="utf-8", newline="")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("export", type=Path)
    parser.add_argument("catalog", type=Path)
    args = parser.parse_args()

    with args.catalog.open("r", encoding="utf-8") as stream:
        catalog = json.load(stream)

    allowed_categories = {
        item["name"] for item in catalog["part_form_categories"]
    }
    existing_pairs = {
        (normalized(item["name"]), normalized(item["category"]))
        for item in catalog["parts"]
    }

    candidates: dict[tuple[str, str], tuple[str, str]] = {}
    with open_csv(args.export) as stream:
        for row in csv.DictReader(stream):
            name = " ".join(row["name"].split())
            category = " ".join(row["category"].split())
            if not name or not category:
                continue
            if category not in allowed_categories:
                raise SystemExit(f"Unknown source category: {category!r}")
            key = (normalized(name), normalized(category))
            if key not in existing_pairs:
                candidates.setdefault(key, (name, category))

    max_id = 0
    for item in catalog["parts"]:
        match = re.fullmatch(r"part_(\d+)", item["id"])
        if match:
            max_id = max(max_id, int(match.group(1)))

    additions = sorted(
        candidates.values(),
        key=lambda value: (normalized(value[1]), normalized(value[0])),
    )
    for name, category in additions:
        max_id += 1
        catalog["parts"].append(
            {
                "id": f"part_{max_id:04d}",
                "name": name,
                "category": category,
                "quantity": 0,
                "price": 0,
            }
        )

    catalog["version"] = "2026-08-31.1"

    fd, temp_name = tempfile.mkstemp(
        prefix=args.catalog.name + ".", dir=str(args.catalog.parent)
    )
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as stream:
            json.dump(catalog, stream, ensure_ascii=False, indent=2)
            stream.write("\n")
        os.replace(temp_name, args.catalog)
    except Exception:
        os.unlink(temp_name)
        raise

    print(
        f"Added {len(additions)} missing part/category templates; "
        f"catalog now contains {len(catalog['parts'])} templates."
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

#!/usr/bin/env python3

import csv
import gzip
import json
import sys
from collections import Counter


def open_text(path):
    if path.endswith('.gz'):
        return gzip.open(path, 'rt', encoding='utf-8', newline='')
    return open(path, 'r', encoding='utf-8', newline='')


def main():
    if len(sys.argv) not in (2, 3):
        raise SystemExit('Usage: validate_joomshopping_export.py <parts.csv[.gz]> [photos.csv[.gz]]')

    metrics = Counter()
    with open_text(sys.argv[1]) as handle:
        reader = csv.DictReader(handle)
        required = {'id', 'name', 'category', 'quantity', 'location', 'address', 'photos'}
        if not required.issubset(set(reader.fieldnames or [])):
            raise SystemExit('Parts CSV is missing required columns')
        ids = set()
        for row in reader:
            metrics['parts'] += 1
            if row['id'] in ids:
                metrics['duplicate_part_ids'] += 1
            ids.add(row['id'])
            metrics['empty_names'] += not row['name'].strip()
            metrics['empty_categories'] += not row['category'].strip()
            metrics['numeric_addresses'] += row['address'].isdigit()
            metrics['unexpected_photos'] += row['photos'] != '[]'

    if metrics['parts'] == 0 or any(metrics[key] for key in (
        'duplicate_part_ids', 'empty_names', 'empty_categories', 'numeric_addresses', 'unexpected_photos'
    )):
        print(json.dumps(metrics, ensure_ascii=False, sort_keys=True))
        raise SystemExit('Parts export validation failed')

    if len(sys.argv) == 3:
        source_files = set()
        with open_text(sys.argv[2]) as handle:
            reader = csv.DictReader(handle)
            required = {'product_id', 'photo_url', 'ordering', 'source_file', 'size_bytes'}
            if not required.issubset(set(reader.fieldnames or [])):
                raise SystemExit('Photo manifest is missing required columns')
            for row in reader:
                metrics['photo_manifest_rows'] += 1
                if row['photo_url']:
                    metrics['photo_rows'] += 1
                    if row['source_file'] not in source_files:
                        metrics['photo_bytes'] += int(row['size_bytes'])
                        source_files.add(row['source_file'])
                if row['product_id'] not in ids:
                    metrics['orphan_photo_rows'] += 1
        if metrics['orphan_photo_rows']:
            print(json.dumps(metrics, ensure_ascii=False, sort_keys=True))
            raise SystemExit('Photo manifest contains products absent from parts export')

    print(json.dumps(metrics, ensure_ascii=False, sort_keys=True))


if __name__ == '__main__':
    main()

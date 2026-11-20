#!/usr/bin/env python3
"""
Инструмент статического анализа и автоисправления использования пакета `fmt` в Go-проекте Avtoplaneta.

Возможности:
  1. Детекция и автозамена статических `fmt.Errorf("...")` -> `errors.New("...")`
  2. Детекция неструктурированного логирования `fmt.Print`, `fmt.Printf`, `fmt.Println` (запрещены AGENTS.md)
  3. Детекция избыточных `fmt.Sprintf("...")` без спецификаторов формата

Использование:
  python3 scripts/audit_fmt.py [путь]              # Режим отчета (по умолчанию backend/)
  python3 scripts/audit_fmt.py --fix [путь]        # Режим автоисправления статических fmt.Errorf -> errors.New
"""

import os
import re
import sys
from collections import defaultdict

RE_FMT_ERRORF_STATIC = re.compile(r'fmt\.Errorf\(\s*("([^"\\]|\\.)*")\s*\)')
RE_FMT_SPRINTF_STATIC = re.compile(r'fmt\.Sprintf\(\s*("([^"\\]|\\.)*")\s*\)')
RE_FMT_PRINT = re.compile(r'\bfmt\.(Print|Printf|Println)\s*\(')

def has_format_specifier(literal_str: str) -> bool:
    """Проверяет наличие реальных спецификаторов формата (%) в строке Go."""
    s = literal_str.strip('"').replace('%%', '')
    return '%' in s

def audit_file(filepath: str, is_test: bool):
    findings = []
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            lines = f.readlines()
    except Exception:
        return findings

    for idx, line in enumerate(lines, start=1):
        stripped = line.strip()
        if stripped.startswith('//') or stripped.startswith('*'):
            continue

        # 1. fmt.Errorf со статичной строкой
        for match in RE_FMT_ERRORF_STATIC.finditer(line):
            lit = match.group(1)
            if not has_format_specifier(lit):
                findings.append({
                    'type': 'STATIC_ERRORF',
                    'line': idx,
                    'code': stripped,
                    'literal': lit,
                    'message': f'fmt.Errorf({lit}) без спецификаторов формата -> заменить на errors.New({lit})',
                    'is_test': is_test
                })

        # 2. fmt.Sprintf со статичной строкой
        for match in RE_FMT_SPRINTF_STATIC.finditer(line):
            lit = match.group(1)
            if not has_format_specifier(lit):
                findings.append({
                    'type': 'STATIC_SPRINTF',
                    'line': idx,
                    'code': stripped,
                    'literal': lit,
                    'message': f'fmt.Sprintf({lit}) без спецификаторов формата -> избыточен',
                    'is_test': is_test
                })

        # 3. fmt.Print / Println / Printf в продакшен-коде
        if RE_FMT_PRINT.search(line):
            findings.append({
                'type': 'FMT_PRINT_LOGGING',
                'line': idx,
                'code': stripped,
                'literal': None,
                'message': 'fmt.Print* используется вместо структурированного slog (нарушение AGENTS.md)',
                'is_test': is_test
            })

    return findings

def fix_static_errorf_in_file(filepath: str) -> int:
    """Заменяет статические fmt.Errorf на errors.New и добавляет импорт errors при необходимости."""
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    changed = 0

    def replacer(match):
        nonlocal changed
        lit = match.group(1)
        if not has_format_specifier(lit):
            changed += 1
            return f'errors.New({lit})'
        return match.group(0)

    new_content = RE_FMT_ERRORF_STATIC.sub(replacer, content)

    if changed > 0:
        # Проверяем наличие импорта "errors"
        if '"errors"' not in new_content:
            # Добавляем в блок import
            if 'import (\n' in new_content:
                new_content = new_content.replace('import (\n', 'import (\n\t"errors"\n', 1)
            elif 'import (' in new_content:
                new_content = re.sub(r'import\s*\(', 'import (\n\t"errors"', new_content, count=1)

        # Проверяем, остался ли fmt в файле
        if 'fmt.' not in new_content:
            # Удаляем импорт "fmt"
            new_content = re.sub(r'\t"fmt"\n', '', new_content)
            new_content = re.sub(r'"fmt"\n', '', new_content)

        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(new_content)

    return changed

def scan_path(target_path: str, fix_mode: bool = False):
    results = defaultdict(list)
    exclude_dirs = {'.git', 'vendor', 'gen', 'node_modules', '.idea', '.vscode'}
    fixed_count = 0

    if os.path.isfile(target_path):
        is_test = target_path.endswith('_test.go')
        findings = audit_file(target_path, is_test)
        if findings:
            results[target_path] = findings
        if fix_mode:
            fixed_count += fix_static_errorf_in_file(target_path)
        return results, fixed_count

    for dirpath, dirnames, filenames in os.walk(target_path):
        dirnames[:] = [d for d in dirnames if d not in exclude_dirs]

        for fname in filenames:
            if not fname.endswith('.go'):
                continue
            is_test = fname.endswith('_test.go')
            full_path = os.path.join(dirpath, fname)

            findings = audit_file(full_path, is_test)
            if findings:
                results[full_path] = findings

            if fix_mode:
                fixed_count += fix_static_errorf_in_file(full_path)

    return results, fixed_count

def main():
    args = sys.argv[1:]
    fix_mode = False
    if '--fix' in args:
        fix_mode = True
        args.remove('--fix')

    target = args[0] if args else 'backend'

    print(f"==================================================")
    print(f"  Аудит пакета `fmt` для: {target}")
    print(f"  Режим автоисправления: {'ВКЛЮЧЕН (--fix)' if fix_mode else 'ВЫКЛЮЧЕН'}")
    print(f"==================================================")

    results, fixed_count = scan_path(target, fix_mode=fix_mode)

    counts = defaultdict(int)
    for filepath, findings in sorted(results.items()):
        print(f"\n📂 {filepath}")
        for f in findings:
            counts[f['type']] += 1
            tag = " [TEST]" if f['is_test'] else ""
            print(f"  L{f['line']:<4}{tag} [{f['type']}]: {f['message']}")
            print(f"         > {f['code']}")

    print("\n" + "=" * 50)
    print("ИТОГ АУДИТА:")
    for k, v in counts.items():
        print(f"  - {k}: {v} шт.")
    total = sum(counts.values())
    print(f"Всего замечаний: {total}")
    if fix_mode:
        print(f"Автоматически исправлено (STATIC_ERRORF -> errors.New): {fixed_count} мест")

if __name__ == '__main__':
    main()

import sys
import os
from openpyxl import load_workbook

def parse_xlsx_to_parts(xlsx_file):
    wb = load_workbook(xlsx_file)
    ws = wb.active  # Первый лист

    parts = []

    for row in ws.iter_rows(min_row=2, values_only=True):  # Пропустить заголовок
        if not row or not row[1] or not row[4]:  # Пропустить пустые строки или без имени/категории
            continue

        # Столбцы: B=1, E=4, I=8, J=9, K=10, V=21
        name = row[1] if row[1] else ''
        category = row[4] if row[4] else ''
        description = row[21] if len(row) > 21 and row[21] else name  # Использовать V для описания, fallback на имя
        front_rear = row[8] if len(row) > 8 and row[8] else ''
        left_right = row[9] if len(row) > 9 and row[9] else ''
        top_bottom_raw = row[10] if len(row) > 10 and row[10] else ''
        if top_bottom_raw.lower() == 'верхняя':
            top_bottom = 'Верх'
        elif top_bottom_raw.lower() == 'нижняя':
            top_bottom = 'Низ'
        else:
            top_bottom = top_bottom_raw.capitalize()

        parts.append({
            'name': name,
            'category': category,
            'description': description,
            'front_rear': front_rear,
            'left_right': left_right,
            'top_bottom': top_bottom,
            'quantity': 0,
            'price': 1000
        })

    return parts

def generate_js_code(parts):
    lines = []
    for part in parts:
        line = f"       {{ name: \"{part['name']}\", category: \"{part['category']}\", description: \"{part['description']}\", quantity: {part['quantity']}, price: {part['price']}"
        if part['front_rear']:
            line += f", front_rear: \"{part['front_rear']}\""
        if part['left_right']:
            line += f", left_right: \"{part['left_right']}\""
        if part['top_bottom']:
            line += f", top_bottom: \"{part['top_bottom']}\""
        line += " },"
        lines.append(line)
    return "\n".join(lines)

def main():
    if len(sys.argv) < 2 or len(sys.argv) > 3:
        print("Использование: python generate_parts.py <xlsx_file> [output_file]")
        sys.exit(1)

    xlsx_file = sys.argv[1]
    output_file = sys.argv[2] if len(sys.argv) == 3 else None

    if not os.path.exists(xlsx_file):
        print(f"Файл {xlsx_file} не найден")
        sys.exit(1)

    parts = parse_xlsx_to_parts(xlsx_file)
    js_code = generate_js_code(parts)

    result = "// Сгенерированные данные запчастей\n" + js_code

    if output_file:
        with open(output_file, 'w', encoding='utf-8') as f:
            f.write(result)
        print(f"Результат сохранён в {output_file}")
    else:
        print(result)

if __name__ == "__main__":
    main()
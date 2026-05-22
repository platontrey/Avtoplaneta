import re

def get_realistic_price(name, category):
    name_lower = name.lower()
    category_lower = category.lower()

    # Болты
    if 'болт' in name_lower:
        return 100

    # Двигатель и крупные компоненты
    if 'двигатель' in name_lower or 'блок цилиндров' in name_lower:
        return 100000

    # Трансмиссия
    if category_lower == 'трансмиссия':
        if 'акпп' in name_lower or 'мкпп' in name_lower:
            return 50000 if 'акпп' in name_lower else 30000
        if 'дифференциал' in name_lower:
            return 20000
        return 5000

    # Подвеска
    if 'подвеска' in category_lower or 'подвеска' in name_lower:
        if 'амортизатор' in name_lower:
            return 2000
        if 'рычаг' in name_lower or 'кулак' in name_lower:
            return 3000
        return 1000

    # Тормозная система
    if 'тормоз' in category_lower or 'тормоз' in name_lower:
        if 'диск' in name_lower:
            return 1500
        if 'суппорт' in name_lower:
            return 2500
        return 500

    # Электрооснащение
    if category_lower == 'электрооснащение':
        if 'аккумулятор' in name_lower:
            return 5000
        if 'генератор' in name_lower or 'стартер' in name_lower:
            return 8000
        if 'блок управления' in name_lower:
            return 3000
        return 1000

    # Оптика
    if category_lower == 'оптика':
        if 'фара' in name_lower:
            return 3000
        return 500

    # Кузов
    if 'кузов' in category_lower:
        if 'дверь' in name_lower or 'крыло' in name_lower or 'капот' in name_lower:
            return 5000
        if 'стекло' in name_lower:
            return 2000
        return 1000

    # Система охлаждения
    if 'охлаждени' in category_lower:
        if 'радиатор' in name_lower:
            return 3000
        if 'помпа' in name_lower:
            return 2000
        return 500

    # Диски и шины
    if 'диски и шины' in category_lower:
        if 'шина' in name_lower:
            return 3000
        if 'диск' in name_lower:
            return 5000
        return 200

    # По умолчанию
    return 1000

def update_prices_in_file(file_path):
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()

    # Найти массив commonParts
    pattern = r'(\s+const commonParts: CommonPart\[\] = \[)(.*?)(\s+\];)'
    match = re.search(pattern, content, re.DOTALL)
    if not match:
        print("Массив commonParts не найден")
        return

    prefix = match.group(1)
    array_content = match.group(2)
    suffix = match.group(3)

    # Разделить на элементы
    items = re.findall(r'\{[^}]*\}', array_content)

    updated_items = []
    for item in items:
        # Извлечь name и category
        name_match = re.search(r"name:\s*\"([^\"]+)\"", item)
        category_match = re.search(r"category:\s*\"([^\"]+)\"", item)
        if name_match and category_match:
            name = name_match.group(1)
            category = category_match.group(1)
            price = get_realistic_price(name, category)
            # Заменить price
            item = re.sub(r'price:\s*\d+', f'price: {price}', item)
        updated_items.append(item)

    new_array_content = '\n'.join(updated_items)
    new_content = prefix + new_array_content + suffix

    # Заменить в файле
    content = re.sub(pattern, new_content, content, flags=re.DOTALL)

    with open(file_path, 'w', encoding='utf-8') as f:
        f.write(content)

    print("Цены обновлены")

if __name__ == "__main__":
    update_prices_in_file('../frontend/src/components/DefectReport.tsx')
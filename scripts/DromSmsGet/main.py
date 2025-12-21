import os
import json
import time
import requests
from bs4 import BeautifulSoup

# --- НАСТРОЙКИ ---
# Если у тебя остался файл drom_session.json от Playwright - скрипт возьмет куки оттуда.
# Если нет - вставь строку Cookie ниже вручную (инструкция далее).
SESSION_FILE = 'drom_session.json'

# User-Agent обязателен, чтобы не заблокировали
HEADERS = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
    'X-Requested-With': 'XMLHttpRequest', # Важно! Говорит сайту, что мы просим JSON
    'Accept': 'application/json, text/javascript, */*; q=0.01',
}

def load_cookies_from_playwright(session_file):
    """Преобразует куки из формата Playwright в формат requests"""
    if not os.path.exists(session_file):
        return None
    
    with open(session_file, 'r', encoding='utf-8') as f:
        data = json.load(f)
    
    cookies = {}
    # Playwright хранит куки в списке словарей внутри ключа "cookies"
    for c in data.get('cookies', []):
        cookies[c['name']] = c['value']
    return cookies

def clean_dialog_data(raw_json, dialog_id):
    """Очистка HTML из JSON (та же логика, что и раньше)"""
    interlocutor_name = raw_json.get('interlocutor', 'Неизвестный')
    html_content = raw_json.get('dialog', '')
    
    clean_messages = []
    
    if html_content:
        soup = BeautifulSoup(html_content, 'html.parser')
        msg_containers = soup.select('.bzr-dialog__msg-container')
        
        for container in msg_containers:
            try:
                msg_id = container.get('data-message-id')
                block = container.select_one('.bzr-dialog__message')
                if not block: continue
                
                text_el = block.select_one('.bzr-dialog__text')
                text = text_el.get_text(strip=True) if text_el else "[Вложение]"
                
                date_el = block.select_one('.bzr-dialog__message-dt')
                time_val = date_el.get_text(strip=True) if date_el else ""
                
                classes = block.get('class', [])
                if 'bzr-dialog__message_out' in classes:
                    author = "Я"
                    direction = "outgoing"
                else:
                    author = interlocutor_name
                    direction = "incoming"

                check_icon = block.select_one('.bzr-dialog__message-check')
                is_read = check_icon.get('data-state') == 'read' if check_icon else False
                
                clean_messages.append({
                    "id": msg_id,
                    "author": author,
                    "direction": direction,
                    "time": time_val,
                    "text": text,
                    "is_read": is_read
                })
            except: continue

    return {
        "dialog_id": dialog_id,
        "interlocutor": interlocutor_name,
        "messages_count": len(clean_messages),
        "messages": clean_messages
    }

def main():
    session = requests.Session()
    session.headers.update(HEADERS)
    
    # 1. Загрузка куки
    # Пробуем взять из файла Playwright
    cookies = load_cookies_from_playwright(SESSION_FILE)
    
    if cookies:
        print(f"Загружены куки из {SESSION_FILE}")
        session.cookies.update(cookies)
    else:
        print("Файл сессии не найден!")
        print("Пожалуйста, зайдите в браузер -> F12 -> Network.")
        print("Обновите страницу диалогов, нажмите на запрос 'inbox-list'.")
        print("Скопируйте всё содержимое заголовка 'Cookie' и вставьте сюда:")
        cookie_str = input("Cookie: ").strip()
        # Простой парсинг строки куки
        cookie_dict = {i.split('=')[0].strip(): i.split('=')[1].strip() for i in cookie_str.split(';') if '=' in i}
        session.cookies.update(cookie_dict)

    # 2. Получаем список диалогов (через API)
    print("\nПолучаю список диалогов...")
    try:
        url_list = 'https://my.drom.ru/personal/messaging/inbox-list?ajax=1&fromIndex=0&count=50&list=personal'
        resp = session.get(url_list)
        
        if resp.status_code != 200:
            print(f"Ошибка доступа! Код: {resp.status_code}. Возможно, куки протухли.")
            return

        # Ищем ID диалогов в ответе
        import re
        dialog_ids = re.findall(r'"id":\s*(\d{7,12})', resp.text)
        
        # Убираем дубликаты
        dialog_ids = list(set(dialog_ids))
        
        # Добавляем твой тестовый, если его нет
        if '1771417825' not in dialog_ids: dialog_ids.append('1771417825')
        
        print(f"Найдено диалогов: {len(dialog_ids)}")

    except Exception as e:
        print(f"Ошибка при получении списка: {e}")
        return

    # 3. Скачиваем каждый диалог
    for i, d_id in enumerate(dialog_ids, 1):
        print(f"[{i}/{len(dialog_ids)}] Скачиваю диалог {d_id}...", end=" ")
        
        try:
            # Тот самый API запрос
            url_view = f'https://my.drom.ru/personal/messaging/view?dialogId={d_id}&json=true&flat-layout=false&ajax=1'
            resp = session.get(url_view)
            
            if resp.status_code == 200:
                raw_data = resp.json()
                clean_data = clean_dialog_data(raw_data, d_id)
                
                filename = f"dialog_{d_id}.json"
                with open(filename, "w", encoding="utf-8") as f:
                    json.dump(clean_data, f, ensure_ascii=False, indent=4)
                
                print(f"OK ({clean_data['messages_count']} сообщ.)")
            else:
                print(f"Ошибка {resp.status_code}")
                
        except Exception as e:
            print(f"Сбой: {e}")
        
        # Пауза 0.5 сек, чтобы не забанили IP
        time.sleep(0.5)

    print("\nГотово! Браузер не использовался.")

if __name__ == '__main__':
    main()
import os
import json
import requests
import time

# --- НАСТРОЙКИ ---
SESSION_FILE = 'drom_session.json'

# Заголовки, чтобы притвориться браузером
HEADERS = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
    'X-Requested-With': 'XMLHttpRequest',
    'Origin': 'https://my.drom.ru',
    'Accept': 'application/json, text/javascript, */*; q=0.01',
}

def load_cookies(session_file):
    """Загрузка куки из файла Playwright"""
    if not os.path.exists(session_file):
        return None
    with open(session_file, 'r', encoding='utf-8') as f:
        data = json.load(f)
    cookies = {}
    for c in data.get('cookies', []):
        cookies[c['name']] = c['value']
    return cookies

def send_message(session, dialog_id, message_text):
    """Функция отправки сообщения"""
    
    # URL тот же, что и для просмотра, но метод POST
    url = 'https://my.drom.ru/personal/messaging/view'
    
    # Параметры URL (Query String)
    params = {
        'dialogId': dialog_id,
        'json': 'true',
        'ajax': '1',
        'flat-layout': 'false'
    }
    
    # Тело запроса (Form Data)
    payload = {
        'message': message_text,
        'post': 'Отправить' # Обязательное поле, эмулирует нажатие кнопки
    }
    
    # Важно: обновляем Referer под конкретный диалог, иначе 403 Forbidden
    session.headers.update({
        'Referer': f'https://my.drom.ru/personal/messaging-modal/dialog-{dialog_id}'
    })

    try:
        response = session.post(url, params=params, data=payload)
        
        if response.status_code == 200:
            # Дром в ответ присылает JSON с обновленным диалогом
            # Можно проверить, есть ли там наше сообщение, но достаточно статуса 200
            print(f"✅ Сообщение отправлено!")
            return True
        else:
            print(f"❌ Ошибка отправки: {response.status_code}")
            print("Ответ сервера:", response.text[:200])
            return False
            
    except Exception as e:
        print(f"Ошибка соединения: {e}")
        return False

def main():
    session = requests.Session()
    session.headers.update(HEADERS)
    
    # 1. Загрузка куки
    cookies = load_cookies(SESSION_FILE)
    if cookies:
        print(f"Загружены куки из {SESSION_FILE}")
        session.cookies.update(cookies)
    else:
        print("Файл сессии не найден! Введите куки вручную.")
        cookie_str = input("Cookie string: ").strip()
        cookie_dict = {i.split('=')[0].strip(): i.split('=')[1].strip() for i in cookie_str.split(';') if '=' in i}
        session.cookies.update(cookie_dict)

    # 2. Выбор диалога
    print("\n--- ОТПРАВКА СООБЩЕНИЙ ---")
    # Можно использовать твой ID для теста
    target_id = input("Введите ID диалога (Enter для 1771417825): ").strip()
    if not target_id:
        target_id = '1771417825'

    print(f"\nВыбран диалог: {target_id}")
    print("Пишите сообщения. Для выхода введите 'exit'.")
    print("-" * 30)

    # 3. Чат-луп
    while True:
        msg = input("Вы: ").strip()
        
        if msg.lower() in ['exit', 'quit', 'выход']:
            break
        
        if not msg:
            continue
            
        success = send_message(session, target_id, msg)
        
        if not success:
            print("Попробуйте обновить куки.")
            break

if __name__ == '__main__':
    main()
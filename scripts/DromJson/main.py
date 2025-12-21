import os
import json
import time
import sys
import subprocess
from bs4 import BeautifulSoup

# --- НАСТРОЙКИ ПУТЕЙ ---
# Определяем папку, где лежит этот скрипт
try:
    SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
except NameError:
    SCRIPT_DIR = os.getcwd()

# Файл сессии теперь лежит строго рядом со скриптом
AUTH_FILE = os.path.join(SCRIPT_DIR, 'drom_session.json')

# --- НАСТРОЙКИ ВХОДА ---
LOGIN = '79521847448'
PASSWORD = 'qewret123'

LIST_URL = 'https://my.drom.ru/personal/messaging-modal?switchPosition=dialogs'
MAX_SCROLLS = 5

# --- АВТО-УСТАНОВКА ---
try:
    from playwright.sync_api import sync_playwright
except ImportError:
    subprocess.check_call([sys.executable, "-m", "pip", "install", "playwright"])
    subprocess.check_call([sys.executable, "-m", "playwright", "install", "chromium"])
    from playwright.sync_api import sync_playwright

def parse_messages_bs4(html):
    """Оригинальный парсер сообщений"""
    soup = BeautifulSoup(html, 'html.parser')
    messages_data = []
    
    msg_blocks = soup.select('.bzr-dialog__message')
    if not msg_blocks: msg_blocks = soup.select('.b-message')

    for block in msg_blocks:
        try:
            text_el = block.select_one('.bzr-dialog__text, .b-message__text')
            text = text_el.get_text(strip=True) if text_el else "[Вложение]"
            date_el = block.select_one('.bzr-dialog__message-dt, .b-message__info')
            date = date_el.get_text(strip=True) if date_el else ""
            classes = " ".join(block.get('class', []))
            author = "Собеседник" if 'in' in classes else "Я"
            messages_data.append({'author': author, 'date': date, 'text': text})
        except: continue
    return messages_data

def perform_login(page):
    """Функция автоматического входа"""
    print("--- НАЧИНАЮ АВТОМАТИЧЕСКИЙ ВХОД ---")
    page.goto('https://my.drom.ru/sign')
    
    print(f"Ввожу логин: {LOGIN}")
    page.fill('input[name="sign"]', LOGIN)
    
    print("Ввожу пароль...")
    page.fill('input[name="password"]', PASSWORD)
    
    print("Нажимаю кнопку 'Войти'...")
    page.click('button[type="submit"]')
    
    print("\n" + "="*50)
    print("ВАЖНО: Если Дром просит СМС-код или Капчу - введите их в браузере!")
    print("Как только увидите, что вы вошли в кабинет - нажмите ENTER в этой консоли.")
    print("="*50 + "\n")
    input("Нажмите ENTER, чтобы продолжить...")
    
    print(f"Сохраняю сессию в файл: {AUTH_FILE}")
    page.context.storage_state(path=AUTH_FILE)

def collect_dialog_ids(page):
    """Оригинальный сбор ID диалогов"""
    print("--- Сбор списка диалогов ---")
    
    found_ids = set()
    
    # Перехватчик ответов
    def handle_response(response):
        try:
            if "application/json" in response.headers.get("content-type", ""):
                text = response.text()
                import re
                ids = re.findall(r'"id":\s*(\d{7,12})', text) # Ищем ID в JSON
                for i in ids: found_ids.add(i)
        except: pass

    page.on("response", handle_response)
    page.goto(LIST_URL)
    
    # Скроллим немного, чтобы инициировать загрузку
    for _ in range(5):
        page.mouse.wheel(0, 500)
        time.sleep(1)
        
    page.remove_listener("response", handle_response)
    
    # Фильтрация
    real_ids = [uid for uid in found_ids if len(uid) > 6]
    if '1771417825' not in real_ids: real_ids.append('1771417825') # Страховка
    
    print(f"Найдено диалогов: {len(real_ids)}")
    return real_ids

def main():
    with sync_playwright() as p:
        # ЗАПУСК БРАУЗЕРА
        browser = p.chromium.launch(headless=False) 
        
        # Проверяем, есть ли сохраненная сессия по ПОЛНОМУ ПУТИ
        if os.path.exists(AUTH_FILE):
            print(f"Найден файл сессии {AUTH_FILE}. Входим без пароля.")
            context = browser.new_context(storage_state=AUTH_FILE)
            page = context.new_page()
        else:
            print("Файл сессии не найден. Будем входить по логину/паролю.")
            context = browser.new_context()
            page = context.new_page()
            perform_login(page)

        # --- ОСНОВНАЯ ЛОГИКА ---
        
        # 1. Собираем ID (старым рабочим методом)
        dialog_ids = collect_dialog_ids(page)
        
        # 2. Парсим каждый диалог
        for i, d_id in enumerate(dialog_ids, 1):
            print(f"[{i}/{len(dialog_ids)}] Обработка диалога {d_id}...")
            url = f'https://my.drom.ru/personal/messaging-modal/dialog-{d_id}'
            
            try:
                page.goto(url)
                try: page.wait_for_selector('.bzr-dialog__message', timeout=5000)
                except: pass
                
                # Прокрутка
                for _ in range(MAX_SCROLLS):
                    if page.locator('.bzr-dialog__message').count() > 0:
                        page.locator('.bzr-dialog__message').first.scroll_into_view_if_needed()
                        page.keyboard.press("Home")
                        page.wait_for_timeout(1500)
                    else: break
                
                # Сохранение
                content = page.content()
                history = parse_messages_bs4(content)
                
                # ИЗМЕНЕНИЕ ЗДЕСЬ: используем SCRIPT_DIR для пути
                filename = os.path.join(SCRIPT_DIR, f"dialog_{d_id}.json")
                
                with open(filename, "w", encoding="utf-8") as f:
                    json.dump(history, f, ensure_ascii=False, indent=4)
                
                print(f"   -> Сохранено {len(history)} сообщ. в {filename}")
                
            except Exception as e:
                print(f"Ошибка: {e}")

        print("\nГотово! Сессия сохранена, в следующий раз пароль не потребуется.")
        browser.close()

if __name__ == '__main__':
    main()
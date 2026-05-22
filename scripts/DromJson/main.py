import os
import json
import time
import sys
import subprocess
from bs4 import BeautifulSoup

# --- НАСТРОЙКИ ПУТЕЙ ---
try:
    SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
except NameError:
    SCRIPT_DIR = os.getcwd()

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
    """Парсер сообщений"""
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

def get_interlocutor_name(page):
    """Пытается найти имя собеседника в заголовке диалога"""
    name = "Неизвестный"
    try:
        # Варианты селекторов заголовка на Дроме
        selectors = [
            '.bzr-dialog-header__title',   # Современный дизайн
            '.b-messaging-dialog__person', # Старый дизайн
            '.bzr-dialog-header a',        # Ссылка на профиль в шапке
            'div[class*="header"] a[href*="user"]' # Универсальный поиск
        ]
        
        for sel in selectors:
            if page.locator(sel).count() > 0:
                name = page.locator(sel).first.inner_text().strip()
                if name: break
    except Exception as e:
        print(f"Не удалось получить имя: {e}")
    
    return name

def perform_login(page):
    print("--- ВХОД ---")
    page.goto('https://my.drom.ru/sign')
    page.fill('input[name="sign"]', LOGIN)
    page.fill('input[name="password"]', PASSWORD)
    page.click('button[type="submit"]')
    print("\n" + "="*50)
    print("Если нужна капча - введите в браузере! После входа нажмите ENTER.")
    print("="*50 + "\n")
    input("Нажмите ENTER...")
    page.context.storage_state(path=AUTH_FILE)

def collect_dialog_ids(page):
    print("--- Сбор списка диалогов ---")
    found_ids = set()
    
    def handle_response(response):
        try:
            if "application/json" in response.headers.get("content-type", ""):
                text = response.text()
                import re
                ids = re.findall(r'"id":\s*(\d{7,12})', text)
                for i in ids: found_ids.add(i)
        except: pass

    page.on("response", handle_response)
    page.goto(LIST_URL)
    
    for _ in range(5):
        page.mouse.wheel(0, 500)
        time.sleep(1)
        
    page.remove_listener("response", handle_response)
    
    real_ids = [uid for uid in found_ids if len(uid) > 6]
    # Добавляем твой ID из примера для гарантии
    if '1771417825' not in real_ids: real_ids.append('1771417825')
    
    print(f"Найдено диалогов: {len(real_ids)}")
    return real_ids

def main():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True) 
        
        if os.path.exists(AUTH_FILE):
            print(f"Вход по сессии {AUTH_FILE}")
            context = browser.new_context(storage_state=AUTH_FILE)
            page = context.new_page()
        else:
            context = browser.new_context()
            page = context.new_page()
            perform_login(page)

        dialog_ids = collect_dialog_ids(page)
        
        for i, d_id in enumerate(dialog_ids, 1):
            print(f"[{i}/{len(dialog_ids)}] Обработка ID: {d_id}...")
            url = f'https://my.drom.ru/personal/messaging-modal/dialog-{d_id}'
            
            try:
                page.goto(url)
                
                # Ждем появления сообщений (это значит, что и заголовок прогрузился)
                try: 
                    page.wait_for_selector('.bzr-dialog__message', timeout=5000)
                except: 
                    pass
                
                # 1. Сначала собираем имя (пока мы наверху или до скролла)
                interlocutor = get_interlocutor_name(page)
                print(f"   Собеседник: {interlocutor}")

                # 2. Скроллим историю
                for _ in range(MAX_SCROLLS):
                    if page.locator('.bzr-dialog__message').count() > 0:
                        page.locator('.bzr-dialog__message').first.scroll_into_view_if_needed()
                        page.keyboard.press("Home")
                        page.wait_for_timeout(1000)
                    else: break
                
                # 3. Парсим сообщения
                content = page.content()
                history = parse_messages_bs4(content)
                
                # 4. Формируем красивый JSON
                final_data = {
                    "dialog_id": d_id,
                    "interlocutor_name": interlocutor,
                    "url": url,
                    "messages_count": len(history),
                    "messages": history
                }

                filename = os.path.join(SCRIPT_DIR, f"dialog_{d_id}.json")
                with open(filename, "w", encoding="utf-8") as f:
                    json.dump(final_data, f, ensure_ascii=False, indent=4)
                
                print(f"   -> Сохранено. Файл: {filename}")
                
            except Exception as e:
                print(f"Ошибка: {e}")

        browser.close()

if __name__ == '__main__':
    main()
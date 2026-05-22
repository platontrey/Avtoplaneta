import requests
import json
import time
import os
from bs4 import BeautifulSoup
from urllib.parse import urljoin

class IlcatsUniversalParser:
    def __init__(self, cookie_file="cookies.json", output_file="parsed_data.json"):
        self.session = requests.Session()
        self.session.headers.update({
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
            'Referer': 'https://www.ilcats.ru/'
        })
        self.output_file = output_file
        self.visited_urls = set()
        self.all_parts_data = []
        
        # Загрузка кук
        self.load_cookies(cookie_file)

    def load_cookies(self, filename):
        if not os.path.exists(filename):
            print(f"ZORG-CRITICAL: Файл {filename} не найден! Сначала запусти main.py")
            exit()
        
        with open(filename, "r", encoding="utf-8") as f:
            cookies = json.load(f)
            for c in cookies:
                self.session.cookies.set(c['name'], c['value'])
        print(f"ZORG-SYSTEM: Куки загружены. Двигатель прогрет.")

    def get_soup(self, url):
        """Безопасный запрос с проверкой защиты"""
        if url in self.visited_urls:
            return None
        self.visited_urls.add(url)

        # Задержка, чтобы не положить сервер и не словить бан
        time.sleep(0.4) 
        
        try:
            response = self.session.get(url)
        except Exception as e:
            print(f"ZORG-ERROR: Ошибка сети: {e}")
            return None

        # Проверка на капчу
        if "smartcaptcha" in response.text and "checkbox-iframe" in response.text:
            print("\n!!! ZORG-ALERT: ОБНАРУЖЕНА КАПЧА !!!")
            print(">>> Скрипт остановлен. Запусти main.py, обнови куки и перезапусти парсер.")
            print(f">>> Остановка на URL: {url}")
            self.save_data() # Сохраняем то, что успели
            exit()
            
        return BeautifulSoup(response.text, 'html.parser')

    def parse_parts_table(self, soup, url):
        """Парсинг конечной таблицы с запчастями"""
        table = soup.select_one("div.Info table")
        if not table:
            return 0

        current_group = "General"
        extracted_count = 0
        
        # Попытка вытащить название схемы из заголовка страницы
        page_title = soup.select_one("div.Top li:last-child span")
        scheme_name = page_title.get_text(strip=True) if page_title else "Unknown Scheme"

        rows = table.find_all("tr")
        for row in rows:
            # Заголовок подгруппы внутри таблицы
            header = row.find("th")
            if header and "colspan" in header.attrs:
                current_group = header.get_text(strip=True)
                continue

            # Данные запчасти
            number_div = row.select_one("div.number a")
            if number_div:
                code = number_div.get_text(strip=True)
                
                count_div = row.select_one("div.count")
                count = count_div.get_text(strip=True) if count_div else "1"
                
                desc_div = row.select_one("div.additionalDescription")
                note = desc_div.get_text(strip=True) if desc_div else ""
                
                # Формируем запись
                item = {
                    "scheme": scheme_name,
                    "subgroup": current_group,
                    "code": code,
                    "count": count,
                    "note": note,
                    "url": url
                }
                self.all_parts_data.append(item)
                extracted_count += 1
        
        return extracted_count

    def process_url(self, url, depth=0):
        """Рекурсивный обход: решает, идти глубже или парсить"""
        indent = "  " * depth
        
        # Определяем тип страницы по URL
        if "function=getParts" in url:
            # ЭТО КОНЕЧНАЯ СТРАНИЦА - ПАРСИМ
            print(f"{indent}ZORG-HARVEST: Сбор данных со схемы...")
            soup = self.get_soup(url)
            if soup:
                count = self.parse_parts_table(soup, url)
                print(f"{indent}>>> Извлечено {count} запчастей.")
        
        else:
            # ЭТО ПАПКА - ИЩЕМ ССЫЛКИ ВНУТРЬ
            print(f"{indent}ZORG-SCAN: Сканирование категории...")
            soup = self.get_soup(url)
            if not soup: return

            # Ищем ссылки в списках (Groups, SubGroups, Models)
            # Обычно это div.List или div.ModelList
            links = soup.select("div.List div.name a, div.ModelList div.name a")
            
            if not links:
                # Если ссылок нет, возможно, мы наткнулись на ImageArea (редкий случай верстки)
                # Попробуем распарсить как запчасти на всякий случай
                if soup.select("div.Info table"):
                     count = self.parse_parts_table(soup, url)
                     print(f"{indent}>>> (Fallback) Извлечено {count} запчастей.")
                return

            print(f"{indent}-> Найдено {len(links)} вложений.")
            
            for link in links:
                href = link['href']
                next_url = urljoin("https://www.ilcats.ru", href)
                
                # Фильтр: не идем назад, не идем на смену языка
                if "language=" in href and "function=" not in href: continue
                
                # РЕКУРСИЯ
                self.process_url(next_url, depth + 1)

    def save_data(self):
        with open(self.output_file, "w", encoding="utf-8") as f:
            json.dump(self.all_parts_data, f, indent=4, ensure_ascii=False)
        print(f"\nZORG-SYSTEM: Данные ({len(self.all_parts_data)} записей) сохранены в {self.output_file}")

if __name__ == "__main__":
    # --- НАСТРОЙКИ ---
    # Сюда вставь ссылку на МОДЕЛЬ (не на бренд, чтобы не ждать вечность)
    # Пример: Abarth 500 (2008-...)
    START_URL = "https://www.ilcats.ru/?brand=abarth&function=getGroups&model=500A&modelCode=150&cat=3R"
    
    parser = IlcatsUniversalParser(output_file="abarth_parts.json")
    
    try:
        print(f"ZORG-Ω: Начало операции. Цель: {START_URL}")
        parser.process_url(START_URL)
    except KeyboardInterrupt:
        print("\nZORG-STOP: Принудительная остановка пользователем.")
    
    parser.save_data()
    print("ZORG-Ω: Работа завершена.")
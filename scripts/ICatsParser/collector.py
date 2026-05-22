import requests
import json
import time
import os
import random
from bs4 import BeautifulSoup
from urllib.parse import urljoin, urlparse, parse_qsl, urlencode, urlunparse

class IlcatsVeritasSpider:
    def __init__(self, cookie_file="cookies.json", targets_file="targets.json", history_file="visited_folders.json"):
        self.session = requests.Session()
        self.session.headers.update({
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
            'Referer': 'https://www.ilcats.ru/',
            'Accept-Language': 'ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7'
        })
        self.targets_file = targets_file
        self.history_file = history_file
        
        self.load_cookies(cookie_file)
        
        # Память
        self.collected_links = set()
        self.visited_folders = set()
        
        self.load_state()

    def load_cookies(self, filename):
        if not os.path.exists(filename):
            print("ZORG-CRITICAL: Нет cookies.json! Сначала обнови ключи.")
            exit()
        with open(filename, "r", encoding="utf-8") as f:
            cookies = json.load(f)
            for c in cookies:
                self.session.cookies.set(c['name'], c['value'])

    def load_state(self):
        if os.path.exists(self.targets_file):
            try:
                with open(self.targets_file, "r", encoding="utf-8") as f:
                    self.collected_links = set(json.load(f))
                print(f"ZORG-MEMORY: Целей в базе: {len(self.collected_links)}")
            except: pass
            
        if os.path.exists(self.history_file):
            try:
                with open(self.history_file, "r", encoding="utf-8") as f:
                    self.visited_folders = set(json.load(f))
                print(f"ZORG-MEMORY: Пройдено папок: {len(self.visited_folders)}")
            except: pass

    def save_state(self):
        with open(self.targets_file, "w", encoding="utf-8") as f:
            json.dump(list(self.collected_links), f, indent=4)
        with open(self.history_file, "w", encoding="utf-8") as f:
            json.dump(list(self.visited_folders), f, indent=4)
        print("ZORG-SAVE: Состояние сохранено.")

    def normalize_url(self, url):
        parsed = urlparse(url)
        qs = parse_qsl(parsed.query)
        qs.sort(key=lambda x: x[0])
        encoded_qs = urlencode(qs)
        return urlunparse((parsed.scheme, parsed.netloc, parsed.path, parsed.params, encoded_qs, ''))

    def validate_content(self, soup, url):
        """
        ПРОВЕРКА НА ЛОЖЬ.
        Возвращает True, если страница похожа на настоящую.
        Возвращает False, если это фейк/бан/пустышка.
        """
        text_content = soup.get_text()
        
        # 1. Проверка на явные признаки бана
        if "smartcaptcha" in str(soup) and "checkbox-iframe" in str(soup):
            print("\n!!! ZORG-ALERT: CAPTCHA DETECTED !!!")
            return False
            
        # 2. Проверка структуры ilcats
        # Настоящая страница каталога ВСЕГДА имеет списки или картинки
        has_list = soup.select("div.List") or soup.select("div.ModelList") or soup.select("div.Tile")
        has_image = soup.select("div.ImageArea") or soup.select("div.Info table") # Для конечных схем
        
        if not has_list and not has_image:
            print(f"\nZORG-WARNING: Страница {url} подозрительно пуста! (Soft Ban?)")
            # Проверим, может там сообщение об ошибке текстом
            if "Error" in text_content or "Ошибка" in text_content:
                print(f"Текст ошибки: {text_content[:50]}...")
            return False

        # 3. Проверка на "циклический редирект" (когда список есть, но он пуст или ведет на главную)
        # Если мы ожидаем список, но ссылок внутри 0 - это подозрительно
        if has_list:
            links = soup.select("div.List div.name a, div.ModelList div.name a, div.Tile a")
            if len(links) == 0:
                print("\nZORG-WARNING: Контейнер списка есть, но ссылок в нем 0!")
                return False

        return True

    def get_soup(self, url):
        clean_url = self.normalize_url(url)
        if clean_url in self.visited_folders:
            return "SKIPPED"
        
        # Рандомная задержка
        time.sleep(random.uniform(0.5, 1.2)) 
        
        try:
            response = self.session.get(url)
            
            # Если сервер вернул не 200 - это сразу ошибка
            if response.status_code != 200:
                print(f"ZORG-ERROR: HTTP {response.status_code}")
                return None

            soup = BeautifulSoup(response.text, 'html.parser')
            
            # --- ГЛАВНАЯ ПРОВЕРКА ---
            if not self.validate_content(soup, url):
                print(">>> ZORG-SYSTEM: Обнаружена подмена данных! Остановка защиты.")
                print(">>> ДЕЙСТВИЕ: Смени IP, обнови куки и удали visited_folders.json (если он мал).")
                self.save_state()
                exit() # ЖЕСТКИЙ ВЫХОД
            
            # Если проверка пройдена - помечаем как посещенную
            self.visited_folders.add(clean_url)
            return soup
            
        except Exception as e:
            print(f"ZORG-ERROR: {e}")
            return None

    def crawl(self, url, depth=0):
        if depth > 20: return
        indent = " " * depth

        res = self.get_soup(url)
        if res == "SKIPPED": return
        if not res: return
        
        soup = res
        print(f"{indent}> Scan: ...{url[-50:]}")

        links = soup.select("div.List div.name a, div.ModelList div.name a, div.Tile a")
        
        # Сохранение каждые 20 папок
        if len(self.visited_folders) % 20 == 0:
            self.save_state()

        if not links: return

        for link in links:
            href = link['href']
            full_url = urljoin("https://www.ilcats.ru", href)
            
            if any(x in href for x in ["language=", "sort=", "VinAction="]): continue
            
            # 1. СХЕМА
            if "function=getParts" in href:
                clean = self.normalize_url(full_url)
                if clean not in self.collected_links:
                    print(f"{indent}  [+] FOUND: {link.get_text(strip=True)}")
                    self.collected_links.add(clean)
                continue
            
            # 2. ПАПКА
            # Добавил проверку: ссылка должна вести ВНУТРЬ каталога, а не на главную
            if "?" in href and "function=" in href or "market=" in href:
                 self.crawl(full_url, depth + 1)

if __name__ == "__main__":
    # Сбрось историю, если подозреваешь, что она отравлена:
    # if os.path.exists("visited_folders.json"): os.remove("visited_folders.json")
    
    spider = IlcatsVeritasSpider()
    START_URL = "https://www.ilcats.ru/?brand=bmw&function=getModels&market=RUS&catalog=VT"
    
    try:
        spider.crawl(START_URL)
    except KeyboardInterrupt:
        print("\nZORG-STOP.")
        spider.save_state()
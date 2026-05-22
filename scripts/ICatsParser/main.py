from seleniumbase import Driver
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.common.action_chains import ActionChains
import time
import random
import json

class YandexSlayer:
    def __init__(self):
        print("ZORG-Ω: Активация протокола 'GLOBAL ACCESS'...")
        self.driver = Driver(uc=True, headless=False)
        self.wait = WebDriverWait(self.driver, 20)

    def human_sleep(self, min_t=0.5, max_t=1.5):
        time.sleep(random.uniform(min_t, max_t))

    def quantum_drag(self, element, distance):
        """Плавный сдвиг с овердрайвом"""
        action = ActionChains(self.driver)
        action.click_and_hold(element)
        
        target_distance = distance + 50 # Перелет
        steps = random.randint(15, 22)
        total_time = random.uniform(0.3, 0.5)
        
        current_x_int = 0
        
        for i in range(steps):
            t = (i + 1) / steps
            ease = 1 - pow(1 - t, 3) # Cubic Ease Out
            
            exact_next_x = target_distance * ease
            next_x_int = int(round(exact_next_x))
            
            delta_x = next_x_int - current_x_int
            delta_y = random.randint(-1, 1) if i % 3 == 0 else 0
            
            if delta_x != 0 or delta_y != 0:
                action.move_by_offset(delta_x, delta_y)
                action.pause(total_time / steps)
            
            current_x_int += delta_x
            
        action.pause(0.2)
        action.release()
        action.perform()

    def get_access(self, url):
        print(f"ZORG-Ω: Вход на {url}")
        self.driver.get(url)
        self.human_sleep(4, 6)

        try:
            # 1. Проверяем, есть ли капча вообще
            print("ZORG-Ω: Сканирование на наличие угроз...")
            try:
                iframe = self.wait.until(EC.presence_of_element_located(
                    (By.CSS_SELECTOR, "iframe[data-testid='checkbox-iframe']")
                ))
                print("ZORG-Ω: Капча обнаружена. Начинаю взлом.")
                self.driver.switch_to.frame(iframe)
                self.human_sleep(1, 2)
                
                # Ищем слайдер
                thumb = self.wait.until(EC.presence_of_element_located((By.CSS_SELECTOR, "[data-testid='thumb']")))
                track = self.driver.find_element(By.CLASS_NAME, "Track")
                
                # Считаем
                track_w = float(self.driver.execute_script("return arguments[0].getBoundingClientRect().width", track))
                thumb_w = float(self.driver.execute_script("return arguments[0].getBoundingClientRect().width", thumb))
                if track_w < 10: track_w = track.size['width']
                if thumb_w < 5: thumb_w = thumb.size['width']
                
                distance = track_w - thumb_w
                if distance <= 0: distance = 200

                # Двигаем
                self.quantum_drag(thumb, distance)
                
                self.driver.switch_to.default_content()
                self.human_sleep(3, 5)
                print("ZORG-Ω: Капча нейтрализована.")
                
            except Exception as e:
                print(f"ZORG-INFO: Капча не найдена или уже пройдена ({str(e)[:50]}).")

            # 2. Сохраняем куки
            cookies = self.driver.get_cookies()
            if len(cookies) > 0:
                with open("cookies.json", "w", encoding="utf-8") as f:
                    json.dump(cookies, f, indent=4)
                print(f"ZORG-Ω: УСПЕХ. Сохранено {len(cookies)} ключей доступа в 'cookies.json'.")
            else:
                print("ZORG-CRITICAL: Куки не получены.")

        except Exception as e:
            print(f"ZORG-FATAL: {e}")
        finally:
            self.driver.quit()

if __name__ == "__main__":
    bot = YandexSlayer()
    # Идем на главную, чтобы получить глобальный доступ
    bot.get_access("https://www.ilcats.ru/")
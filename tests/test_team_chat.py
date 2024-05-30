import time
import pytest
import random
import string
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.common.action_chains import ActionChains

# import ipdb; ipdb.set_trace()

@pytest.fixture()
def chrome_driver():
    driver = webdriver.Chrome()
    yield driver
    driver.quit()

@pytest.fixture()
def main_page(chrome_driver):
    chrome_driver.get("http://localhost:3000/channels")

@pytest.fixture()
def auth(chrome_driver, main_page):
    username_input = chrome_driver.find_element(By.XPATH, "//input[@id='username']")
    password_input = chrome_driver.find_element(By.XPATH, "//input[@id='password']")
    login_button = chrome_driver.find_element(By.XPATH, "//input[@name='login']")
    # Вводим данные в поля
    username_input.send_keys("roma")
    password_input.send_keys("roma")
    login_button.click()

def test_search_employees_and_send_direct_message(chrome_driver,main_page, auth):
    # Находим элемент поиска и вводим текст "misha"
    search_input = chrome_driver.find_element(By.XPATH, "//input[@placeholder='Search']")
    search_input.send_keys("misha")
    
    # Ожидаем и нажимаем на результат поиска с текстом "misha myadelets"
    search_result = WebDriverWait(chrome_driver, 10).until(
        EC.element_to_be_clickable((By.XPATH, "//p[text()='misha myadelets']"))
    )
    search_result.click() # после нажатия попадаем в личный чат с сотрудником 
    
    # Находим элемент для ввода сообщения и вводим текст "Hello, Misha"
    message_input = chrome_driver.find_element(By.XPATH, "//input[@placeholder='Type a message...']")
    message_input.send_keys("Hello, Misha")
    
    # Нажимаем Enter, чтобы отправить сообщение
    message_input.send_keys("\n")
    
    # Проверяем наличие отправленного сообщения с текстом "Hello"
    sent_message = WebDriverWait(chrome_driver, 10).until(
        EC.presence_of_element_located((By.XPATH, "//div[text()='Hello, Misha']"))
    )
    
    assert sent_message is not None

# @pytest.mark.skip
def test_auth(chrome_driver, auth):
    # Проверяем, что произошел переход на нужный URL
    WebDriverWait(chrome_driver, 10).until(EC.url_to_be("http://localhost:3000/channels"))
    assert chrome_driver.current_url == "http://localhost:3000/channels"

# @pytest.mark.skip
def test_create_group_channel(chrome_driver, auth):
    # Функция для генерации случайной строки из 10 букв
    def random_string(length=10):
        letters = string.ascii_lowercase
        return ''.join(random.choice(letters) for i in range(length))
    
    # Находим все элементы с классом linktag и считаем их количество
    WebDriverWait(chrome_driver, 10).until(
        EC.presence_of_all_elements_located((By.XPATH, "//a[@class='linktag']"))
    )
    initial_channel_count = len(chrome_driver.find_elements(By.XPATH, "//a[@class='linktag']"))
    
    # Нажимаем на кнопку "Create Channel"
    create_channel_button = chrome_driver.find_element(By.XPATH, "//button[text()='Create Channel']")
    create_channel_button.click()
    
    # Вводим имя канала в поле "Channel Name"
    channel_name = "New Group Channel " + random_string(10)
    channel_name_input = chrome_driver.find_element(By.XPATH, "//input[@placeholder='Channel Name']")
    channel_name_input.clear()
    channel_name_input.send_keys(channel_name)
    
    # Нажимаем на кнопку "Create"
    create_button = chrome_driver.find_element(By.XPATH, "//button[text()='Create']")
    create_button.click()
    
    # Ждем и проверяем, что количество каналов увеличилось на 1
    WebDriverWait(chrome_driver, 10).until(
        EC.presence_of_all_elements_located((By.XPATH, "//a[@class='linktag']"))
    )
    new_channel_count = len(chrome_driver.find_elements(By.XPATH, "//a[@class='linktag']"))
    assert new_channel_count == initial_channel_count + 1
    
    # Нажимаем на новый канал
    new_channel_link = chrome_driver.find_element(By.XPATH, f"//p[text()='{channel_name}']")
    new_channel_link.click()
    
    # Проверяем, что заголовок чата содержит текст "New Group Channel"
    chat_header = WebDriverWait(chrome_driver, 10).until(
        EC.presence_of_element_located((By.XPATH, f"//span[text()='{channel_name}']"))
    )
    assert chat_header is not None
    
    # Проверяем, что элемент h3 содержит текст "Recipients—1"
    recipients_header = chrome_driver.find_element(By.XPATH, "//h3")
    assert "Recipients—1" in recipients_header.text

# @pytest.mark.skip
def test_add_user_to_channel(chrome_driver, auth):
    first_group_channel = WebDriverWait(chrome_driver, 10).until(
        EC.element_to_be_clickable((By.XPATH, "//p[contains(text(),'New Group Channel')]"))
    )
    first_group_channel.click()
    
    # Найти элемент с id="add_user_icon" и кликнуть по нему
    add_user_icon = WebDriverWait(chrome_driver, 10).until(
        EC.element_to_be_clickable((By.ID, "add_user_icon"))
    )
    add_user_icon.click()
    
    # Найти элемент с placeholder="username or role" и ввести "misha"
    username_input = WebDriverWait(chrome_driver, 10).until(
        EC.presence_of_element_located((By.XPATH, "//input[@placeholder='username or role']"))
    )
    username_input.send_keys("misha")
    
    # Найти кнопку с текстом "Add" и кликнуть по ней
    WebDriverWait(chrome_driver, 10).until(
        EC.presence_of_element_located((By.XPATH, "//button[text()='Add']"))
    )
    add_button = chrome_driver.find_element(By.XPATH, "//button[text()='Add']")
    add_button.click()
    
    # Проверить, что на странице появилось системное сообщение о добавлении нового участника
    system_message = WebDriverWait(chrome_driver, 10).until(
        EC.presence_of_element_located((By.XPATH, "//div[contains(text(),'added misha myadelets')]"))
    )
    
    assert system_message is not None

def test_set_reaction(chrome_driver, auth):
    # Кликаем по первому элементу, который найдется по XPath //p[contains(text(),'New Group Channel')]
    first_group_channel = WebDriverWait(chrome_driver, 10).until(
        EC.element_to_be_clickable((By.XPATH, "//p[contains(text(),'New Group Channel')]"))
    )
    first_group_channel.click()
    
    # Вводим сообщение "Hello everyone" и отправляем его
    message_input = chrome_driver.find_element(By.XPATH, "//input[@placeholder='Type a message...']")
    message_input.send_keys("Hello everyone")
    message_input.send_keys("\n")
    
    # Ожидаем появления отправленного сообщения
    sent_message = WebDriverWait(chrome_driver, 10).until(
        EC.presence_of_element_located((By.XPATH, "//div[text()='Hello everyone']"))
    )
    
    # Сохраняем количество реакций под сообщениями
    initial_reactions_count = len(chrome_driver.find_elements(By.XPATH, "//div[@class='flex gap-2 bg-sky-600 p-1 rounded-lg px-2']"))
    
    # Находим все сообщения и кликаем правой кнопкой мыши по последнему сообщению
    messages = chrome_driver.find_elements(By.XPATH, "//div[@class='relative w-full flex my-1 hover:bg-zinc-900']")
    last_message = messages[-1]
    actions = ActionChains(chrome_driver)
    actions.context_click(last_message).perform()
    
    # Находим по id реакцию сердца и кликаем на нее
    heart_reaction = WebDriverWait(chrome_driver, 10).until(
        EC.element_to_be_clickable((By.ID, "heart_reaction"))
    )
    heart_reaction.click()
    
    # Проверяем, что количество реакций под сообщениями увеличилось
    new_reactions_count = len(chrome_driver.find_elements(By.XPATH, "//div[@class='flex gap-2 bg-sky-600 p-1 rounded-lg px-2']"))
    assert new_reactions_count > initial_reactions_count
    
    

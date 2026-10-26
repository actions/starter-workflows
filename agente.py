import os
import time
import re
import smtplib
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from bs4 import BeautifulSoup
from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from webdriver_manager.chrome import ChromeDriverManager

# Leitura das variáveis do GitHub Secrets
EMAIL_ORIGEM = os.getenv("EMAIL_ORIGEM")
SENHA_APP = os.getenv("SENHA_APP")
EMAIL_DESTINO = os.getenv("EMAIL_DESTINO")

PALAVRAS_CHAVE = [
    r'\bcelular\b', r'\bsmartphone\b', r'\biphone\b', r'\bsamsung\b', r'\bxiaomi\b',
    r'\btv\b', r'\btelevisao\b', r'\btelevisor\b', r'\bsmart tv\b'
]

URL_EDITAIS = "https://www25.receita.fazenda.gov.br/sle-sociedade/portal/editais-disponiveis"

def inicializar_driver():
    options = webdriver.ChromeOptions()
    options.add_argument("--headless")
    options.add_argument("--no-sandbox")
    options.add_argument("--disable-dev-shm-usage")
    options.add_argument("--disable-gpu")
    options.add_argument("user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")
    
    return webdriver.Chrome(service=Service(ChromeDriverManager().install()), options=options)

def contem_palavras_chave(texto):
    texto_lower = texto.lower()
    return any(re.search(p, texto_lower) for p in PALAVRAS_CHAVE)

def enviar_email(itens):
    if not (EMAIL_ORIGEM and SENHA_APP and EMAIL_DESTINO):
        print("[!] Credenciais de e-mail ausentes. Pulando envio.")
        return

    msg = MIMEMultipart()
    msg["From"] = EMAIL_ORIGEM
    msg["To"] = EMAIL_DESTINO
    msg["Subject"] = f"Alerta Leilão Receita Federal: {len(itens)} lote(s) encontrado(s)"

    corpo = f"Foram encontrados {len(itens)} lote(s)/item(ns) nos editais disponíveis:\n\n"
    for idx, item in enumerate(itens, 1):
        corpo += f"--- Item {idx} ---\n{item}\n\n"
    corpo += f"\nAcesse o portal: {URL_EDITAIS}"

    msg.attach(MIMEText(corpo, "plain"))

    try:
        server = smtplib.SMTP("smtp.gmail.com", 587)
        server.starttls()
        server.login(EMAIL_ORIGEM, SENHA_APP)
        server.send_message(msg)
        server.quit()
        print(f"[+] Alerta enviado com sucesso para {EMAIL_DESTINO}")
    except Exception as e:
        print(f"[!] Erro ao enviar e-mail: {e}")

def buscar():
    driver = inicializar_driver()
    try:
        driver.get(URL_EDITAIS)
        wait = WebDriverWait(driver, 20)
        wait.until(EC.presence_of_element_located((By.XPATH, "//div[contains(@class, 'card') or contains(@class, 'edital') or contains(text(), 'Edital')]")))
        time.sleep(3)

        soup = BeautifulSoup(driver.page_source, 'html.parser')
        cards = soup.find_all(['div', 'tr'], class_=re.compile(r'(card|item|linha|row)', re.I))
        
        lotes = [c.get_text(separator=" ", strip=True) for c in cards if contem_palavras_chave(c.get_text())]

        print(f"[+] Itens encontrados: {len(lotes)}")
        if lotes:
            enviar_email(lotes)
    finally:
        driver.quit()

if __name__ == "__main__":
    buscar()

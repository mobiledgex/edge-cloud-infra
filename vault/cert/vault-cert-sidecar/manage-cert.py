#!/usr/bin/env python3

import argparse
from datetime import datetime
import logging
import os
import random
import time

from utils.vault import vault_login, vault_get_cert
from utils.certs import get_cert_fingerprint

LOG_LEVEL = logging.INFO
CERT_DIR = os.environ.get("CERT_DIR", "/etc/envoy/certs")
CERT_FILENAME = "envoyTlsCerts.crt"
KEY_FILENAME = "envoyTlsCerts.key"
CERT_DATE_DIRNAME = "..data"
CERT_DATA_LINK = os.path.join(CERT_DIR, CERT_DATE_DIRNAME)
CERT_LINK = os.path.join(CERT_DIR, CERT_FILENAME)
KEY_LINK = os.path.join(CERT_DIR, KEY_FILENAME)

def update_cert():
    domain = os.environ["CERT_DOMAIN"]
    logging.info(domain)

    vcert = vault_get_cert(vault_login(), domain)
    fp = get_cert_fingerprint(vcert["cert"])

    certfile = os.path.join(CERT_DIR, CERT_FILENAME)
    keyfile = os.path.join(CERT_DIR, KEY_FILENAME)

    try:
        with open(certfile, "r") as f:
            current_cert = f.read()
    except FileNotFoundError:
        current_cert = ""

    try:
        current_fp = get_cert_fingerprint(current_cert)
    except Exception as e:
        logging.error(e)
        current_fp = None

    if current_fp == fp:
        logging.info("Certificate still valid")
        return

    logging.info("Updating certificate")

    # Do the kubernetes-style symlink dance to update certs automatically
    # and in a way that envoy will recognise for hot-reload
    cert_dir = os.path.join(CERT_DIR,
                            datetime.now().strftime("..certs_%Y_%m_%d_%H_%M_%S.%f"))
    os.mkdir(cert_dir)

    with open(os.path.join(cert_dir, "envoyTlsCerts.crt"), "w") as f:
        f.write(vcert["cert"])

    with open(os.path.join(cert_dir, "envoyTlsCerts.key"), "w") as f:
        f.write(vcert["key"])

    tmplink = os.path.join(CERT_DIR, "..tmp")
    os.symlink(os.path.basename(cert_dir), tmplink)
    os.replace(tmplink, CERT_DATA_LINK)

    if not os.path.exists(CERT_LINK):
        os.symlink(os.path.join(CERT_DATE_DIRNAME, CERT_FILENAME), CERT_LINK)

    if not os.path.exists(KEY_LINK):
        os.symlink(os.path.join(CERT_DATE_DIRNAME, KEY_FILENAME), KEY_LINK)

def main():
    try:
        interval = int(os.environ["CERT_RENEW_INTERVAL"])
    except KeyError:
        interval = 60*60*12 # 12 hours
    logging.info(f"Cert renew check every {interval} seconds")

    start_time = time.time()
    while True:
        update_cert()
        time.sleep(interval - ((time.time() - start_time) % interval))

if __name__ == "__main__":
    logging.basicConfig(level=LOG_LEVEL, format="%(asctime)s [%(levelname)s] %(message)s")
    main()

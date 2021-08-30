#!/usr/bin/env python3

import argparse
import json
import logging
import os
import subprocess
from yaml import load, dump, Loader, Dumper

from utils.certs import get_cert_fingerprint
from utils.vault import vault_login, vault_get_cert, vault_cert_domain
from utils.kubernetes import kubectl, create_tls_secret, \
    get_tls_cert_from_secret, patch_tls_cert_in_secret

LOG_LEVEL = logging.INFO

def dump_config():
    config = {
        "configVersion": "v1",
        "schedule": [
            {
                "name": "Update certs every 12 hours",
                "crontab": "0 */12 * * *",
            },
        ],
    }
    if os.environ.get("DEBUG") == "true":
        config["schedule"].append({
            "name": "Debug mode: update certs every two minutes",
            "crontab": "*/2 * * * *",
        })

    print(json.dumps(config))

def update_all_vaultcerts():
    p = kubectl("get vaultcerts")
    vcs = load(p.stdout, Loader=Loader)
    cert_cache = {}

    token = vault_login()

    npatched = 0
    for vc in vcs["items"]:
        name = vc["metadata"]["name"]
        namespace = vc["metadata"]["namespace"]
        domains = vc["spec"]["domain"]
        secret = vc["spec"]["secretName"]

        vcdomain = vault_cert_domain(domains)
        if vcdomain not in cert_cache:
            vcert = vault_get_cert(token, domains)
            fingerprint = get_cert_fingerprint(vcert["cert"])
            cert_cache[vcdomain] = {
                "cert": vcert,
                "fingerprint": fingerprint,
            }
        cert = cert_cache[vcdomain]

        try:
            sec_cert = get_tls_cert_from_secret(secret, namespace)
            sec_cert_fp = get_cert_fingerprint(sec_cert)
        except subprocess.CalledProcessError:
            # VC present, but TLS secret absent; create secret
            create_tls_secret({"object": vc})
            sec_cert_fp = cert["fingerprint"]
            npatched += 1

        if sec_cert_fp != cert["fingerprint"]:
            patch_tls_cert_in_secret(secret, namespace, cert["cert"])
            npatched += 1

    logging.info(f"All certs up-to-date. {npatched} patched in this run.")

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--config", action="store_true")
    args = parser.parse_args()

    logging.basicConfig(level=LOG_LEVEL, format="[%(levelname)s] %(message)s")

    if args.config:
        dump_config()
    else:
        update_all_vaultcerts()

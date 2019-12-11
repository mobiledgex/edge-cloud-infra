#!/usr/bin/python

import requests
import sys
from urllib3.exceptions import InsecureRequestWarning

requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

port = sys.argv[1] if len(sys.argv) > 1 else 8200

try:
    r = requests.get("https://localhost:{port}/v1/sys/health".format(port=port),
                     verify=False)
    status = r.json()
    print("vault_status available=True,initialized={initialized},sealed={sealed},standby={standby}".format(
        **status))
except:
    print('vault_status available=False')

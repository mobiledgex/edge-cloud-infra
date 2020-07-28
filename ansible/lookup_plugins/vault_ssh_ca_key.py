from __future__ import (absolute_import, division, print_function)
__metaclass__ = type

DOCUMENTATION = """
lookup: vault_ssh_ca_key
short_description: return vault SSH CA key
description:
  - return vault SSH CA key
options:
  vault_addr:
    description: The vault to connect to
    default: The "vault_address" ansible variable
    required: False
"""

EXAMPLES="""
  - set_fact:
      ssh_ca_key: "{{ query('vault_ssh_ca_key') }}"
  - debug: var=ssh_ca_key
"""

from ansible.errors import AnsibleError, AnsibleParserError
from ansible.plugins.lookup import LookupBase
from ansible.utils.display import Display

try:
    import requests
    HAS_REQUESTS = True
except ImportError as e:
    HAS_REQUESTS = False

display = Display()

class LookupModule(LookupBase):

    def run(self, terms, variables=None, **kwargs):
        if variables is not None:
            self._templar.available_variables = variables
        myvars = getattr(self._templar, '_available_variables', {})

        vault_addr_key = 'vault_address'
        try:
            vault_addr = kwargs.get(vault_addr_key, myvars[vault_addr_key])
        except KeyError:
            raise AnsibleError("Could not find vault address variable: {0}".format(vault_addr_key))

        url = "{0}/v1/ssh/public_key".format(vault_addr)
        display.vv("SSH CA public key URL: {0}".format(url))
        r = requests.get(url)
        if r.status_code != requests.codes.ok:
            display.vvv("Request: {0}".format(r.request.__dict__))
            raise AnsibleError("Vault SSH CA key lookup return response code: {0}".format(r.status_code))
        display.vv("SSH CA key: {0}".format(r.text))

        return [ r.text ]

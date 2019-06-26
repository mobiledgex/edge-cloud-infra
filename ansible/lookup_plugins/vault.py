from __future__ import (absolute_import, division, print_function)
__metaclass__ = type

from ansible.errors import AnsibleError, AnsibleParserError
from ansible.plugins.lookup import LookupBase
from ansible.utils.display import Display

import json
import re

try:
    import requests
    HAS_REQUESTS = True
except ImportError as e:
    HAS_REQUESTS = False

display = Display()

class LookupModule(LookupBase):

    def run(self, terms, variables=None, **kwargs):
        if not HAS_REQUESTS:
            raise AnsibleError(
                'python requests package is required for vault lookups')

        if variables is not None:
            self._templar.available_variables = variables
        myvars = getattr(self._templar, '_available_variables', {})

        ret = []

        if len(terms) != 1:
            raise AnsibleError('vault lookup needs exactly one path argument')

        path = re.sub(r'^secret/(?:data/)?', 'secret/data/', terms[0])

        vault_addr_key = 'vault_address'
        try:
            vault_addr = kwargs.get(vault_addr_key, myvars[vault_addr_key])
        except KeyError:
            raise AnsibleError("Could not find vault address variable: {0}".format(vault_addr_key))

        vault_auth = {}
        app_role_key = 'ansible_app_role'
        for p in ('role_id', 'secret_id'):
            try:
                vault_auth[p] = kwargs.get(p, None)
                if not vault_auth[p]:
                    vault_auth[p] = myvars[app_role_key][p]
            except KeyError:
                raise AnsibleError("Vault \"{0}\" needs to be set in \"{1}\" variable or lookup params".format(
                        p, app_role_key))

        url = "{0}/v1/auth/approle/login".format(vault_addr)
        r = requests.post(url, data=json.dumps(vault_auth))
        if r.status_code != requests.codes.ok:
            raise AnsibleError("Vault lookup return response code: {0}".format(r.status_code))

        try:
            token = r.json()['auth']['client_token']
        except Exception as e:
            raise AnsibleError("Failed to retrieve client token: {0}".format(e))

        url = "{0}/v1/{1}".format(vault_addr, path)
        r = requests.get(url, headers={'X-Vault-Token': token})
        if r.status_code != requests.codes.ok:
            raise AnsibleError("Vault lookup of path \"{0}\" returned response code \"{1}\"".format(
                url, r.status_code))

        try:
            resp = r.json()['data']
        except Exception as e:
            raise AnsibleError("Failed to retrieve vault data: {0}: {1}".format(
                path, e))

        ret.append(resp)

        return ret

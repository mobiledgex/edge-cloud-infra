from __future__ import (absolute_import, division, print_function)
__metaclass__ = type

DOCUMENTATION = """
lookup: vault
short_description: look up key values in Hashicorp vault
description:
  - Look up kv secret values from a Hashicorp vault
options:
  _terms:
    description:
        - Paths of the secrets to look up in the vault
        - Values returned in a dict with the key being the basename of path,
          with dashes replaced by underscores
        - Key can be overridden by specifying lookup in the form of "path:key"
    required: True
  vault_addr:
    description: The vault to connect to
    default: The "vault_address" ansible variable
    required: False
  role_id:
    description: Vault role ID to generate the login token for
    default: The "ansible_app_role.role_id" ansible variable
    required: False
  secret_id:
    description: Vault secret ID of the role to generate the login token for
    default: The "ansible_app_role.secret_id" ansible variable
    required: False
"""

EXAMPLES="""
  - name: Look up influx DB creds and GCP service account
    set_fact:
      vault_lookup: "{{ lookup('vault', influxdb_path, gcp_path) }}"
    vars:
      influxdb_path: "secret/EU/accounts/influxdb"
      gcp_path: "secret/ansible/main/gcp-registry-reader-service-account:gcp"

  - debug: var=vault_lookup.influxdb.data
  - debug: var=vault_lookup.gcp.data
"""

from ansible.errors import AnsibleError, AnsibleParserError
from ansible.plugins.lookup import LookupBase
from ansible.utils.display import Display

import json
import os
import re

try:
    import requests
    HAS_REQUESTS = True
except ImportError as e:
    HAS_REQUESTS = False

display = Display()

class LookupModule(LookupBase):

    def _lookup_path(self, vault_addr, path, token):
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

        return resp

    def run(self, terms, variables=None, **kwargs):
        if not HAS_REQUESTS:
            raise AnsibleError(
                'python requests package is required for vault lookups')

        if variables is not None:
            self._templar.available_variables = variables
        myvars = getattr(self._templar, '_available_variables', {})

        ret = []

        if len(terms) < 1:
            raise AnsibleError('vault lookup needs at least one path argument')

        try:
          paths = [ re.sub(r'^secret/(?:data/)?', 'secret/data/', x) for x in terms ]
        except Exception as e:
          raise AnsibleError("Error loading lookup paths; undefined variable in list, perhaps?")

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

        resp = {}
        for item in paths:
            tokens = item.split(':', 1)
            path = tokens.pop(0)
            if tokens:
                key = tokens.pop()
            else:
                key = os.path.basename(path).replace('-', '_')

            resp[key] = self._lookup_path(vault_addr, path, token)

        ret.append(resp)

        return ret

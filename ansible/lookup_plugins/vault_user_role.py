from __future__ import (absolute_import, division, print_function)
__metaclass__ = type

DOCUMENTATION = """
lookup: vault_user_role
short_description: policies for vault user role
description:
  - get list of policies for given vault user role
options:
  _terms:
    description:
        - Role to look up
    required: True
"""

EXAMPLES="""
  - set_fact:
      policies: "{{ query('vault_user_role', 'viewer') }}"
  - debug: var=policies
"""

from ansible.errors import AnsibleError, AnsibleParserError
from ansible.plugins.lookup import LookupBase
from ansible.utils.display import Display

import re

display = Display()

class LookupModule(LookupBase):

    def flatten_list(self, l):
        nl = set()
        for item in l:
            if isinstance(item, list):
                nl |= self.flatten_list(item)
            else:
                nl.add(item)

        return nl

    def run(self, terms, variables=None, **kwargs):
        if variables is not None:
            self._templar.available_variables = variables
        myvars = getattr(self._templar, '_available_variables', {})

        ret = []

        if len(terms) < 1:
            raise AnsibleError('Role to look up not provided')

        roles = re.split(r'\s*,\s*', terms[0])
        policies = []

        for role in roles:
            try:
                policies.append(myvars['vault_user_roles'][role])
            except KeyError:
                display.v("Github user role not found: {0}".format(role))

        return sorted(self.flatten_list(policies))

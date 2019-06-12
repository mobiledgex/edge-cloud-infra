from __future__ import (absolute_import, division, print_function)
__metaclass__ = type

from ansible.errors import AnsibleFilterError
from ansible.utils import helpers

from distutils.version import LooseVersion
import re

ANSIBLE_METADATA = {
    'metadata_version': '1.0',
}

def semver_sort(versions, re_filter=None, reverse=False):
    full_ver_map = {}
    ver_list = []
    if re_filter:
        ngroups = re.compile(re_filter).groups
        if ngroups > 1:
            raise AnsibleFilterError("Too many capture groups in re_filter: {0}".format(
                re_filter))

        for v in versions:
            m = re.search(re_filter, v)
            if not m:
                continue
            if ngroups > 0:
                ver_list.append(m.group(1))
                full_ver_map[m.group(1)] = v
            else:
                ver_list.append(v)
    else:
        ver_list = versions

    try:
        sorted_ver_list = sorted(ver_list, key=LooseVersion, reverse=reverse)
    except ValueError as e:
        raise AnsibleFilterError("{0}: need an re_filter, perhaps?".format(
            str(e)))

    return [ full_ver_map[x] if x in full_ver_map else x for x in sorted_ver_list ]

# ---- Ansible filters ----
class FilterModule(object):
    ''' Semver sort '''

    def filters(self):
        return {
            'semver_sort': semver_sort
        }

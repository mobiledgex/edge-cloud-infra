#!/usr/bin/python

from ansible.module_utils.basic import *
from ansible.module_utils.vault import vault_request

def vault_api_call(vault, api, method="GET", data={}, success_code=200, check_mode=False):
    if check_mode and method != "GET":
        return (False,
                {"msg": "Skipping {0} API call in check mode".format(method)})

    has_changed = True

    meta = {}

    kwargs = {
        "method": method,
        "success_code": success_code,
    }
    if data:
        kwargs["json"] = data

    r = vault(api, **kwargs)
    meta["response"] = r

    return (has_changed, meta)

def main():
    fields = {
        "api": {"required": True, "type": "str"},
        "method": {"type": "str", "default": "GET"},
        "data": {"type": "dict"},
        "success_code": {"type": "int", "default": 200},
        "vault_addr": {"required": True, "type": "str"},
        "vault_token": {"required": True, "type": "str"},
    }

    module = AnsibleModule(argument_spec=fields,
                           supports_check_mode=True)
    vault = vault_request(module.params["vault_addr"], module.params["vault_token"])
    has_changed, result = vault_api_call(vault,
                                         api=module.params["api"],
                                         method=module.params["method"],
                                         data=module.params["data"],
                                         success_code=module.params["success_code"],
                                         check_mode=module.check_mode)
    module.exit_json(changed=has_changed, meta=result)

if __name__ == '__main__':
    main()

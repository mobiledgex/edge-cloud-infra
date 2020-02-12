#!/usr/bin/python

from ansible.module_utils.basic import *
from ansible.module_utils.vault import vault_request

PARAM_DEFAULTS = {
    "period": 0,
    "policies": [],
    "token_max_ttl": 2592000, # 720h
    "token_ttl": 60,
    "token_type": "default",
}

def vault_role_present(vault, data, check_mode=False):
    has_changed = False
    create_role = False

    role = data.pop("name")
    meta = {}
    set_params = {}

    r = vault("auth/approle/role", method="LIST")
    if role in r["data"]["keys"]:
        r = vault("auth/approle/role/{0}".format(role))

        for param in PARAM_DEFAULTS:
            current = r["data"][param]
            required = data[param]
            if isinstance(current, list):
                current = sorted(current)
                required = sorted(required)
            if current != required:
                set_params[param] = required
                has_changed = True
    else:
        # Role does not exist
        has_changed = True
        create_role = True
        set_params = data

    if has_changed:
        if check_mode:
            meta["msg"] = "skipped, running in check mode"
        else:
            vault("auth/approle/role/{0}".format(role), method="POST",
                  success_code=204,
                  json=set_params)
            meta["msg"] = "role updated"
            meta["changed_params"] = set_params
    else:
        meta["msg"] = "no change"

    if not (check_mode and create_role):
        # Role exists or was created
        r = vault("auth/approle/role/{0}/role-id".format(role))
        meta["role_id"] = r["data"]["role_id"]

    return (has_changed, meta)

def vault_role_absent(vault, data=None, check_mode=False):
    has_changed = False

    role = data["name"]
    meta = {}

    r = vault("sys/policies/acl", method="LIST")
    if policy in r["data"]["keys"]:
        has_changed = True

    if has_changed:
        if check_mode:
            meta["msg"] = "skipped, running in check mode"
        else:
            r = vault("sys/policies/acl/{0}".format(policy), method="DELETE",
                      success_code=204)
            meta["msg"] = "policy deleted"

    return (has_changed, meta)

def main():
    fields = {
        "name": {"required": True, "type": "str"},
        "period": {"type": "int", "default": PARAM_DEFAULTS["period"]},
        "policies": {"type": "list", "default": PARAM_DEFAULTS["policies"]},
        "token_max_ttl": {"type": "int", "default": PARAM_DEFAULTS["token_max_ttl"]},
        "token_ttl": {"type": "int", "default": PARAM_DEFAULTS["token_ttl"]},
        "token_type": {"type": "str", "default": PARAM_DEFAULTS["token_type"]},
        "vault_addr": {"required": True, "type": "str"},
        "vault_token": {"required": True, "type": "str"},
        "state": {
            "default": "present",
            "choices": ["present", "absent"],
            "type": "str",
        },
    }
    choice_map = {
        "present": vault_role_present,
        "absent": vault_role_absent,
    }

    module = AnsibleModule(argument_spec=fields, supports_check_mode=True)
    vault = vault_request(
        module.params.pop("vault_addr"),
        module.params.pop("vault_token")
    )
    has_changed, result = choice_map.get(module.params.pop("state"))(
        vault, data=module.params, check_mode=module.check_mode)
    module.exit_json(changed=has_changed, meta=result)

if __name__ == '__main__':
    main()

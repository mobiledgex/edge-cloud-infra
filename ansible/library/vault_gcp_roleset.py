#!/usr/bin/python

from ansible.module_utils.basic import AnsibleModule
from ansible.module_utils.vault import vault_request

from string import Template

vault_addr = None

resource_tmpl = Template("""
resource "$resource" {
  roles = [$roles]
}
""")

class VaultGcpRolesetError(Exception):
    pass

def run_module():
    module_args = dict(
        name=dict(type="str", required=True),
        secret_type=dict(type="str", required=False, default="service_account_key"),
        bindings=dict(type="list", required=True),
        vault_addr=dict(type="str", required=True),
        vault_token=dict(type="str", required=True),
        state=dict(type="str", choices=["present", "absent"], default="present"),
    )

    result = dict(
        changed=False,
        data={},
    )

    module = AnsibleModule(
        argument_spec=module_args,
        supports_check_mode=True
    )

    vault = vault_request(module.params["vault_addr"], module.params["vault_token"])
    r = vault("gcp/roleset/{}".format(module.params["name"]))
    result["data"] = r

    current_secret_type = r["data"]["secret_type"]
    current_bindings = r["data"]["bindings"]

    if module.params["secret_type"] != current_secret_type:
        raise VaultGcpRolesetError("Can't change secret type for roleset (current: {})".format(
            current_secret_type))

    bindings_diff = {
        "added": [],
        "removed": [],
        "changed": [],
    }
    params_diff = []

    need_bindings = ""
    for b in module.params["bindings"]:
        resource = b["resource"]
        roles = b["roles"]
        roles_list = '"' + '", "'.join(roles) + '"'
        need_bindings += resource_tmpl.substitute(resource=resource, roles=roles_list)

        current_roles = current_bindings.pop(resource, None)
        if not current_roles:
            bindings_diff["removed"].append(resource)
        elif sorted(current_roles) != sorted(roles):
            bindings_diff["changed"].append(resource)

    if current_bindings:
        bindings_diff["added"] = current_bindings.keys()

    result["bindings_diff"] = bindings_diff
    for bucket in bindings_diff.keys():
        if len(bindings_diff[bucket]) > 0:
            result["changed"] = True
            break

    if module.check_mode:
        return module.exit_json(**result)

    if result["changed"]:
        # Update roleset
        r = vault("gcp/roleset/{}".format(module.params["name"]),
              method="POST",
              json={
                  "secret_type": module.params["secret_type"],
                  "bindings": need_bindings,
              },
              success_code=204)
        result["response"] = r

    module.exit_json(**result)

def main():
    run_module()

if __name__ == '__main__':
    main()

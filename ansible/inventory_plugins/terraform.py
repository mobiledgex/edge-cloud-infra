from __future__ import (absolute_import, division, print_function)
__metaclass__ = type

DOCUMENTATION = '''
    author: Venky (venky.tumkur@mobiledgex.com)
    name: terraform
    short_description: terraform inventory source
    description:
        - Get inventory hosts from the local terraform state file
        - Uses a YAML configuration file that ends with .tf.yml
    extends_documentation_fragment:
      - constructed
'''

EXAMPLES = r'''
    # ansible-playbook -i staging.tf.yml mexplay.yml
'''

from collections import defaultdict
import json
import os
import re
import subprocess

from ansible.errors import AnsibleError, AnsibleParserError
from ansible.plugins.inventory import BaseInventoryPlugin, Constructable

class GoogleComputeInstance:
    resource_type = "google_compute_instance"
    registry = []

    def __init__(self, res):
        self.id = f"{res['module']}.{res['type']}.{res['name']}"

        inst = res["instances"][0]

        groups_label = inst["attributes"]["labels"].get("groups")
        if not groups_label:
            self.groups = set()
        else:
            self.groups = set(groups_label.split(","))

        self.is_vault = inst["attributes"]["labels"].get("vault") == "true"
        self.ip = inst["attributes"]["network_interface"][0]["access_config"][0]["nat_ip"]

        GoogleComputeInstance.registry.append(self)

    @classmethod
    def instances(cls):
        return cls.registry

    def hostname(self, vault=False):
        for cfrec in CloudflareRecord.registry:
            if cfrec.gcinst == self.id \
                    and (not vault or cfrec.name.startswith("vault")):
                return cfrec.name
        return None

    def __str__(self):
        return self.id

class CloudflareRecord:
    resource_type = "cloudflare_record"
    registry = []

    @classmethod
    def record_for_component(cls, component):
        for rec in cls.registry:
            if rec.component == component:
                return rec
        return None

    def __init__(self, res):
        self.id = f"{res['module']}.{res['type']}.{res['name']}"

        inst = res["instances"][0]

        self.name = inst["attributes"]["name"]
        self.ip = inst["attributes"]["value"]
        self.gcinst = inst["dependencies"][0]

        m = re.match(r'module\.(.*)_dns$', res["module"])
        if m:
            self.component = m.groups(1)
        else:
            self.component = None

        CloudflareRecord.registry.append(self)


    def __str__(self):
        return self.id

class InventoryModule(BaseInventoryPlugin, Constructable):

    NAME = 'terraform'

    def verify_file(self, path):
        valid = False
        if super(InventoryModule, self).verify_file(path):
            if path.endswith('.tf.yml'):
                valid = True
        return valid

    def add_host_vars(self, host, host_vars):
        for name, val in host_vars.items():
            self.inventory.set_variable(host, name, val)
        self._set_composite_vars(self.get_option('compose'), host_vars, host, strict=True)

    def parse(self, inventory, loader, path, cache=True):
        super(InventoryModule, self).parse(inventory, loader, path, cache)

        config = self._read_config_data(path)
        tfdir = os.path.join(os.path.dirname(os.path.realpath(__file__)), '../../terraform', config['dir'])
        tfstate = os.path.join(tfdir, "terraform.tfstate")

        groups = defaultdict(list)

        self.inventory.add_group("vault")

        with open(tfstate) as f:
            state = json.load(f)

        for res in state["resources"]:
            if res["type"] == CloudflareRecord.resource_type:
                cfrec = CloudflareRecord(res)

            elif res["type"] == GoogleComputeInstance.resource_type:
                gcrec = GoogleComputeInstance(res)

        deploy_environ = re.sub(r'\.tf\.ya?ml$', '', os.path.basename(path))
        host_vars = { "deploy_environ": deploy_environ }

        host_vars_file = os.path.join(os.path.dirname(path), "_tmpl", "vars.json")
        with open(host_vars_file) as f:
            addl_host_vars = json.load(f)
        host_vars.update(addl_host_vars)

        for gcrec in GoogleComputeInstance.instances():
            if gcrec.is_vault:
                host = gcrec.hostname(vault=True)
                self.inventory.add_host(host, group="vault")
                self.add_host_vars(host, host_vars)

            for component in gcrec.groups:
                cfrec = CloudflareRecord.record_for_component(component)
                if cfrec:
                    host = cfrec.name
                    self.inventory.add_group(component)
                    self.inventory.add_host(host, group=component)
                    self.add_host_vars(host, host_vars)

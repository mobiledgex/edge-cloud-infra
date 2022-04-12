#!/usr/bin/python3
# Copyright 2022 MobiledgeX, Inc
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


import argparse
from collections import OrderedDict
import json
import os
import re
import sys

API_ORDER = [
    '/v1/registerclient',
    '/v1/findcloudlet',
    '/v1/verifylocation',
]
TAG_SWITCH = {
    'App': 'Application',
    'AppInst': 'Application Instance',
    'ClusterInst': 'Cluster Instance',
    'MatchEngineApi': 'Edge REST API',
}

def splice_sample(operation, sample_dir):
    operation['x-code-samples'] = []
    for sample in os.listdir(sample_dir):
        tokens = sample.split(".")
        if tokens.pop(0) != 'sample':
            continue
        lang = tokens.pop(0)
        label = lang
        if tokens:
            lang = tokens.pop(0)
        with open(os.path.join(sample_dir, sample)) as f:
            code = f.read()
        operation['x-code-samples'].append({
            "lang": lang,
            "label": label,
            "source": code,
        })

def add_logo(sw, logo):
    sw['info']['x-logo'] = {
        "version": "1.0",
        "url": logo,
        "backgroundColor": "#fafafa",
    }

def set_version(sw, vers):
    if sw['info']['version'] == "version not set":
        sw['info']['version'] = vers

def set_environ(sw, environ):
    if 'description' in sw['info']:
        sw['info']['description'] = re.sub(r'{{ENVIRON}}', environ,
                                           sw['info']['description'])

def splice_samples(sw, samples):
    if samples:
        api_eps = set([d for d in os.listdir(samples) if os.path.isdir(os.path.join(samples, d))])
    else:
        api_eps = set()

    for path in sw['paths']:
        for op in sw['paths'][path]:
            opid = sw['paths'][path][op]['operationId']
            if opid in api_eps:
                splice_sample(sw['paths'][path][op],
                              os.path.join(samples, opid))

def order_apis(sw):
    paths = OrderedDict()
    for api in API_ORDER:
        spec = sw['paths'].pop(api, None)
        if spec:
            paths[api] = spec

    for x in sorted(sw['paths']):
        paths[x] = sw['paths'][x]

    sw['paths'] = paths

def switch_tags(sw):
    for path in sw['paths']:
        for op in sw['paths'][path]:
            if not 'tags' in sw['paths'][path][op]:
                continue
            for i, tag in enumerate(sw['paths'][path][op]['tags']):
                if tag in TAG_SWITCH:
                    sw['paths'][path][op]['tags'][i] = TAG_SWITCH[tag]

    for group_idx, tag_group, in enumerate(sw.get('x-tagGroups', [])):
        if 'tags' in tag_group:
            for i, tag in enumerate(tag_group['tags']):
                if tag in TAG_SWITCH:
                    sw['x-tagGroups'][group_idx]['tags'][i] = TAG_SWITCH[tag]

def remove_hidden_params_from_object(obj):
    for prop in list(obj.get('properties', {})):
        for param in ('title', 'description'):
            value = obj['properties'][prop].get(param)
            if value and "_(hidden)_" in value:
                del obj['properties'][prop]
                break

def remove_hidden_params(sw):
    for defn in sw['definitions']:
        if sw['definitions'][defn]['type'] != "object":
            continue
        remove_hidden_params_from_object(sw['definitions'][defn])

def set_required_params_for_object(obj):
    if 'required' in obj:
        # Object already has a required property list
        return

    reqd = []
    found_optional = False
    for prop in obj.get('properties', {}):
        nsubs = 0
        for param in ('title', 'description'):
            value = obj['properties'][prop].get(param)
            if value:
                (nvalue, nsub) = re.subn(r'\s*_\(optional\)_\s*', '', value)
                obj['properties'][prop][param] = nvalue
                nsubs += nsub

        if nsubs > 0:
            found_optional = True
        else:
            reqd.append(prop)

    # Only set required props list if there is at least one "_(optional)_" tag
    if found_optional:
        obj['required'] = reqd

def set_required_params(sw):
    for defn in sw['definitions']:
        if sw['definitions'][defn]['type'] != "object":
            continue
        set_required_params_for_object(sw['definitions'][defn])

# Remove periods from summaries generated by go-swagger
def fix_go_swagger_summaries(sw):
    for path in sw['paths']:
        for op in sw['paths'][path]:
            summary = sw['paths'][path][op].get('summary', '')
            if summary and summary.endswith('.'):
                sw['paths'][path][op]['summary'] = summary.rstrip('.')

def set_host(sw, host):
    sw["host"] = host

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--samples", "-s", help="Code samples directory")
    parser.add_argument("--logo", "-l", help="Logo URL", default="/swagger/logo.svg")
    parser.add_argument("--environ", "-e", help="Deployment environment", default="mexdemo")
    parser.add_argument("--version", "-v", help="Version string, if not present in swagger",
                        default="1.0")
    parser.add_argument("--host", help="API host")
    args = parser.parse_args()

    sw = json.load(sys.stdin)

    fix_go_swagger_summaries(sw)
    if args.samples:
        splice_samples(sw, args.samples)
    switch_tags(sw)
    order_apis(sw)
    remove_hidden_params(sw)
    set_required_params(sw)
    add_logo(sw, args.logo)
    set_version(sw, args.version)
    if args.host:
        set_host(sw, args.host)
    if args.environ:
        set_environ(sw, args.environ)

    print(json.dumps(sw, indent=2))

if __name__ == "__main__":
    main()

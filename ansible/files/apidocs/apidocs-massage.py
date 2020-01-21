#!/usr/bin/python3

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

def add_logo(sw):
    sw['info']['x-logo'] = {
        "version": "1.0",
        "url": "/swagger/logo.svg",
        "backgroundColor": "#fafafa",
    }

def set_version(sw, vers):
    if sw['info']['version'] == "version not set":
        sw['info']['version'] = vers

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

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--samples", "-s", help="Code samples directory")
    parser.add_argument("--version", "-v", help="Version string, if not present in swagger",
                        default="1.0")
    args = parser.parse_args()

    sw = json.load(sys.stdin)

    if args.samples:
        splice_samples(sw, args.samples)
    order_apis(sw)
    add_logo(sw)
    set_version(sw, args.version)

    print(json.dumps(sw, indent=2))

if __name__ == "__main__":
    main()

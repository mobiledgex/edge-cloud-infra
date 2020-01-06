#!/usr/bin/python3

import argparse
from collections import OrderedDict
import json
import os
import re

CODE_SAMPLES_DIR_DEF = '/var/swagger/code-samples'
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

def splice_samples(swagger, samples):
    api_eps = set([d for d in os.listdir(samples) if os.path.isdir(os.path.join(samples, d))])
    with open(swagger) as f:
        sw = json.load(f)

    for path in sw['paths']:
        for op in sw['paths'][path]:
            opid = sw['paths'][path][op]['operationId']
            if opid in api_eps:
                splice_sample(sw['paths'][path][op],
                              os.path.join(samples, opid))

    paths = OrderedDict()
    for api in API_ORDER:
        spec = sw['paths'].pop(api, None)
        if spec:
            paths[api] = spec

    for x in sorted(sw['paths']):
        paths[x] = sw['paths'][x]

    sw['paths'] = paths

    add_logo(sw)

    return json.dumps(sw, indent=2)

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("swagger", help="Swagger file to splice the code samples into")
    parser.add_argument("output", help="File to write output swagger to")
    parser.add_argument("--samples", "-s", help="Code samples directory",
                        default=CODE_SAMPLES_DIR_DEF)
    args = parser.parse_args()

    output = splice_samples(args.swagger, args.samples)

    with open(args.output, "w") as f:
        f.write(output)

if __name__ == "__main__":
    main()

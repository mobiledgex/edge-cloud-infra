#!/usr/bin/python3

import argparse
import json
import re
import sys

def remove_mobiledgex_references(sw):
    for item in list(sw):
        if isinstance(sw[item], dict):
            remove_mobiledgex_references(sw[item])
        elif item == "x-go-package":
            del sw[item]
        elif item == "description":
            sw[item] = re.sub(r'MobiledgeX\s*', '', sw[item]).capitalize()

def fix_introduction(sw, title=None, description=None):
    try:
        if title:
            sw['info']['title'] = title
        else:
            sw['info']['title'] = re.sub(r'MobiledgeX\s*', '', sw['info']['title'])

        if not description:
            del sw['info']['description']
        else:
            sw['info']['description'] = description
    except Exception:
        pass

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--title", "-t", help="Replacement title text")
    parser.add_argument("--description", "-d", help="Replacement description text")
    args = parser.parse_args()

    sw = json.load(sys.stdin)

    remove_mobiledgex_references(sw)
    fix_introduction(sw, title=args.title, description=args.description)

    print(json.dumps(sw, indent=2))

if __name__ == "__main__":
    main()

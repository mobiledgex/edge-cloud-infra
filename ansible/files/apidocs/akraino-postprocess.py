#!/usr/bin/python3

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

def fix_introduction(sw):
    try:
        sw['info']['title'] = re.sub(r'MobiledgeX\s*', '', sw['info']['title'])
        del sw['info']['description']
    except Exception:
        pass

def main():
    sw = json.load(sys.stdin)

    remove_mobiledgex_references(sw)
    fix_introduction(sw)

    print(json.dumps(sw, indent=2))

if __name__ == "__main__":
    main()

from argparse import Namespace
from base64 import b64decode
import datetime
import sys

from gvm.protocols.gmp import Gmp

report_format_id = "a3810a62-1f62-11e1-9219-406186ea4fc5"

def main(gmp: Gmp, args: Namespace) -> None:
    if len(args.script) < 2:
        print("usage: gvm-script scan-report.py <scan-id>")
        sys.exit()

    scan_id = args.argv[1]

    response_xml = gmp.get_tasks(filter_string=scan_id)
    tasks_xml = response_xml.xpath('task')

    ntasks = len(tasks_xml)
    if ntasks < 1:
        sys.exit(f"Could not locate task: {scan_id}")

    if ntasks > 1:
        sys.exit(f"Scan pattern matched {ntasks} tasks")

    task = tasks_xml[0]
    try:
        report_id = task.xpath('last_report/report')[0].get('id')
    except:
        sys.exit(f"Report not ready for task: {scan_id}")

    response = gmp.get_report(
            report_id=report_id,
            report_format_id=report_format_id)
    report_element = response.find("report")
    content = report_element.find("report_format").tail

    print(b64decode(content.encode('ascii')).decode('ascii'))

# Only run from within "gvm-script"
if __name__ == '__gmp__':
    main(gmp, args)

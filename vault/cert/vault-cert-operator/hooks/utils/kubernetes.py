import base64
import logging
import os
import subprocess
import tempfile
from yaml import load, dump, Loader, Dumper

from utils.vault import vault_login, vault_get_cert

def kubectl(command, namespace="", output="yaml", check=True,
            capture_output=True, text=True, **kwargs):
    if not isinstance(command, list):
        command = command.split()
    if command[0] != "kubectl":
        command.insert(0, "kubectl")

    # Add namespace into command line
    if namespace:
        command.append(f"--namespace={namespace}")
    else:
        logging.debug("Running command in all namespaces")
        command.append("--all-namespaces")

    if output:
        command.append(f"--output={output}")

    logging.debug(f"Command: {command}")
    return subprocess.run(command, check=check, capture_output=capture_output,
                          text=text, **kwargs)

def create_tls_secret(binding):
    meta = binding["object"]["metadata"]
    name = meta["name"]
    namespace = meta["namespace"]

    spec = binding["object"]["spec"]
    domains = spec["domain"]
    secretname = spec["secretName"]

    p = kubectl(["get", "secret", secretname], namespace=namespace, check=False)
    if p.returncode == 0:
        print(f"Secret {secretname} exists in namespace {namespace}")
    else:
        # Create secret
        with tempfile.TemporaryDirectory() as tmpdir:
            crtfile = os.path.join(tmpdir, "cert")
            keyfile = os.path.join(tmpdir, "key")

            logging.info(f"Creating TLS secret {secretname} for {name} in namespace {namespace}")
            vcert = vault_get_cert(vault_login(), domains)

            with open(crtfile, "w") as f:
                f.write(vcert["cert"])

            with open(keyfile, "w") as f:
                f.write(vcert["key"])

            kubectl(["create", "secret", "tls", secretname,
                     f"--cert={crtfile}", f"--key={keyfile}"],
                    namespace=namespace)

def delete_tls_secret(binding):
    meta = binding["object"]["metadata"]
    name = meta["name"]
    namespace = meta["namespace"]

    spec = binding["object"]["spec"]
    domains = spec["domain"]
    secretname = spec["secretName"]

    logging.info(f"Deleting TLS secret {secretname} for {name} from namespace {namespace}")
    p = kubectl(["delete", "secret", secretname],
                namespace=namespace, check=False, output=None)
    if p.returncode != 0:
        if "NotFound" in p.stderr:
            logging.info(f"Secret {secretname} does not exist in namespace {namespace}")
        else:
            raise Exception(f"Error deleting secret {secretname} ({namespace}): {p.stderr}")

def get_tls_cert_from_secret(secretname, namespace):
    p = kubectl(["get", "secret", secretname], namespace=namespace)
    secret = load(p.stdout, Loader=Loader)
    return base64.b64decode(secret["data"]["tls.crt"]).decode('ascii')

def patch_tls_cert_in_secret(secretname, namespace, cert):
    logging.info(f"Patching {secretname} in namespace {namespace}")
    b64cert = base64.b64encode(cert["cert"].encode("ascii")).decode("ascii")
    b64key = base64.b64encode(cert["key"].encode("ascii")).decode("ascii")
    patch = dump({
        "data": {
            "tls.crt": b64cert,
            "tls.key": b64key,
        },
    }, Dumper=Dumper)
    kubectl(["patch", "secret", secretname, "--patch", patch],
            namespace=namespace)

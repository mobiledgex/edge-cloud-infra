import subprocess

def get_cert_fingerprint(cert):
    p = subprocess.Popen(["openssl", "x509", "-noout", "-fingerprint"],
                         stdout=subprocess.PIPE,
                         stderr=subprocess.PIPE,
                         stdin=subprocess.PIPE,
                         text=True)
    out, err = p.communicate(input=cert)
    if p.returncode != 0:
        raise Exception(f"Failed to load cert: {p.returncode} {err}")

    if not out.startswith("SHA1 Fingerprint="):
        raise Exception(f"Error retrieving fingerprint: {out}: {err}")

    return out.strip()

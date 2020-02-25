import requests 

class VaultException(Exception):
    pass

def vault_request(addr, token):
    def vault_api_call(path, method="GET", success_code=requests.codes.ok,
                       **kwargs):
        url = "{0}/v1/{1}".format(addr, path)
        data = {
            "headers": { "X-Vault-Token": token }
        }
        data.update(kwargs)
        r = requests.request(method, url, **data)
        if r.status_code != success_code:
            raise VaultException("Got: {0} {1} (Expected: {2})".format(
                r.status_code, r.text, success_code))

        if r.status_code == requests.codes.no_content:
            return ''
        return r.json()
    return vault_api_call

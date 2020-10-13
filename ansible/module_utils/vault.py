import requests 

class VaultException(Exception):
    pass

def vault_request(addr, token):
    def vault_api_call(path, method="GET", success_code=[requests.codes.ok],
                       raw_response=False, **kwargs):
        url = "{0}/v1/{1}".format(addr, path)
        if method == "LIST":
            method = "GET"
            url += "?list=true"

        data = {
            "headers": { "X-Vault-Token": token }
        }
        data.update(kwargs)
        r = requests.request(method, url, **data)
        if not isinstance(success_code, list):
            success_code = [ success_code ]
        if r.status_code not in success_code:
            raise VaultException("Got: {3} {4} {5}: {0} {1} (Expected one of: {2})".format(
                r.status_code, r.text, success_code, method, url, data))

        if r.status_code == requests.codes.no_content:
            return ''
        if raw_response:
            return r.text
        return r.json()
    return vault_api_call

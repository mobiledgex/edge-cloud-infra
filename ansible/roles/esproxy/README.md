## Creating an API key for esproxy

```
http --auth elastic POST https://3dd88757c3df44ac8960e53fc6a9a2d5.us-central1.gcp.cloud.es.io:9243/_security/api_key <<EOT
{
  "name": "esproxy",
  "role_descriptors": {
    "esproxy": {
      "cluster": [ "monitor" ],
      "index": [
        {
          "names": [ "events-*" ],
          "privileges": [
            "create",
            "create_index",
            "write",
            "read",
            "monitor"
          ]
        }
      ]
    }
  }
}
EOT
```

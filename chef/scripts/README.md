### Knife Exec Scripts

Reference : https://docs.chef.io/workstation/knife_exec/

##### Supported scripts:
* `set_diagnostics.rb`: Used to set diagnostics on upcoming chef-client run of a node, will make chef-client collect logs and send it to artifactory. Hence, artifactory token is required to execute this script
```
knife exec set_diagnostics.rb <node-name> <tar-file-name> <artifactory-token>
```

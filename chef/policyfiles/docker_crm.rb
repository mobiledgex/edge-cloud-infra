# Uses docker containers to setup cloudlet services

# A name that describes what the system you're building with Chef does.
name 'docker_crm'

# Where to find external cookbooks:
default_source :chef_server, "https://chef.mobiledgex.net/organizations/mobiledgex"

# run_list: chef-client will run these recipes in the order specified.
run_list 'chef_client_updater::default', 'recipe[runstatus_handler]', 'recipe[run_cmd]', 'recipe[setup_infra]', 'recipe[preflight_crm_checks]', 'recipe[run_diagnostics]', 'recipe[setup_services::docker]'

# Specify a custom source for a single cookbook:
cookbook 'runstatus_handler', '= 1.0.0'
cookbook 'setup_infra', '= 1.0.0'
cookbook 'preflight_crm_checks', '= 1.0.0'
cookbook 'setup_services', '= 1.0.31'
cookbook 'docker', '= 7.7.0'
cookbook 'run_diagnostics', '= 1.0.0'
cookbook 'run_cmd', '= 1.0.0'
cookbook 'chef_client_updater', '= 3.11.0'

# Set chef-client version
default['chef_client_updater']['version'] = '17.2.29'

# Set edgeCloudVersion (i.e. edge-cloud docker base image version) for all the cloudlets
override['main']['edgeCloudVersion'] = '2021-04-15-1'
# By default, commercialCerts is not on. Hence add override to turn it on for all the cloudlets
override['main']['crmserver']['args']['commercialCerts'] = ""

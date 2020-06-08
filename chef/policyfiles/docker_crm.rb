# Uses docker containers to setup cloudlet services

# A name that describes what the system you're building with Chef does.
name 'docker_crm'

# This lets you source cookbooks from your chef-repo.
default_source :chef_repo, '../'

# Where to find external cookbooks:
default_source :supermarket

# run_list: chef-client will run these recipes in the order specified.
run_list 'recipe[runstatus_handler]', 'recipe[setup_infra]', 'recipe[preflight_crm_checks]', 'recipe[setup_services::docker]'

# Specify a custom source for a single cookbook:
cookbook 'runstatus_handler', '= 1.0.0'
cookbook 'setup_infra', '= 1.0.0'
cookbook 'preflight_crm_checks', '= 1.0.0'
cookbook 'setup_services', '= 1.0.0'
cookbook 'docker', '= 6.0.0'

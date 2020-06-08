# Base role for common recipes to be executed

# A name that describes what the system you're building with Chef does.
name 'base'

# This lets you source cookbooks from your chef-repo.
default_source :chef_repo, '../'

# Where to find external cookbooks:
default_source :supermarket

# run_list: chef-client will run these recipes in the order specified.
run_list 'recipe[runstatus_handler@1.0.0]'

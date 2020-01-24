MobiledgeX Development Setup Procedure

setup_dev_env.sh sets up your environment locally so that you can run edgebox.
It essentially automates the steps on 'Getting Started' Confluence page (https://mobiledgex.atlassian.net/wiki/spaces/SWDEV/pages/22478869/Getting+Started)


Make sure your git setup is working.
   - you have a git account (aka GITHUB_ID) that is linked to mobilegex github account.
       * You have to provide your GITHUB_ID to Sunay to link it to MobiledgeX github account. 
   - you have git personal token setup. Getting Started page has instructions.
   - your  GITHUB_ID has access to vault secrets. Venky sets this one up. 
   - If you do not have working GITHUB_ID, you will need access through VAULT ROLE. 

Backups:
  - If you have pre-existing  repos with existing changes and choose to force_remove_preexisting_repos, backup is performed before deletig the repos.
  - backup is compressed tar file located at ~/edge-cloud-repos-backup.<timestamp>.tgz path in your home directory.
  - Please delete the backups you don't need.
  

Get the setup_dev_env.sh script from github.com mobiledgex repo edge-cloud-infra under getting-started directory. Save it in /tmp directory on your macbook under your home directory and run it from there,

  ./setup_dev_env.sh
  


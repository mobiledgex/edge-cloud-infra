install_edgebox.sh sets up your environment locally so that you can run edgebox. 
It essentially automates the steps on 'Getting Started' Confluence page (https://mobiledgex.atlassian.net/wiki/spaces/SWDEV/pages/22478869/Getting+Started)


Make sure your git setup is working.
   - you have a git account (aka GITHUB_ID) that is linked to mobilegex github account.
       * You have to provide your GITHUB_ID to Sunay to link it to MobiledgeX github account. 
   - you have git personal token setup. Getting Started page has instructions.
   - your  GITHUB_ID has access to vault secrets. Venky sets this one up. 
   - If you do not have working GITHUB_ID, you will need access through VAULT ROLE. 

Backups:
  - Every time you run install_edgebox.sh it will  make backup of pre-existing edge-cloud-infra, edge-cloud and edge-proto repo 
  - Backups are located at ~/edge-cloud-repos-backup.<timestamp>.tgz in your home directory.
  - Please delete the backups you don't need.
  

Get the install_edgebox.sh script from github.com mobiledgex repo edge-cloud-infra under e2e-tests/edgebox directory.
Save it in a directory (say edgebox) on your macbook under your home directory and run it from there,

  ./install_edgebox.sh
  
Run edgebox_start.py: If you need more details check Step 6 at Confluence Page https://mobiledgex.atlassian.net/wiki/spaces/SWDEV/pages/410484743/Edge-in-a-box+setup

   cd ~/go/src/github.com/mobiledgex/edge-cloud-infra/e2e-tests/edgebox
   ./edgebox_startup.py


Create Cluster Instance from web console 
   Using your browser go to console-stage.mobiledgex.net
   Verify that the cloudlet you specified in previous step is availabel in the region you selected
   Create a cluster instance in the cloudlet you specified from 'Cluster Instances' page
   Launch an App Instance for 'MobiledgeX SDK Demo' app in the Cluster Instance you created
   Launch an App Instance for 'Face Detection Demo' app in the Cluster Instance you created

Using Your MobiledgeX SDK Demo app demonstrate Face Detection Demo App

Cleanup: 
   cd ~/go/src/github.com/mobiledgex/edge-cloud-infra/e2e-tests/edgebox
  ./edgebox_cleanup.py 

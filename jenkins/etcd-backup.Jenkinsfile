pipeline {
    options {
        timeout(time: 10, unit: 'MINUTES')
    }
    agent any
    environment {
        ANSIBLE_VAULT_PASSWORD_FILE = credentials('ansible-mex-vault-pass-file')
        ARM_ACCESS_KEY = credentials('azure-storage-access-key')
    }
    stages {
        stage('Backup') {
            steps {
                dir(path: 'ansible') {
                    ansiColor('xterm') {
                        sh label: 'Run ansible playbook', script: '''$!/bin/bash
export ANSIBLE_FORCE_COLOR=true
for DEPLOY_ENVIRON in mexdemo staging; do
    ./deploy.sh -p etcd-backup.yml "$DEPLOY_ENVIRON"
done
                        '''
                    }
                }
            }
        }
    }
}

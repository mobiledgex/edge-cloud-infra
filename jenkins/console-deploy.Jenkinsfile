pipeline {
    options {
        timeout(time: 30, unit: 'MINUTES')
    }
    agent any
    environment {
        ANSIBLE_VAULT_PASSWORD_FILE = credentials('ansible-mex-vault-pass-file')
        TAG = "${params.TAG}"
        DEPLOY_ENVIRONMENT = "${params.DEPLOY_ENVIRONMENT}"
    }
    parameters {
        string name: 'TAG', description: 'Console version (tag) to deploy'
        string name: 'DEPLOY_ENVIRONMENT', defaultValue: 'staging', description: 'Environment to deploy to'
    }
    stages {
        stage('Set up build tag') {
            steps {
                script {
                    assert params.TAG != null : "TAG not provided for deployment"
                }
                script {
                    currentBuild.displayName = "${params.TAG}"
                }
            }
        }
        stage('Deploy') {
            steps {
                dir(path: 'ansible') {
                    ansiColor('xterm') {
                        sh label: 'Run ansible playbook', script: '''$!/bin/bash
export ANSIBLE_FORCE_COLOR=true
ansible-playbook -i "${DEPLOY_ENVIRONMENT}" -e "console_version=${TAG}" -e @ansible-mex-vault.yml -l console mexplat.yml
                        '''
                    }
                }
            }
        }
    }
}

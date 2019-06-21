pipeline {
    options {
        timeout(time: 30, unit: 'MINUTES')
    }
    agent any
    environment {
        ANSIBLE_VAULT_PASSWORD_FILE = credentials('ansible-mex-vault-pass-file')
    }
    parameters {
        string name: 'TAG'
        string name: 'DEPLOY_ENVIRONMENT', defaultValue: 'staging'
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
                    sh label: 'Run ansible playbook', script: '''$!/bin/bash
echo ansible-playbook -i "${params.DEPLOY_ENVIRONMENT}" -e "console_version=${params.TAG}" -e @ansible-mex-vault.yml -l console mexplat.yml
                    '''
                }
            }
        }
    }
}

pipeline {
    options {
        timeout(time: 30, unit: 'MINUTES')
    }
    agent any
    environment {
        DOCKER_BUILD_TAG = sh(returnStdout: true, script: 'date +"%Y-%m-%d" | tr -d "\n"')
        ANSIBLE_VAULT_PASSWORD_FILE = credentials('ansible-mex-vault-pass-file')
        ARM_ACCESS_KEY = credentials('azure-storage-access-key')
        GCP_AUTH_KIND = 'serviceaccount'
        GCP_SERVICE_ACCOUNT_FILE = credentials('jenkins-terraform-gcp-credentials')
        GOOGLE_CLOUD_KEYFILE_JSON = credentials('jenkins-terraform-gcp-credentials')
    }
    stages {
        stage('Set up build tag') {
            steps {
                script {
                    currentBuild.displayName = "${DOCKER_BUILD_TAG}"
                }
            }
        }
        stage('Deploy') {
            steps {
                dir(path: 'ansible') {
                    sh label: 'Run ansible playbook', script: '''$!/bin/bash
ansible-playbook -i staging -e "edge_cloud_version=${DOCKER_BUILD_TAG}" -e @ansible-mex-vault.yml mexplat.yml
                    '''
                }
            }
        }
    }
}

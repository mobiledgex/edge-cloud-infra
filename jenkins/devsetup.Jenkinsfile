pipeline {
    options {
        timeout(time: 45, unit: 'MINUTES')
    }
    agent any
    parameters {
        string(name: 'DOCKER_BUILD_TAG', defaultValue: '', description: 'Docker build tag for the custom build')
    }
    environment {
        DEFAULT_DOCKER_BUILD_TAG = sh(returnStdout: true, script: 'date +"%Y-%m-%d" | tr -d "\n"')
        ANSIBLE_VAULT_PASSWORD_FILE = credentials('ansible-mex-vault-pass-file')
        ARM_ACCESS_KEY = credentials('azure-storage-access-key')
        GCP_AUTH_KIND = 'serviceaccount'
        GCP_SERVICE_ACCOUNT_FILE = credentials('jenkins-terraform-gcp-credentials')
        GOOGLE_CLOUD_KEYFILE_JSON = credentials('jenkins-terraform-gcp-credentials')
        GITHUB_CREDS = credentials('ansible-github-credentials')
        AZURE_SERVICE_PRINCIPAL = credentials('azure-service-principal')
        AZURE_SUBSCRIPTION_ID = credentials('azure-subscription-id')
        AZURE_TENANT = credentials('azure-tenant-id')
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
                    ansiColor('xterm') {
                        sh label: 'Run ansible playbook', script: '''$!/bin/bash
export GITHUB_USER="${GITHUB_CREDS_USR}"
export GITHUB_TOKEN="${GITHUB_CREDS_PSW}"
export AZURE_CLIENT_ID="${AZURE_SERVICE_PRINCIPAL_USR}"
export AZURE_SECRET="${AZURE_SERVICE_PRINCIPAL_PSW}"
export ANSIBLE_FORCE_COLOR=true

[[ -n "$DOCKER_BUILD_TAG" ]] || DOCKER_BUILD_TAG="$DEFAULT_DOCKER_BUILD_TAG"
ansible-playbook -e @ansible-mex-vault-development.yml -e "edge_cloud_version=${DOCKER_BUILD_TAG}" -i development mexplat.yml
                        '''
                    }
                }
            }
        }
    }
}

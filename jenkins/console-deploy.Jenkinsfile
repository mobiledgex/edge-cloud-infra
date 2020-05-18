pipeline {
    options {
        timeout(time: 30, unit: 'MINUTES')
    }
    agent any
    environment {
        ANSIBLE_ROLE = credentials('staging-vault-ansible-role')
        GITHUB_CREDS = credentials('ansible-github-credentials')
        TAG = "${params.TAG}"
    }
    parameters {
        string name: 'TAG', defaultValue: 'latest', description: 'Console version (tag) to deploy'
    }
    stages {
        stage('Set up build tag') {
            steps {
                script {
                    currentBuild.displayName = "${params.TAG}"
                }
            }
        }
        stage('Deploy') {
            steps {
                dir(path: 'ansible') {
                    ansiColor('xterm') {
                        sh label: 'Run ansible playbook', script: '''#!/bin/bash
export ANSIBLE_FORCE_COLOR=true
export GITHUB_USER="${GITHUB_CREDS_USR}"
export GITHUB_TOKEN="${GITHUB_CREDS_PSW}"
export VAULT_ROLE_ID="${ANSIBLE_ROLE_USR}"
export VAULT_SECRET_ID="${ANSIBLE_ROLE_PSW}"

CMD=( ./deploy.sh -y -s setup,mc )
[[ "$TAG" != "latest" ]] && CMD+=( -C ${TAG} )
"${CMD[@]}" staging console
                        '''
                    }
                }
            }
        }
    }
}

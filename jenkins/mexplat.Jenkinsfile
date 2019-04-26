pipeline {
    options {
        timeout(time: 30, unit: 'MINUTES')
    }
    agent any
    environment {
        DOCKER_BUILD_TAG = sh(returnStdout: true, script: 'date +"%Y-%m-%d" | tr -d "\n"')
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
ansible-playbook -i development --extra-vars "edge_cloud_version=${DOCKER_BUILD_TAG}" mexplat.yml
                    '''
                }
            }
        }
    }
}

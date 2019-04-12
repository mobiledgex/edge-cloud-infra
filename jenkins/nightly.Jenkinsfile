def BUILD_TAG = 'latest'

pipeline {
    options {
        timeout(time: 30, unit: 'MINUTES')
    }
    agent any
    stages {
        stage('Set up build tag') {
            steps {
                script {
                    BUILD_TAG = sh(returnStdout: true, script: 'date +"%Y-%m-%d"')
                }
            }
        }
        stage('Checkout') {
            steps {
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud-infra') {
                    checkout scm
                }
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud') {
                    git url: 'git@github.com:mobiledgex/edge-cloud.git'
                }
            }
        }
        stage('Edge-Cloud Version') {
            steps {
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud-infra') {
                    sh label: 'make edge-cloud-version-set', script: '''$!/bin/bash
export PATH=$PATH:$HOME/go/bin:$WORKSPACE/go/bin
export GOPATH=$WORKSPACE/go
make edge-cloud-version-set
                    '''
                }
            }
        }
        stage('Pull in dependencies') {
            steps {
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud') {
                    sh label: 'make dep', script: '''#!/bin/bash
export PATH=$PATH:$HOME/go/bin:$WORKSPACE/go/bin
export GOPATH=$WORKSPACE/go
make dep
                    '''
                }
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud-infra') {
                    sh label: 'infra make dep', script: '''#!/bin/bash
export PATH=$PATH:$HOME/go/bin:$WORKSPACE/go/bin
export GOPATH=$WORKSPACE/go
make dep
                    '''
                }
            }
        }
        stage('Docker Image') {
            steps {
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud') {
                    sh label: 'make build-docker', script: '''#!/bin/bash
TAG="${BUILD_TAG}" make build-docker
                    '''
                }
                script {
                    currentBuild.displayName = sh(returnStdout: true,
                        script: "docker run --rm registry.mobiledgex.net:5000/mobiledgex/edge-cloud:${BUILD_TAG}")
                }
            }
        }
    }
    post {
        success {
            slackSend color: 'good', message: "Build Successful - ${env.JOB_NAME} #${env.BUILD_NUMBER} (<${env.RUN_DISPLAY_URL}|Open>)"
        }
        failure {
            slackSend color: 'warning', message: "Build Failed - ${env.JOB_NAME} #${env.BUILD_NUMBER} (<${env.RUN_DISPLAY_URL}|Open>)"
        }
    }
}

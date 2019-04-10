pipeline {
    options {
        timeout(time: 30, unit: 'MINUTES')
    }
    agent any
    stages {
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
        stage('Build') {
            steps {
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud') {
                    sh label: 'make clean', script: '''#!/bin/bash
export PATH=$PATH:$HOME/go/bin:$WORKSPACE/go/bin
export GOPATH=$WORKSPACE/go
make clean
                    '''
                }
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud') {
                    sh label: 'make dep', script: '''#!/bin/bash
export PATH=$PATH:$HOME/go/bin:$WORKSPACE/go/bin
export GOPATH=$WORKSPACE/go
make dep
                    '''
                }
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud') {
                    sh label: 'make tools', script: '''#!/bin/bash
export PATH=$PATH:$HOME/go/bin:$WORKSPACE/go/bin
export GOPATH=$WORKSPACE/go
make tools
                    '''
                }
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud-infra') {
                    sh label: 'infra make dep', script: '''#!/bin/bash
export PATH=$PATH:$HOME/go/bin:$WORKSPACE/go/bin
export GOPATH=$WORKSPACE/go
make dep
                    '''
                }
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud-infra') {
                    sh label: 'make', script: '''#!/bin/bash
export PATH=$PATH:$HOME/go/bin:$WORKSPACE/go/bin
export GOPATH=$WORKSPACE/go
make
                    '''
                }
            }
        }
        stage('Docker Image') {
            steps {
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud') {
                    sh label: 'make build-docker', script: '''#!/bin/bash
TAG="$( date +'%Y-%m-%d' )" make build-docker
                    '''
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

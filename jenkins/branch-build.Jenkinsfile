pipeline {
    options {
        timeout(time: 30, unit: 'MINUTES')
    }
    agent any
    parameters {
        string(name: 'BRANCH', defaultValue: '', description: 'Branch to build')
    }
    environment {
        DOCKER_BUILD_TAG = """${sh(
            returnStdout: true,
            script: '''
	    	echo -n "${BRANCH}-`date +'%Y-%m-%d'`"
            '''
        )}"""
    }
    stages {
        stage('Set up build tag') {
            steps {
                script {
                    currentBuild.displayName = "${DOCKER_BUILD_TAG}"
                }
            }
        }
        stage('Checkout') {
            steps {
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud-infra') {
                    git url: 'git@github.com:mobiledgex/edge-cloud-infra.git',
		    	branch: "${BRANCH}"
                }
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud') {
                    git url: 'git@github.com:mobiledgex/edge-cloud.git',
		    	branch: "${BRANCH}"
                }
                dir(path: 'go/src/github.com/mobiledgex/edge-proto') {
                    git url: 'git@github.com:mobiledgex/edge-proto.git',
		    	branch: "${BRANCH}"
                }
            }
        }
        stage('Force Clean') {
            steps {
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud') {
                    sh label: 'git clean edge-cloud', script: 'git clean -f -d -x'
                }
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud-infra') {
                    sh label: 'git clean edge-cloud-infra', script: 'git clean -f -d -x'
                }
            }
        }
        stage('Docker Image') {
            steps {
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud') {
                    sh label: 'make build-docker', script: '''#!/bin/bash
TAG="${DOCKER_BUILD_TAG}" make build-docker
                    '''
                }
                script {
                    currentBuild.displayName = sh(returnStdout: true,
                        script: "docker run --rm registry.mobiledgex.net:5000/mobiledgex/edge-cloud:${DOCKER_BUILD_TAG} version")
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

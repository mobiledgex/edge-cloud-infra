pipeline {
    options {
        timeout(time: 30, unit: 'MINUTES')
    }
    agent any
    environment {
        PAGERDUTY_INTEGRATION_KEY = credentials('pagerduty-service-integration-key')
    }
    stages {
        stage('Checkout') {
            steps {
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud-infra') {
                    checkout scm
                }
            }
        }
        stage('Build') {
            steps {
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud-infra/docker') {
                    sh label: 'make', script: '''#!/bin/bash
REGISTRY=harbor.mobiledgex.net/mobiledgex make publish
                    '''
                }
            }
        }
    }
    post {
        success {
            slackSend color: 'good', message: "Build Successful - ${env.JOB_NAME} #${env.BUILD_NUMBER} (<${env.RUN_DISPLAY_URL}|Open>)"
            pagerduty(resolve: true,
                      serviceKey: "${PAGERDUTY_INTEGRATION_KEY}",
                      incidentKey: "jenkins-${env.JOB_NAME}",
                      incDescription: "Build Successful - ${env.JOB_NAME} #${env.BUILD_NUMBER}",
                      incDetails: "${env.RUN_DISPLAY_URL}")
        }
        failure {
            slackSend color: 'warning', message: "Build Failed - ${env.JOB_NAME} #${env.BUILD_NUMBER} (<${env.RUN_DISPLAY_URL}|Open>)"
            pagerduty(resolve: false,
                      serviceKey: "${PAGERDUTY_INTEGRATION_KEY}",
                      incidentKey: "jenkins-${env.JOB_NAME}",
                      incDescription: "Build Failure - ${env.JOB_NAME} #${env.BUILD_NUMBER}",
                      incDetails: "${env.RUN_DISPLAY_URL}")
        }
    }
}

pipeline {
    options {
        timeout(time: 45, unit: 'MINUTES')
    }
    agent any
    stages {
        stage('Checkout') {
            steps {
                dir(path: 'edge-cloud-ui') {
                    git url: 'git@github.com:mobiledgex/edge-cloud-ui.git'
                }
                script {
                    dir(path: 'edge-cloud-ui') {
                        currentBuild.displayName = sh(returnStdout: true,
                            script: "git describe --tags")
                    }
                }
            }
        }
        stage('Docker Image') {
            steps {
                dir(path: 'edge-cloud-ui') {
                    sh "make build && make publish"
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

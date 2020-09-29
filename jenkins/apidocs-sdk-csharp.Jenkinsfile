pipeline {

    options {
        timeout(time: 5, unit: 'MINUTES')
    }
    agent any

    environment {
        BUILD_TAG = sh(returnStdout: true, script: 'date +"%Y-%m-%d" | tr -d "\n"')
    }

    stages {
        stage('Set up build tag') {
            steps {
                script {
                    currentBuild.displayName = "${BUILD_TAG}"
                }
            }
        }
        stage('Checkout') {
            steps {
                dir(path: 'edge-cloud-sdk-csharp') {
                    git url: 'git@github.com:mobiledgex/edge-cloud-sdk-csharp.git'
                }
            }
        }
        stage('Generate docs') {
            steps {
                dir(path: 'edge-cloud-sdk-csharp/rest') {
                    sh 'make generate-doxygen'
                }
            }
        }
        stage('Upload to Artifactory') {
            steps {
                dir(path: 'edge-cloud-sdk-csharp/rest/Doxygen') {
                    rtUpload (
                        serverId: "artifactory",
                        spec:
                            """{
                                "files": [
                                    {
                                        "pattern": "html.zip",
                                        "target": "apidocs/edge-cloud-sdk-csharp/${BUILD_TAG}/html.zip"
                                    }
                                ]
                            }"""
                    )
                }
            }
        }
    }
}

pipeline {
    options {
        timeout(time: 6, unit: 'HOURS')
    }
    agent any
    environment {
        GITLAB_TOKEN = credentials('gitlab-backup-token')
    }
    stages {
        stage('Run the backup job') {
            steps {
                sh 'mgmt/registry/gitlab-backup.py'
            }
        }
    }
}

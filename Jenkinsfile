pipeline {

    agent {
        label 'docker-18.09.6'
    }

    options {
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '10'))
        disableConcurrentBuilds()
        disableResume()
        timeout(time: 5, unit: 'MINUTES')
        timestamps()
    }

    stages {
        stage('Build') {
            steps {
                echo 'Build Stage'
            }
        }
        stage('Test') {
            steps {
                echo 'Test Stage'
            }
        }
        stage('Deploy') {
            steps {
                echo 'Deploy Stage '
            }
        }
    }

    post {
        always {
            echo 'End'
        }
        success {
            echo 'Success'
        }
        failure {
            echo 'Failure'
        }
    }
}

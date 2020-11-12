
pipeline {

    agent {
        label 'docker'
    }

    options {
        skipDefaultCheckout(true)
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '10', artifactNumToKeepStr: '10'))
        disableConcurrentBuilds()
        disableResume()
        timeout(time: 10, unit: 'MINUTES')
        timestamps()
    }

    stages {

        stage('Checkout') {
            steps {
                script {
                    currentBuild.displayName = "#${currentBuild.number} - ? (?)"
                }
                deleteDir()
                checkout(scm)
            }
        }

    }

    post {
        always {
            script {
                echo "stub"
            }
        }
        success {
            cleanWs()
        }
        failure {
            cleanWs()
        }
    }
}

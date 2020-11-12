
def artifactory = Artifactory.server "artifactory"

pipeline {

    agent {
        label 'docker'
    }

    parameters {
        text(defaultValue: null, description: 'version to test', name: 'VERSION')
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

        stage('Setup') {
            steps {
                script {
                    if (params.VERSION == null || params.VERSION == "") {
                        error('missing parameter VERSION')
                    }
                }
            }
        }

        stage('Checkout') {
            steps {
                script {
                    currentBuild.displayName = "#${currentBuild.number} - main (${params.VERSION})"
                }
                deleteDir()
                checkout(scm)
            }
        }

    }

    post {
        always {
            script {
                echo "will perf test ${params.VERSION}"
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

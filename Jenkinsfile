
def dockerOptions() {
    String options = "--pull "
    options += "--label 'org.opencontainers.image.source=${env.GIT_URL}' "
    options += "--label 'org.opencontainers.image.created=${env.RFC3339_DATETIME}' "
    options += "--label 'org.opencontainers.image.revision=${env.GIT_COMMIT}' "
    options += "--label 'org.opencontainers.image.licenses=${env.LICENSE}' "
    options += "--label 'org.opencontainers.image.authors=${env.PROJECT_AUTHOR}' "
    options += "--label 'org.opencontainers.image.title=${env.PROJECT_TITLE}' "
    options += "--label 'org.opencontainers.image.description=${env.PROJECT_DESCRIPTION}' "
    options += "."
    return options
}

pipeline {

    agent {
        label 'master'
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

        stage('Setup') {
            steps {
                script {
                    env.RFC3339_DATETIME = sh(
                        script: 'date --rfc-3339=ns',
                        returnStdout: true
                    ).trim()
                    env.LICENSE = "Apache-2.0"                     // fixme read from sources
                    env.PROJECT_NAME = "Lake"                      // fixme read from sources
                    env.PROJECT_DESCRIPTION = "Lake message relay" // fixme read from sources
                    env.PROJECT_AUTHOR = "Jan Cajthaml <jan.cajthaml@gmail.com>"
                    env.GOPATH = "${WORKSPACE}"
                    env.HOME = "${WORKSPACE}"
                    env.XDG_CACHE_HOME = "${WORKSPACE}/.cache"
                }
            }
        }

        stage('Sync Dependencies') {
            agent {
                docker {
                    image 'jancajthaml/go:latest'
                    reuseNode true
                }
            }
            steps {
                dir("services/lake") {
                    sh "go mod vendor"
                }
            }
        }

        stage('Unit Test') {
            agent {
                docker {
                    image 'jancajthaml/go:latest'
                    reuseNode true
                }
            }
            steps {
                sh "ln -s ${env.HOME}/services/lake ${env.HOME}/github.com/jancajthaml-openbank/lake"

                dir("services/lake") {
                    sh "${env.HOME}/dev/lifecycle/test --pkg lake --output ${env.HOME}/reports"
                    echo sh(
                        script: 'ls -la reports',
                        returnStdout: true
                    ).trim()
                }
            }
        }
    }


    post {
        always {
            echo 'End'
            deleteDir()
        }
        success {
            echo 'Success'
        }
        failure {
            echo 'Failure'
        }
    }
}

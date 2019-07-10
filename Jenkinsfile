
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
    }

    stages {

        ws("${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_ID}/") {
            withEnv(["GOPATH=${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_ID}"]) {

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
                            env.PATH="${GOPATH}/bin:$PATH"
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
                        echo sh(
                            script: 'pwd',
                            returnStdout: true
                        ).trim()

                        echo sh(
                            script: 'ls -la',
                            returnStdout: true
                        ).trim()

                        //sh "mkdir -p /go/src/github.com/jancajthaml-openbank"
                        //sh "ln -s ./services/lake /go/src/github.com/jancajthaml-openbank/lake"
                        //sh "rm go.sum"
                        //sh "go mod vendor"
                    }
                }
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

def DOCKER_IMAGE_AMD64

def dockerOptions() {
    String options = "--pull "
    options += "--label 'org.opencontainers.image.source=${env.GIT_URL}' "
    options += "--label 'org.opencontainers.image.created=${env.RFC3339_DATETIME}' "
    options += "--label 'org.opencontainers.image.revision=${env.GIT_COMMIT}' "
    options += "--label 'org.opencontainers.image.licenses=${env.LICENSE}' "
    options += "--label 'org.opencontainers.image.authors=${env.PROJECT_AUTHOR}' "
    options += "--label 'org.opencontainers.image.title=${env.PROJECT_NAME}' "
    options += "--label 'org.opencontainers.image.description=${env.PROJECT_DESCRIPTION}' "
    options += "."
    return options
}

pipeline {

    agent {
        label 'docker'
    }

    options {
        skipDefaultCheckout true
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '10', artifactNumToKeepStr: '10'))
        disableConcurrentBuilds()
        disableResume()
        timeout(time: 4, unit: 'HOURS')
        timestamps()
    }

    stages {

        stage('Checkout') {
            steps {
                deleteDir()
                checkout scm
            }
        }

        stage('Setup') {
            steps {
                script {
                    env.RFC3339_DATETIME = sh(
                        script: 'date --rfc-3339=ns',
                        returnStdout: true
                    ).trim()

                    env.VERSION_MAIN = sh(
                        script: 'git fetch --tags --force 2> /dev/null; tags=\$(git tag --sort=-v:refname | head -1) && ([ -z \${tags} ] && echo v0.0.0 || echo \${tags})',
                        returnStdout: true
                    ).trim() - 'v'

                    env.VERSION_META = sh(
                        script: "echo ${env.BRANCH_NAME} | sed 's:.*/::'",
                        returnStdout: true
                    ).trim()

                    env.LICENSE = "Apache-2.0"                           // fixme read from sources
                    env.PROJECT_NAME = "openbank lake"                   // fixme read from sources
                    env.PROJECT_DESCRIPTION = "OpenBanking lake service" // fixme read from sources
                    env.PROJECT_AUTHOR = "Jan Cajthaml <jan.cajthaml@gmail.com>"
                    env.HOME = "${WORKSPACE}"
                    env.GOPATH = "${WORKSPACE}/go"
                    env.XDG_CACHE_HOME = "${env.GOPATH}/.cache"
                    env.PROJECT_PATH = "${env.GOPATH}/src/github.com/jancajthaml-openbank/lake"

                    sh """
                        mkdir -p \
                            ${env.GOPATH}/src/github.com/jancajthaml-openbank && \
                        mv \
                            ${WORKSPACE}/services/lake \
                            ${env.GOPATH}/src/github.com/jancajthaml-openbank/lake
                    """

                    echo "VERSION_MAIN: ${VERSION_MAIN}"
                    echo "VERSION_META: ${VERSION_META}"
                }
            }
        }

        stage('Fetch Dependencies') {
            agent {
                docker {
                    image 'jancajthaml/go:latest'
                    args "--tty --entrypoint=''"
                    reuseNode true
                }
            }
            steps {
                dir(env.PROJECT_PATH) {
                    sh "pwd"

                    sh "ls -la"

                    sh """
                        ${HOME}/dev/lifecycle/sync \
                        --source ${WORKSPACE}/go/src/github.com/jancajthaml-openbank/lake
                    """
                }
            }
        }

        stage('Quality Gate') {
            agent {
                docker {
                    image 'jancajthaml/go:latest'
                    args "--tty --entrypoint=''"
                    reuseNode true
                }
            }
            steps {
                dir(env.PROJECT_PATH) {
                    sh """
                        ${HOME}/dev/lifecycle/lint \
                        --source ${WORKSPACE}/go/src/github.com/jancajthaml-openbank/lake
                    """
                    sh """
                        ${HOME}/dev/lifecycle/sec \
                        --source ${WORKSPACE}/go/src/github.com/jancajthaml-openbank/lake
                    """
                }
            }
        }

        stage('Unit Test') {
            agent {
                docker {
                    image 'jancajthaml/go:latest'
                    args "--tty --entrypoint='' -v ${env.WORKSPACE}:${env.WORKSPACE} -v ${env.WORKSPACE_TMP}:${env.WORKSPACE_TMP}"
                    reuseNode true
                }
            }
            steps {
                dir(env.PROJECT_PATH) {
                    sh """
                        ${HOME}/dev/lifecycle/test \
                        --source ${WORKSPACE}/go/src/github.com/jancajthaml-openbank/lake \
                        --output ${HOME}/reports
                    """
                }
            }
        }

        stage('Package') {
            agent {
                docker {
                    image 'jancajthaml/go:latest'
                    args '--tty'
                    reuseNode true
                }
            }
            steps {
                dir(env.PROJECT_PATH) {
                    sh """
                        ${HOME}/dev/lifecycle/package \
                        --arch linux/amd64 \
                        --source ${WORKSPACE}/go/src/github.com/jancajthaml-openbank/lake \
                        --output ${HOME}/packaging/bin
                    """
                }
                dir(env.PROJECT_PATH) {
                    sh """
                        ${HOME}/dev/lifecycle/debian \
                        --version ${env.VERSION_MAIN}+${env.VERSION_META} \
                        --arch amd64 \
                        --pkg lake \
                        --source ${HOME}/packaging
                    """
                }
            }
        }

        stage('Package Docker') {
            steps {
                script {
                    DOCKER_IMAGE_AMD64 = docker.build("openbank/lake:${env.GIT_COMMIT}", dockerOptions())
                }
            }
        }

    }

    post {
        always {
            script {
                sh "docker rmi -f registry.hub.docker.com/openbank/lake:amd64-${env.VERSION_MAIN}-${env.VERSION_META} || :"
                sh "docker rmi -f lake:amd64-${env.GIT_COMMIT} || :"
            }
            script {
                dir('reports') {
                    archiveArtifacts(
                        allowEmptyArchive: true,
                        artifacts: 'perf-tests/**/*'
                    )
                    archiveArtifacts(
                        allowEmptyArchive: true,
                        artifacts: 'blackbox-tests/**/*'
                    )
                }
                dir('packaging/bin') {
                    archiveArtifacts(
                        allowEmptyArchive: true,
                        artifacts: '*'
                    )
                }
                publishHTML(target: [
                    allowMissing: true,
                    alwaysLinkToLastBuild: false,
                    keepAll: true,
                    reportDir: 'reports/unit-tests',
                    reportFiles: 'lake-coverage.html',
                    reportName: 'Lake | Unit Test Coverage'
                ])
                junit(
                    allowEmptyResults: true,
                    testResults: 'reports/unit-tests/lake-results.xml'
                )
            }
            cleanWs()
        }
        success {
            echo 'Success'
        }
        failure {
            echo 'Failure'
        }
    }
}

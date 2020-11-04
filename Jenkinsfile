def DOCKER_IMAGE_AMD64


def dockerOptions() {
    String options = "--pull "
    options += "--label 'org.opencontainers.image.source=${env.CHANGE_BRANCH}' "
    options += "--label 'org.opencontainers.image.created=${env.RFC3339_DATETIME}' "
    options += "--label 'org.opencontainers.image.revision=${env.GIT_COMMIT}' "
    options += "--label 'org.opencontainers.image.licenses=${env.LICENSE}' "
    options += "--label 'org.opencontainers.image.authors=${env.PROJECT_AUTHOR}' "
    options += "--label 'org.opencontainers.image.title=${env.PROJECT_NAME}' "
    options += "--label 'org.opencontainers.image.description=${env.PROJECT_DESCRIPTION}' "
    options += "."
    return options
}

def getVersion() {
    String[] versions = (sh(
        script: 'git fetch --tags --force 2> /dev/null; tags=\$(git tag --sort=-v:refname | head -1) && ([ -z \${tags} ] && echo v0.0.0 || echo \${tags})',
        returnStdout: true
    ).trim() - 'v').split('\\.')
    String major = versions[0]
    String minor = versions[1]
    Integer patch = Integer.parseInt(versions[2], 10)
    String[] log = sh(
        script: "TZ=UTC git log --pretty='format:%cd,%h' --abbrev=4 --date=format-local:'%Y%m%d,%H%M' | head -1",
        returnStdout: true
    ).trim().split(',')
    String version = "${major}.${minor}.${patch + 1}b${log[0]}${Integer.parseInt(log[1], 10)}${Long.parseLong(log[2], 16)}"
    return version
}

def rtServer = Artifactory.server "artifactory"
def rtDocker = Artifactory.docker server: rtServer
def buildInfo = Artifactory.newBuildInfo()

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
        timeout(time: 30, unit: 'MINUTES')
        timestamps()
    }

    stages {

        stage('Checkout') {
            steps {
                deleteDir()
                //checkout(scm)
            }
        }

        stage('Probe') {
            steps {
                script {
                    rtDocker.pull("${env.ARTIFACTORY_DOCKER_REGISTRY}/docker-local/openbank/lake:1.2.6b20201104102354106", "docker-virtual")
                    //sh "exit 1"
                }
            }
        }

        stage('Setup') {
            steps {
                script {
                    env.RFC3339_DATETIME = sh(
                        script: 'date --rfc-3339=ns',
                        returnStdout: true
                    ).trim()
                    env.GIT_COMMIT = sh(
                        script: 'git log -1 --format="%H"',
                        returnStdout: true
                    ).trim()
                    env.VERSION = getVersion()
                    env.LICENSE = "Apache-2.0"                           // fixme read from sources
                    env.PROJECT_NAME = "openbank lake"                   // fixme read from sources
                    env.PROJECT_DESCRIPTION = "OpenBanking lake service" // fixme read from sources
                    env.PROJECT_AUTHOR = "${CHANGE_AUTHOR_DISPLAY_NAME} <${CHANGE_AUTHOR_EMAIL}>"
                    env.GOPATH = "${env.WORKSPACE}/go"
                    env.XDG_CACHE_HOME = "${env.GOPATH}/.cache"
                    echo "VERSION: ${VERSION}"

                    echo sh(script: 'env|sort', returnStdout: true)

                }
            }
        }

        stage('Ensure images up to date') {
            parallel {
                stage('jancajthaml/go:latest') {
                    steps {
                        script {
                            sh "docker pull jancajthaml/go:latest"
                        }
                    }
                }
                stage('jancajthaml/debian-packager:latest') {
                    steps {
                        script {
                            sh "docker pull jancajthaml/debian-packager:latest"
                        }
                    }
                }
            }
        }

        stage('Fetch Dependencies') {
            agent {
                docker {
                    image 'jancajthaml/go:latest'
                    args "--entrypoint=''"
                    reuseNode true
                }
            }
            steps {
                script {
                    sh """
                        ${env.WORKSPACE}/dev/lifecycle/sync \
                        --source ${env.WORKSPACE}/services/lake
                    """
                }
            }
        }

        stage('Static Analysis') {
            agent {
                docker {
                    image 'jancajthaml/go:latest'
                    args "--entrypoint=''"
                    reuseNode true
                }
            }
            steps {
                script {
                    sh """
                        ${env.WORKSPACE}/dev/lifecycle/lint \
                        --source ${env.WORKSPACE}/services/lake
                    """
                    sh """
                        ${env.WORKSPACE}/dev/lifecycle/sec \
                        --source ${env.WORKSPACE}/services/lake
                    """
                }
            }
        }

        stage('Unit Test') {
            agent {
                docker {
                    image 'jancajthaml/go:latest'
                    args "--entrypoint=''"
                    reuseNode true
                }
            }
            steps {
                script {
                    sh """
                        ${env.WORKSPACE}/dev/lifecycle/test \
                        --source ${env.WORKSPACE}/services/lake \
                        --output ${env.WORKSPACE}/reports/unit-tests
                    """
                }
            }
        }

        stage('Compile') {
            agent {
                docker {
                    image 'jancajthaml/go:latest'
                    args "--entrypoint=''"
                    reuseNode true
                }
            }
            steps {
                script {
                    sh """
                        ${env.WORKSPACE}/dev/lifecycle/package \
                        --arch linux/amd64 \
                        --source ${env.WORKSPACE}/services/lake \
                        --output ${env.WORKSPACE}/packaging/bin
                    """
                }
            }
        }

        stage('Package Debian') {
            agent {
                docker {
                    image 'jancajthaml/debian-packager:latest'
                    args "--entrypoint=''"
                    reuseNode true
                }
            }
            steps {
                script {
                    sh """
                        ${env.WORKSPACE}/dev/lifecycle/debian \
                        --version ${env.VERSION} \
                        --arch amd64 \
                        --pkg lake \
                        --source ${env.WORKSPACE}/packaging
                    """
                }
            }
        }

        stage('Package Docker') {
            steps {
                script {
                    DOCKER_IMAGE_AMD64 = docker.build("${env.ARTIFACTORY_DOCKER_REGISTRY}/docker-local/openbank/lake:${env.VERSION}", dockerOptions())
                }
            }
        }

        stage('Publish to Artifactory') {
            steps {
                script {
                    rtDocker.push(DOCKER_IMAGE_AMD64.imageName(), "docker-virtual")
                }
            }
        }

    }

    post {
        always {
            script {
                if (DOCKER_IMAGE_AMD64 != null) {
                    sh "docker rmi -f ${DOCKER_IMAGE_AMD64.id} || :"
                }
            }
            script {
                dir("${env.WORKSPACE}/packaging/bin") {
                    archiveArtifacts(
                        allowEmptyArchive: true,
                        artifacts: '*'
                    )
                }
                publishHTML(target: [
                    allowMissing: true,
                    alwaysLinkToLastBuild: false,
                    keepAll: true,
                    reportDir: "${env.WORKSPACE}/reports/unit-tests",
                    reportFiles: 'lake-coverage.html',
                    reportName: 'Lake | Unit Test Coverage'
                ])
                junit(
                    checksName: 'Unit Test',
                    allowEmptyResults: true,
                    skipPublishingChecks: true,
                    testResults: "${env.WORKSPACE}/reports/unit-tests/lake-results.xml"
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

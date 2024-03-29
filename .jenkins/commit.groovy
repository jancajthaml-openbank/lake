def DOCKER_IMAGE

def dockerOptions() {
    String options = "--pull "
    options += "--label 'org.opencontainers.image.source=${env.GIT_URL}#${env.GIT_BRANCH}' "
    options += "--label 'org.opencontainers.image.created=${env.RFC3339_DATETIME}' "
    options += "--label 'org.opencontainers.image.revision=${env.GIT_COMMIT}' "
    options += "--label 'org.opencontainers.image.licenses=${env.LICENSE}' "
    options += "--label 'org.opencontainers.image.authors=${env.PROJECT_AUTHOR}' "
    options += "--label 'org.opencontainers.image.title=${env.PROJECT_NAME}' "
    options += "--label 'org.opencontainers.image.description=${env.PROJECT_DESCRIPTION}' "
    options += "-f packaging/docker/amd64/Dockerfile "
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
    String version = "${major}.${minor}.${patch + 1}"
    return version
}

def artifactory = Artifactory.server "artifactory"

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
                    env.GIT_URL = sh(
                        script: 'git ls-remote --get-url',
                        returnStdout: true
                    ).trim()
                    env.GIT_BRANCH = sh(
                        script: 'git name-rev --name-only HEAD',
                        returnStdout: true
                    ).trim() - 'remotes/origin/'
                    env.ARCH = sh(
                        script: 'dpkg --print-architecture',
                        returnStdout: true
                    ).trim()

                    env.VERSION = getVersion()
                    env.LICENSE = "Apache-2.0"
                    env.PROJECT_NAME = "openbank lake"
                    env.PROJECT_DESCRIPTION = "OpenBanking lake service"
                    env.PROJECT_AUTHOR = "Jan Cajthaml <jan.cajthaml@gmail.com>"

                    currentBuild.displayName = "#${currentBuild.number} - ${env.GIT_BRANCH} (${env.VERSION})"
                }
            }
        }

        stage('Fetch Dependencies') {
            agent {
                docker {
                    image "jancajthaml/rust:${env.ARCH}"
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
                    image "jancajthaml/rust:${env.ARCH}"
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
                    image "jancajthaml/rust:${env.ARCH}"
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
                    image "jancajthaml/rust:${env.ARCH}"
                    args "--entrypoint=''"
                    reuseNode true
                }
            }
            steps {
                script {
                    sh """
                        ${env.WORKSPACE}/dev/lifecycle/package \
                        --arch linux/${env.ARCH} \
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
                        --arch ${env.ARCH} \
                        --pkg lake \
                        --source ${env.WORKSPACE}/packaging
                    """
                }
            }
        }

        stage('BlackBox Test') {
            agent {
                docker {
                    image "jancajthaml/bbtest:${env.ARCH}"
                    args """-u 0"""
                    reuseNode true
                }
            }
            options {
                timeout(time: 5, unit: 'MINUTES')
            }
            steps {
                script {
                    cid = sh(
                        script: 'hostname',
                        returnStdout: true
                    ).trim()
                    options = """
                        |-e VERSION=${env.VERSION}
                        |-e META=jenkins
                        |-e CI=true
                        |--volumes-from=${cid}
                        |-v /var/run/docker.sock:/var/run/docker.sock:rw
                        |-v /var/lib/docker/containers:/var/lib/docker/containers:rw
                        |-v /sys/fs/cgroup:/sys/fs/cgroup:ro
                        |-u 0
                    """.stripMargin().stripIndent().replaceAll("[\\t\\n\\r]+"," ").stripMargin().stripIndent()
                    docker.image("jancajthaml/bbtest:${env.ARCH}").withRun(options) { c ->
                        sh "docker exec -t ${c.id} python3 ${env.WORKSPACE}/bbtest/main.py"
                    }
                }
            }
        }

        stage('Package Docker') {
            steps {
                script {
                    DOCKER_IMAGE = docker.build("${env.ARTIFACTORY_DOCKER_REGISTRY}/docker-local/openbank/lake:amd64-${env.VERSION}.jenkins", dockerOptions())
                }
            }
        }

        stage('Documentation') {
            agent {
                docker {
                    image "jancajthaml/rust:${env.ARCH}"
                    args "--entrypoint=''"
                    reuseNode true
                }
            }
            steps {
                script {
                    sh """
                        ${env.WORKSPACE}/dev/lifecycle/documentation \
                        --source ${env.WORKSPACE}/services/lake \
                        --output ${env.WORKSPACE}/reports/docs
                    """
                }
            }
        }

        stage('Publish') {
            steps {
                script {
                    docker.withRegistry("http://${env.ARTIFACTORY_DOCKER_REGISTRY}", 'jenkins-artifactory') {
                        DOCKER_IMAGE.push()
                    }
                    artifactory.upload spec: """
                    {
                        "files": [
                            {
                                "pattern": "${env.WORKSPACE}/packaging/bin/lake-linux-(*)",
                                "target": "generic-local/openbank/lake/${env.VERSION}/linux/{1}/lake",
                                "recursive": "false"
                            },
                            {
                                "pattern": "${env.WORKSPACE}/packaging/bin/lake_(*)_(*).deb",
                                "target": "generic-local/openbank/lake/{1}/linux/{2}/lake.deb",
                                "recursive": "false"
                            }
                        ]
                    }
                    """
                }
            }
        }
    }

    post {
        always {
            script {
                /*
                publishHTML(target: [
                    alwaysLinkToLastBuild: false,
                    keepAll: true,
                    reportDir: "${env.WORKSPACE}/reports/unit-tests/lake-coverage",
                    reportFiles: '*',
                    reportName: 'Unit Test Coverage (Lake)'
                ])
                */

                publishHTML(target: [
                    alwaysLinkToLastBuild: false,
                    keepAll: true,
                    reportDir: "${env.WORKSPACE}/reports/docs",
                    reportFiles: '*',
                    reportName: 'Documentation (Lake)'
                ])
                cucumber(
                    fileIncludePattern: '*',
                    jsonReportDirectory: "${env.WORKSPACE}/reports/blackbox-tests/cucumber"
                )
            }
        }
        success {
            cleanWs()
        }
        failure {
            dir("${env.WORKSPACE}/reports") {
                archiveArtifacts(
                    allowEmptyArchive: true,
                    artifacts: 'blackbox-tests/**/*.log'
                )
            }
            cleanWs()
        }
    }
}

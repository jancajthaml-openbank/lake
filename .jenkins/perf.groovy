
def artifactory = Artifactory.server "artifactory"

pipeline {

    agent {
        label 'docker'
    }

    parameters {
        string(defaultValue: null, description: 'version to test', name: 'VERSION')
        string(defaultValue: '1000', description: 'number of messages to be relayed', name: 'MESSAGES_RELAYED')
    }

    options {
        skipDefaultCheckout(true)
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '10', artifactNumToKeepStr: '10'))
        disableConcurrentBuilds()
        disableResume()
        timeout(time: 5, unit: 'HOURS')
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
                    currentBuild.displayName = "#${currentBuild.number} - ${params.VERSION}"
                }
                deleteDir()
                checkout(scm)
            }
        }

        stage('Download') {
            steps {
                script {
                    artifactory.download spec: """{
                        "files": [
                            {
                                "flat": true,
                                "pattern": "generic-local/openbank/lake/${params.VERSION}/linux/amd64/lake.deb",
                                "target": "${env.WORKSPACE}/packaging/bin/lake_${params.VERSION}_amd64.deb"
                            }
                        ]
                    }"""
                }
            }
        }

        stage('Performance Test') {
            agent {
                docker {
                    image "jancajthaml/bbtest:amd64"
                    args """-u 0"""
                    reuseNode true
                }
            }
            steps {
                script {
                    cid = sh(
                        script: 'hostname',
                        returnStdout: true
                    ).trim()
                    options = """
                        |-e VERSION=${params.VERSION}
                        |-e META=jenkins
                        |-e MESSAGES_PUSHED=${params.MESSAGES_RELAYED}
                        |-e CI=true
                        |--volumes-from=${cid}
                        |--cpus=1
                        |--memory=2g
                        |--memory-swappiness=0
                        |-v /var/run/docker.sock:/var/run/docker.sock:rw
                        |-v /var/lib/docker/containers:/var/lib/docker/containers:rw
                        |-v /sys/fs/cgroup:/sys/fs/cgroup:ro
                        |-u 0
                    """.stripMargin().stripIndent().replaceAll("[\\t\\n\\r]+"," ").stripMargin().stripIndent()
                    docker.image("jancajthaml/bbtest:amd64").withRun(options) { c ->
                        sh "docker exec -t ${c.id} python3 ${env.WORKSPACE}/perf/main.py"
                    }
                }
            }
        }

    }

    post {
        success {
            dir("${env.WORKSPACE}/reports") {
                archiveArtifacts(
                    allowEmptyArchive: true,
                    artifacts: 'perf-tests/graphs/*.png'
                )
            }
            cleanWs()
        }
        failure {
            dir("${env.WORKSPACE}/reports") {
                archiveArtifacts(
                    allowEmptyArchive: true,
                    artifacts: 'perf-tests/**/*'
                )
            }
            cleanWs()
        }
    }
}

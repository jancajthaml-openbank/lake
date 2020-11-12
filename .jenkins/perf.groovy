
def artifactory = Artifactory.server "artifactory"

pipeline {

    agent {
        label 'docker'
    }

    parameters {
        string(defaultValue: null, description: 'version to test', name: 'VERSION')
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
                    currentBuild.displayName = "#${currentBuild.number} - ${params.VERSION}"
                }
                deleteDir()
                checkout(scm)
            }
        }

        stage('Download') {
            steps {
                script {
                    echo "will perf test ${params.VERSION}"

                    sh "mkdir -p ${env.WORKSPACE}/packaging/bin"

                    echo "${env.WORKSPACE}/packaging/bin"
                    echo "before download"
                    sh "ls -lFa ${env.WORKSPACE}/packaging/bin"

                    artifactory.download spec: """{
                        "flat": "true",
                        "files": [
                            {
                                "pattern": "generic-local/openbank/lake/${params.VERSION}/linux/amd64/lake.deb",
                                "target": "${env.WORKSPACE}/packaging/bin/lake_${params.VERSION}_amd64.deb"
                            }
                        ]
                    }"""

                    echo "after download"
                    sh "ls -lFa ${env.WORKSPACE}/packaging/bin"
                }
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

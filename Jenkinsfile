
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

def safeNode(label=null, Closure body) {
    node(label) {
        def path = pwd()
        def branchName = env.BRANCH_NAME
        if (branchName) {
            path = path.split('/')
            def workspaceRoot = path[0..<-1].join('/')
            def currentWs = path[-1]
            // Here is where we make branch names safe for directories -
            // the most common bad character is '/' in 'feature/add_widget'
            // which gets replaced with '%2f', so JOB_NAME will be
            // ${PROJECT_NAME}%2f${BRANCH_NAME}
            def newWorkspace = env.JOB_NAME.replace('/','-')
            newWorkspace = newWorkspace.replace('%2f', '-')
            newWorkspace = newWorkspace.replace('%2F', '-')

            // Add on the '@n' suffix if it was there
            if (currentWs =~ '@') {
                newWorkspace = "${newWorkspace}@${currentWs.split('@')[-1]}"
            }
            path = "${workspaceRoot}/${newWorkspace}"
        }
        ws(path) {
            body()
        }
    }
}

safeNode {

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
}

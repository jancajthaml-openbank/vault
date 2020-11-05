def DOCKER_IMAGE

def dockerOptions() {
    String options = "--pull "
    options += "--label 'org.opencontainers.image.source=${env.GIT_URL}#${env.CHANGE_BRANCH}' "
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
                    currentBuild.displayName = "#${currentBuild.number} - ${env.CHANGE_BRANCH} (?)"
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
                    env.ARCH = sh(
                        script: 'dpkg --print-architecture',
                        returnStdout: true
                    ).trim()

                    env.VERSION = getVersion()
                    env.LICENSE = "Apache-2.0"
                    env.PROJECT_NAME = "openbank vault"
                    env.PROJECT_DESCRIPTION = "OpenBanking vault service"
                    env.PROJECT_AUTHOR = "${env.CHANGE_AUTHOR_DISPLAY_NAME} <${env.CHANGE_AUTHOR_EMAIL}>"
                    env.GOPATH = "${env.WORKSPACE}/go"
                    env.XDG_CACHE_HOME = "${env.GOPATH}/.cache"

                    currentBuild.displayName = "#${currentBuild.number} - ${env.CHANGE_BRANCH} (${env.VERSION})"
                }
            }
        }

        stage('Ensure images up to date') {
            parallel {
                stage('jancajthaml/go') {
                    steps {
                        script {
                            sh "docker pull jancajthaml/go:latest"
                        }
                    }
                }
                stage('jancajthaml/debian-packager') {
                    steps {
                        script {
                            sh "docker pull jancajthaml/debian-packager:latest"
                        }
                    }
                }
                stage('jancajthaml/bbtest') {
                    steps {
                        script {
                            sh "docker pull jancajthaml/bbtest:${env.ARCH}"
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
                        --source ${env.WORKSPACE}/services/vault-rest
                    """
                }
                script {
                    sh """
                        ${env.WORKSPACE}/dev/lifecycle/sync \
                        --source ${env.WORKSPACE}/services/vault-unit
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
                        --source ${env.WORKSPACE}/services/vault-rest
                    """
                    sh """
                        ${env.WORKSPACE}/dev/lifecycle/sec \
                        --source ${env.WORKSPACE}/services/vault-rest
                    """
                }
                script {
                    sh """
                        ${env.WORKSPACE}/dev/lifecycle/lint \
                        --source ${env.WORKSPACE}/services/vault-unit
                    """
                    sh """
                        ${env.WORKSPACE}/dev/lifecycle/sec \
                        --source ${env.WORKSPACE}/services/vault-unit
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
                        --source ${env.WORKSPACE}/services/vault-rest \
                        --output ${env.WORKSPACE}/reports/unit-tests/vault-rest
                    """
                }
                script {
                    sh """
                        ${env.WORKSPACE}/dev/lifecycle/test \
                        --source ${env.WORKSPACE}/services/vault-unit \
                        --output ${env.WORKSPACE}/reports/unit-tests/vault-unit
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
                        --arch linux/${env.ARCH} \
                        --source ${env.WORKSPACE}/services/vault-rest \
                        --output ${env.WORKSPACE}/packaging/bin
                    """
                }
                script {
                    sh """
                        ${env.WORKSPACE}/dev/lifecycle/package \
                        --arch linux/${env.ARCH} \
                        --source ${env.WORKSPACE}/services/vault-unit \
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
                        --pkg vault \
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
                        |-e IMAGE_VERSION=${env.VERSION}
                        |-e UNIT_VERSION=${env.VERSION}
                        |-e UNIT_ARCH=${env.ARCH}
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
                    DOCKER_IMAGE = docker.build("${env.ARTIFACTORY_DOCKER_REGISTRY}/docker-local/openbank/vault:${env.VERSION}", dockerOptions())
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
                                "pattern": "${env.WORKSPACE}/packaging/bin/vault-rest-linux-(*)",
                                "target": "generic-local/openbank/vault/linux/{1}/${env.VERSION}/vault-rest",
                                "recursive": "false"
                            },
                            {
                                "pattern": "${env.WORKSPACE}/packaging/bin/vault-unit-linux-(*)",
                                "target": "generic-local/openbank/vault/linux/{1}/${env.VERSION}/vault-unit",
                                "recursive": "false"
                            },
                            {
                                "pattern": "${env.WORKSPACE}/packaging/bin/vault_(*)_(*).deb",
                                "target": "generic-local/openbank/vault/linux/{2}/{1}/vault.deb",
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
                if (DOCKER_IMAGE != null) {
                    sh "docker rmi -f ${DOCKER_IMAGE.id} || :"
                }
            }
            script {
                publishHTML(target: [
                    allowMissing: true,
                    alwaysLinkToLastBuild: false,
                    keepAll: true,
                    reportDir: "${env.WORKSPACE}/reports/unit-tests/vault-unit",
                    reportFiles: 'vault-unit-coverage.html',
                    reportName: 'Unit Test Coverage (Vault Unit)'
                ])
                publishHTML(target: [
                    allowMissing: true,
                    alwaysLinkToLastBuild: false,
                    keepAll: true,
                    reportDir: "${env.WORKSPACE}/reports/unit-tests/vault-rest",
                    reportFiles: 'vault-rest-coverage.html',
                    reportName: 'Unit Test Coverage (Vault Rest)'
                ])
                junit(
                    allowMissing: true,
                    checksName: 'Unit Test (Vault Unit)',
                    allowEmptyResults: true,
                    skipPublishingChecks: true,
                    testResults: "${env.WORKSPACE}/reports/unit-tests/vault-unit/vault-unit-results.xml"
                )
                junit(
                    allowMissing: true,
                    checksName: 'Unit Test (Vault Rest)',
                    allowEmptyResults: true,
                    skipPublishingChecks: true,
                    testResults: "${env.WORKSPACE}/reports/unit-tests/vault-rest/vault-rest-results.xml"
                )
                cucumber(
                    reportTitle: 'Black Box Test',
                    allowMissing: true,
                    fileIncludePattern: '*',
                    jsonReportDirectory: "${env.WORKSPACE}/reports/blackbox-tests/cucumber"
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
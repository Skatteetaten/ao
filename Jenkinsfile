#!/usr/bin/env groovy

def openshift
def git
def yarn
def go

def scriptVersion='v3.6.1'
fileLoader.withGit('https://git.aurora.skead.no/scm/ao/aurora-pipeline-scripts.git', scriptVersion) {
    go = fileLoader.load('go/go')
    git = fileLoader.load('git/git')
    yarn = fileLoader.load('node.js/yarn')
    openshift = fileLoader.load('openshift/openshift')
}

node {

    stage('Checkout') {
        checkout scm
    }

    stage('Build, Test & coverage') {
        go.buildGoWithJenkinsSh("Go 1.9")
    }

    stage('Copy ao to assets') {
        sh 'mkdir -p ./website/public/assets/macos'
        sh 'mkdir -p ./website/public/assets/windows'
        sh './.go/bin/ao version --json >> ./website/public/assets/version.json'
        sh 'cp ./.go/bin/ao ./website/public/assets'
        sh 'cp ./.go/bin/darwin_amd64/ao ./website/public/assets/macos'
        sh 'cp ./.go/bin/windows_amd64/ao.exe ./website/public/assets/windows'
    }

    String version = git.getTagFromCommit()
    currentBuild.displayName = "${version} (${currentBuild.number})"

    dir('website') {
        stage('Deploy to Nexus') {
          yarn.deployToNexus(version)
        }

        stage('OpenShift Build') {
            artifactId = yarn.getArtifactId()
            groupId = yarn.getGroupId()
            openshift.buildWebleveransepakke(artifactId, groupId, version)
        }
    }

    stage('Clear workspace') {
      step([$class: 'WsCleanup'])
    }
}



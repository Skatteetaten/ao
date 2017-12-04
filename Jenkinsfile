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
        sh 'mkdir ./website/public/assets'
        sh './bin/amd64/ao version --json > ./website/public/assets/version.json'
        sh 'cp ./bin/amd64/ao ./website/public/assets'
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



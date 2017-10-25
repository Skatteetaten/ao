#!/usr/bin/env groovy

def openshift
def git
def yarn
def go

def scriptVersion='v4.0.1'
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
        go.buildGoWithJenkinsSh()
    }

    stage('Copy ao to assets') {
        sh 'cd website'
        sh 'mkdir assets'
        sh 'cp ../bin/amd64/ao ./assets'
    }

    def isMaster = env.BRANCH_NAME == "master"
    String version = git.getTagFromCommit()
    currentBuild.displayName = "${version} (${currentBuild.number})"

    stage('Deploy to Nexus') {
      yarn.deployToNexus(version)
    }

    stage('OpenShift Build') {
        artifactId = yarn.getArtifactId()
        groupId = yarn.getGroupId()
        openshift.buildWebleveransepakke(artifactId, groupId, version)
    }
}



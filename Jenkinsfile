#!/usr/bin/env groovy

def openshift
def git
def yarn
def go

def scriptVersion='v3.5.0'
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
        sh 'mkdir ./website/assets'
        sh 'cp ./bin/amd64/ao ./website/assets'
    }

    def isMaster = env.BRANCH_NAME == "master"
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
}



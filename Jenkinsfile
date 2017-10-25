#!/usr/bin/env groovy

def openshift
def git
def npm
def go

def version='v4.0.1'
fileLoader.withGit('https://git.aurora.skead.no/scm/ao/aurora-pipeline-scripts.git', version) {
    go = fileLoader.load('go/go')
    git = fileLoader.load('git/git')
    npm = fileLoader.load('node.js/npm')
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
        dir 'website'
        sh 'mkdir assets'
        sh 'cp /home/$USER/go/src/github.com/skatteetaten/ao/bin/amd64/ao ./assets'
    }

    def isMaster = env.BRANCH_NAME == "master"
    String version = git.getTagFromCommit()
    currentBuild.displayName = "${version} (${currentBuild.number})"
    if (isMaster) {
      if (!git.tagExists("v${version}")) {
        error "Commit is not tagged. Aborting build."
      }

      npm.version(version)
    }

    stage('Deploy to Nexus') {
      yarn.deployToNexus(version)
    }

    stage('OpenShift Build') {
        artifactId = yarn.getArtifactId()
        groupId = yarn.getGroupId()
        openshift.buildWebleveransepakke(artifactId, groupId, version)
    }
}



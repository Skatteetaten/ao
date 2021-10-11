#!/usr/bin/env groovy

def overrides = [
    scriptVersion  : 'v7',
    pipelineScript: 'https://git.aurora.skead.no/scm/ao/aurora-pipeline-scripts.git',
    credentialsId: "github",
    checkstyle : false,
    jiraFiksetIKomponentversjon: true,
    chatRoom: "#aos-notifications",
    iq: false,
    sonarQube: false,
    openshiftBaseImageVersion: "1.7.0", 
    nodeVersion: "12",
    applicationType: "nodejs",
    versionStrategy: [
      [ branch: 'master', versionHint: '4' ]
    ]
]

def git
def jira
def npm
def go
def openshift
def utilities
def properties
def maven

fileLoader.withGit(overrides.pipelineScript, overrides.scriptVersion) {
  git = fileLoader.load('git/git')
  go = fileLoader.load('go/go')
  jira = fileLoader.load('jira/jira')
  npm = fileLoader.load('node.js/npm')
  openshift = fileLoader.load('openshift/openshift')
  maven = fileLoader.load('maven/maven')
  utilities = fileLoader.load('utilities/utilities')
  properties = fileLoader.load('utilities/properties')
}

Map props = properties.getDefaultProps(overrides)
timestamps {
  node(props.slaveSelector) {
    try {
      stage('Clean Workspace') {
        deleteDir()
        sh 'ls -lah'
      }

      stage('Checkout') {
          checkout scm
      }

      stage("Prepare") {
        dir('website') {
          props.gav = npm.getGav(props)
        }
        utilities.initProps(props, git)

        if (props.nodeVersion) {
          echo 'Using Node version: ' + props.nodeVersion
          npm.setVersion(props.nodeTool)
        }
        utilities.preActions(props)

        if (props.gitBranchName) {
          if (props.isReleaseBuild) {
            if ('aurora-nexus' == props.deployTo && utilities.existInRepository(props.repositoryArtifactUrl)) {
              error "Version already exists in Nexus - aborting job"
            }
            if ('maven-central' == props.deployTo && utilities.existInRepository(props.repositoryArtifactUrl)) {
              error "Version already exists in Maven Central - aborting job"
            }
          }
        }
      }

      if (props.isReleaseBuild && !props.tagExists) {
        stage("Tag") {
          git.tagAndPush(props.credentialsId, "v$props.version")
        }
      }

      stage('Build, Test & coverage') {
        go.buildGoWithJenkinsShUsingGlobalTools("go-1.17")
      }

      stage('Copy ao to assets') {
        sh 'mkdir -p ./website/public/assets/macos'
        sh 'mkdir -p ./website/public/assets/windows'
        sh './.go/bin/linux_amd64/ao version --json --autoanswer-recreate-config n > ./website/public/assets/version.json'
        sh 'cp ./.go/bin/linux_amd64/ao ./website/public/assets'
        sh 'cp ./.go/bin/darwin_amd64/ao ./website/public/assets/macos'
        sh 'cp ./.go/bin/windows_amd64/ao.exe ./website/public/assets/windows'
      }

      dir('website') {
        npm.run("cache verify")
        npm.run("ci")

        if ('aurora-nexus' == props.deployTo) {
          stage('Deploy to Nexus') {
            npm.pack()
            npm.deployToNexus(props.version, props.deliveryBundleClassifier)
          }
        }

        openshift.buildAndTest(props, utilities, npm, maven)
      }

      if (props.jiraFiksetIKomponentversjon && props.isReleaseBuild) {
        jira.updateJira(props)
      }

      stage('Clear workspace') {
        cleanWs()
      }
    } catch (InterruptedException e) {
      currentBuild.result="ABORTED"
      throw e
    } catch (e) {
      currentBuild.result = "FAILURE"
      echo "Failure ${e.message}"
      throw e
    } finally {
      utilities.postActions(props)
    }
  }
}

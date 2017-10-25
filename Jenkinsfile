#!/usr/bin/env groovy

node {

    stage 'Load shared libraries'

    def openshift, git
    def version='v4.0.1'
    fileLoader.withGit('https://git.aurora.skead.no/scm/ao/aurora-pipeline-scripts.git', version) {
        openshift = fileLoader.load('openshift/openshift')
        git = fileLoader.load('git/git')
        go = fileLoader.load('go/go')
        webleveransepakke = fileLoader.load('templates/webleveransepakke')
    }

    stage 'Checkout'
    checkout scm


    stage 'Test og coverage'
    go.buildGoWithJenkinsSh()

    dir 'website'
    sh 'mkdir assets'
    sh 'cp /home/$USER/go/src/github.com/skatteetaten/ao/bin/amd64/ao ./assets'

    def overrides = [
      publishToNpm: false,
      deployToNexus: true,
      openShiftBuild: true
    ]

    webleveransepakke(version, overrides)
}



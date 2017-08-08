#!/usr/bin/env bash
#
# Usage: ./distribute <persistent-volume-name>
#
# Copy a release version of aoc to a Persistent Volume bound to a Persistent Volume Claim used by
# the aoc-update-service apache http server.
# The script checks that the given persistent-volume-name is actually bounded to the update service
#
# Prerequisites:
#   ssh login to an OpenShift node where the correct Perstent Volume is mounted
#   sudo privileges on the OpenShift node to be able to copy the files to the volume
#   oc login to a user with access to the OpenShift project runnint the aoc-update-service
#   The OpenShift user must be named the same as the logged-in linux user
#
# Check parameters
#
pv=$1
if [ -z $pv ]; then
  echo "ERROR: Missing Volume name"
  exit -1
fi
#
# Set nodename on OpenShift node used to populate the PV
#
openshiftnode=uil0paas-utv-node01
openshiftproject=paas-aoc-update
openshiftpvbasedir=/shared/pv/recyclable
#
# Related constants
#
pvcname=aoc-update-htdocs
aocrelease=/home/$USER/go/src/github.com/skatteetaten/ao/bin/amd64/aoc
releaseinfo=releaseinfo.json
tmpreleaseinfo=/tmp/$releaseinfo
remotedir=uil0paas-utv-node01:/home/$USER/aoc-v5
#
# Check for valid oc login
#
ocuser=$(oc whoami 2>/dev/null)
if [ "$ocuser" != "$USER" ]; then
  echo "ERROR: Not logged in as current user"
  exit -1
fi
#
# Check for valid OpenShift Project
#
count=$(oc project $openshiftproject 2>/dev/null | grep $openshiftproject | wc -l)
if [ $count == 0 ]; then
  echo "ERROR: OpenShift project $openshiftproject not available"
  exit
fi
#
# Check that the volume is actually bounded to the correct pvc
#
count=$(oc get pvc 2>/dev/null | grep $pv | grep $pvcname | wc -l)
if [ $count == 0 ]; then
  echo "ERROR: Volume $pv not bound to PVC $pvcname"
  exit -1
fi
#
# Get filename and releaseinfo
#
filename=$($aocrelease version -o filename)
$aocrelease version -o json >$tmpreleaseinfo
#
# Copy files to temporary folder on OpenShift node
#
ssh $openshiftnode "mkdir -p ~/aoc-v5"
scp $aocrelease $remotedir/aoc
scp $tmpreleaseinfo $remotedir/$releaseinfo
#
# Copy the files to the actual volume
#
ssh $openshiftnode "sudo cp ~/aoc-v5/ao $openshiftpvbasedir/$pv/$filename"
ssh $openshiftnode "sudo cp ~/aoc-v5/ao $openshiftpvbasedir/$pv/"
ssh $openshiftnode "sudo cp ~/aoc-v5/$releaseinfo $openshiftpvbasedir/$pv/"
#
# Clean up the temporary folder
#
ssh $openshiftnode "rm ~/aoc-v5/ao"
ssh $openshiftnode "rm ~/aoc-v5/$releaseinfo"

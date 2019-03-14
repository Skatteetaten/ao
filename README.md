```
           ,ggg,     _,gggggg,_
          dP""8I   ,d8P""d8P"Y8b,
         dP   88  ,d8'   Y8   "8b,dP
        dP    88  d8'    `Ybaaad88P'
       ,8'    88  8P       `""""Y8
       d88888888  8b            d8
 __   ,8"     88  Y8,          ,8P
dP"  ,8P      Y8  `Y8,        ,8P'
Yb,_,dP       `8b, `Y8b,,__,,d8P'
 "Y8P"         `Y8   `"Y8888P"'

```

# What is it?

AO is short for Aurora Oc. Just as OC is a CLI for OpenShift, AO is a CLI for Boober.

_(Boober is our take on how to handle the `wall of yaml` challenge of Kubernetes. It reads configuration files with a given
schemaVersion from a git repo (AuroraConfig) and transforms it into Openshift Objects via a AuroraDeploymentConfiguration.
More info in the Boober project on https://github.com/Skatteetaten/boober)_

AO lets you manipulate Boober configuration files, and initiate deploys to OpenShift.

Example:

```
ao login my-project
ao checkout my-project
cd my-project
vi my-test-env/my-app.json
git add .
git commit -m "updated my-app"
git push
ao apply my-test-ent/my-app
```

# License

AO is licensed under the Apache License Version 2.0

# Build?

{go} is your GOPATH, default /home/\<user>/go
Make requires docker

```
mkdir -p {go}/src/github.com/skatteetaten
cd {go}/src/github.com/skatteetaten
git clone https://github.com/Skatteetaten/ao.git
cd ao
glide install
make

```

Windows and macOS versions are built on Linux. To develop and test
on windows, use the go install command instead of make.

# Dependencies?

```
glide up # Update dependencies. Only run when you change something in glide.yaml
glide install # install dependencies
```

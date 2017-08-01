**_Work In Progress_**
# General
AOC works by connecting to the Boober API.  Authentication is handled by sending an 
OpenShift token in the HTTP header.
The token is obtained by using the OpenShift API. It is also possible to use a token obtained
by the oc command ("oc whoami -t").

### Connect to Boober
By default, aoc will scan for OpenShift clusters with Boober instances using the naming 
conventions adopted by the Tax Authority.  The **login** command will call the OpenShift 
API on each reachable cluster to obtain a token.  The cluster information and tokens are 
 stored in a configuration file in the users home directory called _.aoc.json_.  

**Example config file:**
 
````
{
  "apiCluster": "utv",
  "affiliation": "sat",
  "clusters": [
    {
      "name": "prod",
      "url": "https://prod-master.paas.skead.no:8443",
      "token": "",
      "reachable": false
    },
    {
      "name": "utv",
      "url": "https://utv-master.paas.skead.no:8443",
      "token": "gaK5_e06oYh0zor2ZCPDMTluwdE0GJcevvOZu_N2hcI",
      "reachable": true
    }
  ]
}

````
Commands that manipulate the Boober repository will only call the apiCluster.  The deploy command
will however call all the reachable clusters, and Boober will deploy the applications that
is targeted to its specific cluster.

It is possible to override the url by using either the -l or --localhost flag, or by using 
the --serverapi argument.

# Commands
The AOC commands are shaped after the pattern of the OC commands.
 
All the commands have a --help option to explain the usage and the parameters.

### Common options
All commands have a few common options:
````
      --serverapi string   Override default server API address
      --token string       Token to be used for serverapi connections
  -v, --verbose            Log progress to standard out
````
In addition, there are 

## Available commands

````
  create      Creates a vault or a secret in a vault
  delete      Delete a resource
  deploy      Deploy applications in the current affiliation
  edit        Edit a single configuration file or a secret in a vault
  export      Exports auroraconf, vaults or secrets to one or more files
  get         Retrieves information from the repository
  import      Imports a set of configuration files to the central store.
  login       Login to openshift clusters
  logout      Logout of all connected clusters
  ping        Checks for open connectivity from all nodes in the cluster to a specific ip address and port. 
  update      Check for available updates for the aoc client, and downloads the update if available.
  version     Shows the version of the aoc client

````

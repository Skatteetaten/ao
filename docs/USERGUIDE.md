**_Work In Progress_**
# General
AOC works by connecting to the Boober API.  Authentication is handled by sending an 
OpenShift token in the HTTP header.
The token is obtained by using the OpenShift API.

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

# Command Reference
###Common options
--serverapi <http://\<server>:port>   // Address of the boober service

### Login
Format: 
````
aoc login <affiliation>
````

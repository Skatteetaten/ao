**_Work In Progress_**
# What is AO
A command line interface for the Boober API.
  * Deploy one or more ApplicationId (environment/application) to one or more clusters
  * Manipulate AuroraConfig remotely
  * Support modifying AuroraConfig locally
  * Manipulate vaults and secrets

# General
AOC works by connecting to the Boober API.  Authentication is handled by sending an 
OpenShift token in the HTTP header.
The token is obtained by using the OpenShift API. It is also possible to use a token obtained
by the oc command ("oc whoami -t") if you paste it into the ao config file or use the --token or -t flag available on all commands.

### Connect to Boober
By default, aoc will scan for OpenShift clusters with Boober instances using the naming 
conventions adopted by the Tax Authority.  The **login** command will call the OpenShift 
API on each reachable cluster to obtain a token.  The cluster information and tokens are 
 stored in a configuration file in the users home directory called _.aoc.json_.  

Commands that manipulate the Boober repository will only call the apiCluster.  The deploy command
will however call all the reachable clusters, and Boober will deploy the applications that
is targeted to its specific cluster.

It is possible to override the url by using the hidden --localhost flag on the login command.  Using this flag will connect to a boober instance running on the local machine.  AO will use the token from the current active connection in the configuration file.

# Commands
The AO commands are grouped into a number of categories:
- OpenShift Action Commands: Deploys one or more applicatons
- Remote AuroraConfig Commands: Manipulate the AuroraConfig referenced by the last login command
- Local File Commands: Facilitate working with an AuroraConfig as local files
- Commands: Configuration, login and update
 
All the commands have a --help option to explain the usage and the parameters.

### Common options
All commands have a few common options:
````
  -h, --help           help for ao
  -l, --log string     Set log level. Valid log levels are [info, debug, warning, error, fatal] (default "fatal")
  -p, --pretty         Pretty print json output for log
  -t, --token string   OpenShift authorization token to use for remote commands, overrides login
````



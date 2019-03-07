**_Work In Progress_**

# What is AO

AO is a command line interface for the Boober API. It is used to

- Deploy one or more ApplicationDeploymentRef (environment/application) to one or more clusters
- Manipulate AuroraConfig remotely
- Support modifying AuroraConfig locally
- Manipulate vaults and secrets

# General

AO works by connecting to the Boober API. Authentication is handled by sending an OpenShift token in the HTTP header. The token is obtained by using the OpenShift API. It is also possible to use a token obtained by the oc command ("oc whoami -t") if you paste it into the ao config file or use the --token or -t flag available on all commands.

### Connect to Boober

AO uses the configuration file _.ao.json_ in the users home folder to find connection configuration. If the file does not exist, AO will create it.

By default, ao will scan for OpenShift clusters with Boober instances using the naming conventions adopted by the Tax Authority. The **login** command will call the OpenShift API on each reachable cluster to obtain a token. The cluster information and tokens are stored in the configuration file.

If you run Boober outside of the Tax Authority you will have to edit the configuration file to supply your own cluster definitions.

Commands that manipulate the Boober repository will only call the apiCluster. The current API cluster is stored in the configuration file. The command ao adm clusters will display the configuration. The deploy command will however call all the reachable clusters, and Boober will deploy the applications that is targeted to its specific cluster.

It is possible to override the url by using the hidden --localhost flag on the login command. Using this flag will connect to a boober instance running on the local machine. AO will use the token from the current active connection in the configuration file. This is useful when doing development work on Boober, or trying to run Boober against a cluster with no Boober installed.

# Concepts

The AuroraConfig concepts are documented in the Boober project.

The AO CLI supports two modes of working: Remote and Local.

Using the remote AuroraConfig commands the user is able to directly manipulate an AuroraConfig in the remote Boober repository. The commands include add, delete, edit, set and unset, in addition to the vault command used to manipulate secret vaults.

Using the local file commands the user is able to check out an AuroraConfig as a set of files and folders. She may then edit, add and delete files and folders at will without affecting the remote repository. This is only updated by using the SAVE command. It is possible to validate a local config before saving it using the VALIDATE subcommand.

Vaults can only be manipulated remotely using the vault command.

The DEPLOY command will deploy all or parts of an AuroraConfig to OpenShift. It is possible to limit the deploy to a single application or a single environment.

# Access control

By default, every authenticated OpenShift user has access to every AuroraConfig in the Boober repository running on the OpenShift Cluster.

To restrict this access, it is possible to add a .permissions file to a folder. The file must be a valid json file containing a list of OpenShift groups:

It is possible to restrict access to vaults and secrets, use the vault add-permissions command. This will create a hidden .permissions file in the vault repository.

```
{
  "groups": ["group1", "group2"]
}
```

# Commands

The AO commands are grouped into a number of categories:

- OpenShift Action Commands: Deploys one or more applicatons
- Remote AuroraConfig Commands: Manipulate the AuroraConfig referenced by the last login command
- Local File Commands: Facilitate working with an AuroraConfig as local files
- Commands: Configuration, login and update

All the commands have a --help option to explain the usage and the parameters.

### Common options

All commands have a few common options:

```
  -h, --help           help for ao
  -l, --log string     Set log level. Valid log levels are [info, debug, warning, error, fatal] (default "fatal")
  -p, --pretty         Pretty print json output for log
  -t, --token string   OpenShift authorization token to use for remote commands, overrides login
```

### Environment variables

AO uses the \$EDITOR environment variable to determine which editor to use when editing files. If not set, AO will default to "vim".

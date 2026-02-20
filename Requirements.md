# Version manager for vcluster cli
This project is a version manager for vcluster, simillar to projects: https://github.com/tfutils/tfenv and https://github.com/pyenv/pyenv.

It has to provide a cli called vc-env that allows for managing multiple versions of vcluster cli. More about vcluster cli is avaiable here: https://www.vcluster.com/docs/platform/next/install/quick-start-guide

vc-env has to support below commands:
- help -> for displaying help and all avaiable commands
- list -> that list all installed versions
- list-remote -> that list all avaiable versions of vcluster cli
- init -> that initialize setup for vc-env by creating 
- install -> that install specific version or if not specific install the latest
- shell -> set version of vcluster cli in the current shell. If version not passed print the version of vcluster cli that is set in the shell. If noone set print information and fail.
- local -> set local version of vcluster cli. If version not passed print the version of vcluster cli that is set localy. If noone set print information and fail.
- global -> set global version of vcluster cli. If version not passed print the version of vcluster cli that is set globally. If noone set print information and fail.
- which -> print the full path to the vcluster cli binary where it points to.
- version -> print the version of vc-env cli tool.
- uninstall -> unistall the specific version, if not passed fail, if version passed wasn't installed fail.

By fail I mean information for a invoker and finish with non zero exit code.

## Version priority
1. shell - the highest priority. Set in `VCENV_VERSION` environment variable.
2. local - the medium priority. Set in file `.vcluster-version`.
3. global - the lowest priority. Set in `$VCENV_ROOT/version` file.

## Business logic
When user type `vcluster` in the terminal the command has to be process by some shim that is added to the PATH environemnt variable. That shim has to use proper version of vcluster cli based on the below criteria.
1. shell version set or `VCENV_VERSION` environment variable use that versions.
2. if shell version not set and `.vcluster-version` file present in the current directory or parent directories(like .nvmrc).  
3. if shell nor local versions set then use global version. If global versions is not set than fail.

## VCENV_ROOT directory
It's a directory where vc-env store all it's files.
Commands like:
- list
- install
- shell
- local
- global
- which
- uninstall
Should fail with message that vc-env is not initialized and has to be initialized by calling `vc-env init` when `VCENV_ROOT` is not in env variables.

## Commands

### init
When calling `vc-env init` the `versions` directory is created that hold all downloaded versions. Init commands expects that `VCENV_ROOT` env variable is set. If not show message to the user how to do that. Create shim wrapper for vcluster command.
Examples:
```sh
printenv
#output
VCENV_ROOT=/Users/mmpyro/.vcenv

vc-env init
#output
initialized
```

```sh
printenv
#output

vc-env init
#output
VCENV_ROOT not set. Set VCENV_ROOT for your shell using the below commands. 
echo 'export VCENV_ROOT="$HOME/.vcenv"' >> ~/.bashrc
or
echo 'export VCENV_ROOT="$HOME/.vcenv"' >> ~/.zshrc
```

### list
List all versions keeps in `$VCENV_ROOT/versions` directory.
Examples:
```sh
vc-env list

#produces
0.30.0
0.31.0
0.32.0
```

### list-remote
Inspect `https://github.com/loft-sh/vcluster/releases` releases to fetch versions. Use GitHub API and fetch all versions. Pre releases has to be optional when pass flag `--prerelease`.
Examples:
```sh
vc-env list-remote

#produces
0.30.0
0.31.0
0.32.0
```

```sh
vc-env list-remote --prerelease

#produces
0.30.0
0.31.0
0.32.0
0.32.0-alpha
```

### install
Install version passed as argument and store it in $VCENV_ROOT/versions/version-name. It should auto-detect OS/arch and download correct binary.
Examples: 
```sh
vc-env install 0.31.1
```
fetch binary to $VCENV_ROOT/versions/0.31.1/

Already installed version
```sh
vc-env list
#outout
0.31.1

vc-env install 0.31.1
#outout
version 0.31.1 already installed skipping
```

### uninstall
Uninstall version passed as argument. Remove it from $VCENV_ROOT/versions
Examples:

```sh
vc-env list
#output
0.31.1
0.32.0


vc-env uninstall 0.31.1
#output
version 0.31.1 uninstalled

vc-env list
#output
0.32.0
```

### version
Print vc-env version
Examples:
```sh
vc-env version
#output
0.1.0
```

### which
Return an absolut path to the binary where vcluster version points to.
Examples:
```sh
vc-env list
# output
0.31.0

vc-env local 0.31.0
vc-env which
# output
$VCENV_ROOT/versions/0.31.0/vcluster
```

```sh
vc-env list
# output
0.31.0

vc-env shell 0.31.0
vc-env which
# output
$VCENV_ROOT/versions/0.31.0/vcluster
```

```sh
vc-env list
# output
0.31.0

vc-env global 0.31.0
vc-env which
# output
$VCENV_ROOT/versions/0.31.0/vcluster
```

### shell
Set the VCENV_VERSION in the current shell when version passed if not print the version if set. shell function wraps the call
Examples:
```sh
vc-env shell 0.31.0

printenv
#output
VCENV_VERSION=0.31.0
VCENV_ROOT=/Users/mmpyro/.vcenv
```

```sh
vc-env list
#output
0.31.0

vc-env shell 0.31.0

vc-env shell
#output
0.31.0
```

```sh
vc-env list
#output
0.31.0

vc-env shell 0.32.0
#output
version 0.32.0 not installed.
```

```sh
vc-env shell
#output
no shell version configured
```

### local
create file `.vcluster-version` if not present in the current directory. Modify if present.
Examples:
```sh
vc-env list
#output
0.31.0

vc-env local 0.31.0

vc-env local
#output
0.31.0
```

```sh
vc-env list
#output
0.31.0

vc-env local 0.32.0
#output
version 0.32.0 not installed.
```

```sh
vc-env local
#output
no local version configured for this directory
```

### global
Set global version in `$VCENV_ROOT/version` file. If file doesn't exists create it.
Examples:
```sh
vc-env list
#output
0.31.0

vc-env global 0.31.0

vc-env global
#output
0.31.0
```

```sh
vc-env list
#output
0.31.0

vc-env global 0.32.0
#output
version 0.32.0 not installed.
```

```sh
vc-env global
#output
no global version configured for this directory
```

### help
Prints all avaiable vc-env commands with descriptions.
Examples:
```sh
- help -> for displaying help and all avaiable commands
- list -> that list all installed versions
- list-remote -> that list all avaiable versions of vcluster cli
- init -> that initialize setup for vc-env by creating 
- install -> that install specific version or if not specific install the latest
- shell -> set version of vcluster cli in the current shell. If version not passed print the version of vcluster cli that is set in the shell. If noone set print information and fail.
- local -> set local version of vcluster cli. If version not passed print the version of vcluster cli that is set localy. If noone set print information and fail.
- global -> set global version of vcluster cli. If version not passed print the version of vcluster cli that is set globally. If noone set print information and fail.
- which -> print the full path to the vcluster cli binary where it points to.
- version -> print the version of vc-env cli tool.
- uninstall -> unistall the specific version, if not passed fail, if version passed wasn't installed fail.
```

## Implementation
Program has to be implemented using the best coding practicies.
Functionality has to be cover by unit tests.
For testing if vc-env is working use docker to run container with ubuntu image. Mount the vc-env and 
test all cli commands in the isolated environment.

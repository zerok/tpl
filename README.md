<h1 align="center">TPL</h1>

<p align="center">A simple template-to-stdout renderer with support for various data sources.</p>

<p align="center"><a href="/LICENSE"><img alt="MIT Software License" src="https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square"></a>
<a href="https://saythanks.io/to/zerok"><img alt="Say Thanks!" src="https://img.shields.io/badge/Say%20Thanks-!-1EAEDB.svg"></a></p>

------------------------------------------------------------------------------

The original motivation behind tpl was that we had docker-compose files that
needed to reference things like secrets from an external Vault instance. In
order to do that we needed to know the external IP address inside the
configuration file. While things like [container-juggler][] solve that for
docker-compose files, we wanted a slighly more generic tool for use in other
environments. Being able to access secrets directly from a Vault instance was
also very high on the requirements-list 😉

Let's look at a small example where we have a service that should be able to
talk to a Vault server running on the host system:

```
version: "3"
services:
    core-service:
        external_hosts:
            - "vault:{{ .Network.ExternalIP }}"
        ...
```

tpl now parses that docker-compose file, attempts to set all the variables it
find in it (see the reference down below for details on what variables and
functions are available), and writes its output to standard-output:

```
$ tpl docker-compose.yml
version: "3"
services:
    core-service:
        external_hosts:
            - "vault:123.123.123.123"
        ...
```

Depending on what data points you want to use inside your template you might
also have to set specific environment variables (like `VAULT_ADDR` and
`VAULT_TOKEN` for interacting with a Vault server). You can find details
for what datapoints are available and what settings they might require in the
next chapter.


## Installation

You have a couple of options to install tpl:

* If you're on macOS and are using [brew](https://brew.sh/):
  
  ```
  $ brew tap zerok/main https://github.com/zerok/homebrew-tap
  $ brew install zerok/main/tpl
  ```

* If you want to install tpl manually, you can find binaries for all releases
  on [Github](https://github.com/zerok/tpl/releases).

* If you have Go installed, you can also install directly from the master
  branch:
  
  ```
  $ go get -u github.com/zerok/tpl/cmd/tpl
  ```


## Supported data points

### System information

#### Platform name and architecture

`{{ .System.OS }}` and `{{ .System.Arch }}` can be used to inspect the
operating system and architecture tpl is executed on.

#### Shell output

Using `{{ .System.ShellOutput "..." }}` you can open a bash shell, run a
command inside of it, and work with the output of that command. Note that, for
now, `/bin/bash` is hardcoded as the path to the shell to be executed.

### Network information

tpl exposes various data points about your local network.

#### Host IP address

You can access the host's external IP address using 
`{{ .Network.ExternalIP }}`. This can be useful to, for instance, configure
host services inside a docker-compose file.

### File-system

#### File existance

Especially when working with docker-compose, being able to check if a certain
file exists before building volume mounts for it:

```
{{ .FS.Exists "path/to/file" }}
```

### Vault secrets

If you have the environment variables `VAULT_ADDR` and `VAULT_TOKEN` set then
you can also access secrets from that Vault using the following syntax:

```
{{ vault "secrets/path" "fieldname" }}
```

To allow for generic templates to be overriden with local path overrides, 
you can specify a custom path prefix prefix for all secrets with the
`--vault-prefix PREFIX` flag.

For more fine-grained mappings, you can also create a mappings file which
maps a path as it is written inside your template to a path as it should be
looked up in the Vault:

```
$ cat vault.tpl
{{ vault "secret/old-path" "field" }}

$ cat vault-mapping.csv
secret/old-path;secret/new-path

$ vault write secret/new-path "field=test-value"
Success! Data written to: secret/new-path

$ tpl vault.tpl --vault-mapping vault-mapping.csv
test-value
```

**Note:** If you also specify a `--vault-prefix`, this will be applied *before*
the path is mapped.


### Data files

You can load data also from previously generated data-files into the template
using the `--data` flag:

```
$ cat test.tpl
{{ range .Data.items }}> .
{{ end }}

$ cat test.yaml
- 1
- 2
- 3

$ tpl --data=items=test.yaml test.tpl
> 1
> 2
> 3
```

Data can be loaded from files using one of these extensions:

- `.json`
- `.yaml`
- `.yml`


## Different template delimiters

The Go template language used `{{` and `}}` as delimiters for working with
variables or actions by default. This can become quite tedious when working
within systems that also use these characters (or at least make the template
hard to read). For this reason, you can override these delimiters using the
`--left-delimiter` and `--right-delimiter` command-line flags.


## And more...

tpl also bundles [sprig](http://masterminds.github.io/sprig/) which offers lots
of general-purpose template functions. Please see its website for details.


## Third-party libraries

This tool wouldn't be possible (or at least would have been a lot harder to
write) without the great work of the Go community. The following libraries are
used:

* https://github.com/Sirupsen/logrus
* https://github.com/fatih/structs
* https://github.com/golang/snappy
* https://github.com/hashicorp/errwrap
* https://github.com/hashicorp/go-cleanhttp
* https://github.com/hashicorp/go-multierror
* https://github.com/hashicorp/go-rootcerts
* https://github.com/hashicorp/hcl
* https://github.com/hashicorp/vault
* https://github.com/Masterminds/sprig
* https://github.com/mitchellh/go-homedir
* https://github.com/mitchellh/mapstructure
* https://github.com/pkg/errors
* https://github.com/sethgrid/pester
* https://github.com/spf13/pflag
* https://golang.org/x/crypto
* https://golang.org/x/net
* https://golang.org/x/sys
* https://golang.org/x/text

Big thanks to everyone who has contributed to any of these projects!

[container-juggler]: https://github.com/sgeisbacher/container-juggler

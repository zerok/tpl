<h1 align="center">TPL</h1>

<p align="center">A simple template-to-stdout renderer with support for various data sources.</p>

<p align="center"><a href="/LICENSE"><img alt="MIT Software License" src="https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square"></a>
<a href="https://saythanks.io/to/zerok"><img alt="Say Thanks!" src="https://img.shields.io/badge/Say%20Thanks-!-1EAEDB.svg"></a></p>

------------------------------------------------------------------------------

## Supported data points

### Network information

tpl exposes various data points about your local network.

#### Host IP address

You can access the host's external IP address using 
`{{ .Network.ExternalIP }}`. This can be useful to, for instance, configure
host services inside a docker-compose file.

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

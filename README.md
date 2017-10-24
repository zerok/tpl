<div style="text-align: center">
<h1 style="text-align: center">TPL</h1>

<p>A simple template-to-stdout renderer with support for various data sources.</p>

<p>
<a href="/LICENSE"><img alt="MIT Software License" src="https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square"></a>
<a href="https://saythanks.io/to/zerok"><img alt="Say Thanks!" src="https://img.shields.io/badge/Say%20Thanks-!-1EAEDB.svg"></a>
</p>

</div>

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

# tpl

A simple template-to-stdout renderer with support for various data sources.

## Supported data points

### Network information

tpl exposes various data points about your local network.

#### Host IP address

You can access the host's external IP address using 
`{{ .Network.ExternalIP }}`. This can be useful to, for instance, configure
host services inside a docker-compose file.

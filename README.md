# Grafana Loki - Multi tenant proxy

This project has been developed to make easy to deploy a [Grafana Loki Server](https://github.com/grafana/loki) in a multi-tenant way.

There is a lot of understanding about how it works.

- [https://github.com/grafana/loki/issues/701](https://github.com/grafana/loki/issues/701)
- [https://github.com/grafana/loki/issues/525](https://github.com/grafana/loki/issues/525)

This works is almost based on [this issue comment](https://github.com/grafana/loki/issues/701#issuecomment-506504372).

## What is it?

It is a basic golang proxy. It does basic auth, logs the requests and serves as a loki reverse proxy. I think this is the very core functionallity needed to manage a multi-tenant service.

Actually, [Grafana loki](https://github.com/grafana/loki) does not check the auth of any request. The multi-tenant mechanism is based in a request header: `X-Scope-OrgID`. So, if you have untrusted tenants, you have to ensure a tenant uses it's own tenant-id/org-id and does not use any id of other tenants.

### Requirements

To use this proxy, you have to configure your [Grafana Loki server](https://github.com/grafana/loki) with `auth_enabled: true` as described in the [offical docs](https://github.com/grafana/loki/blob/v0.3.0/docs/operations.md#multi-tenancy).

Then put the proxy in front of your [Grafana Loki server](https://github.com/grafana/loki) instance, configure the auth proxy configuration, and run it.

### Run it

```bash
$ loki-multi-tenant-proxy run --loki-server http://localhost:3500 --port 3501 --auth-config ./my-auth-config.yaml
```

Where:

- `--port`: Port used to expose this proxy.
- `--loki-server`: URL of your grafana loki instance.
- `--auth-config`: Authentication configuration file path.

#### Configure the proxy

The auth configuration is very simple. Just create a yaml file `my-auth-config.yaml` with the following structure:

```golang
// Authn Contains a list of users
type Authn struct {
	Users []User `yaml:"users"`
}

// User Identifies a user including the tenant
type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	OrgID    string `yaml:"orgid"`
}
```

An example is available at [configs/multiple.user.yaml](configs/multiple.user.yaml) file:

```yaml
users:
  - username: User-a
    password: pass-a
    orgid: tenant-a
  - username: User-b
    password: pass-b
    orgid: tenant-b
```

A tenant can contains multiple users. But a user is tied to a simple tenant.

#### Configure the Grafana Loki clients, promtail

The default promtail configuration does not have any auth definition, so, after deploy this proxy you have to configure the promtail client configuration to point to this reverse proxy instead of pointing to the original grafana loki server.

Then, dont forget to setup your credential configuration. A simple multi-tenant promtail configuration should looks like:

```yaml
server:
  http_listen_port: 9080
  grpc_listen_port: 0
client:
  url: http://loki-multi-tenant-proxy:3501/api/prom/push
  basic_auth:
    username: User-a
    password: pass-a
scrape_configs:
  - job_name: logs
    static_configs:
      - targets:
          - localhost
        labels:
          job: logs
          __path__: /var/logs/*
```

Note the `client` configuration. The original (single tenant) configuration was something similar to:

```yaml
client:
  url: http://loki-server:3500/api/prom/push
```

## Build it

If you want to build it from this repository, follow the instructions bellow:

```bash
$ docker run -it --entrypoint /bin/bash --rm golang:latest
root@6985c5523ed0:/go# git clone https://github.com/angelbarrera92/loki-multi-tenant-proxy.git
Cloning into 'loki-multi-tenant-proxy'...
remote: Enumerating objects: 88, done.
remote: Counting objects: 100% (88/88), done.
remote: Compressing objects: 100% (64/64), done.
remote: Total 88 (delta 26), reused 78 (delta 20), pack-reused 0
Unpacking objects: 100% (88/88), done
root@6985c5523ed0:/go# cd loki-multi-tenant-proxy/cmd/loki-multi-tenant-proxy/
root@6985c5523ed0:/go# go build
go: finding github.com/urfave/cli v1.21.0
go: finding gopkg.in/yaml.v2 v2.2.2
go: finding github.com/BurntSushi/toml v0.3.1
go: finding gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405
go: downloading github.com/urfave/cli v1.21.0
go: downloading gopkg.in/yaml.v2 v2.2.2
go: extracting github.com/urfave/cli v1.21.0
go: extracting gopkg.in/yaml.v2 v2.2.2
root@6985c5523ed0:/go# ./loki-multi-tenant-proxy
NAME:
   Loki Multitenant Proxy - Makes your Loki server multi tenant

USAGE:
   loki-multi-tenant-proxy [global options] command [command options] [arguments...]

VERSION:
   dev

AUTHOR:
   Ángel Barrera - @angelbarrera92

COMMANDS:
   run      Runs the Loki multi tenant proxy
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

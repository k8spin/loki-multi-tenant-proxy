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


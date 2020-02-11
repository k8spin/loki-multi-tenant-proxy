# Kubernetes demo

## Resources

This demo creates three different namespaces:

- `grafana`
- `tenant-1`
- `tenant-2`

### Grafana Namespace

This demo deploys in the `grafana` namespace:

- Grafana
- Loki
- Loki multi-tenant proxy

### Tenant-{0,1}

In each tenant you will found:

- counter
- log-recolector


## Deploy

```bash
$ kubectl apply -f .
```

### Access grafana

```bash
$ kubectl port-forward svc/grafana -n grafana 3000:3000
```

Access `localhost:3000` in a web browser. Login using `admin:admin`.

Create datasources with the following configuration:

```
name: Loki - Tenant1
type: Loki
url: http://loki-multi-tenant-proxy.grafana.svc.cluster.local:3100
basic-auth:
    username: Tenant1
    password: 1tnaneT
```

and

```
name: Loki - Tenant2
type: Loki
url: http://loki-multi-tenant-proxy.grafana.svc.cluster.local:3100
basic-auth:
    username: Tenant2
    password: 2tnaneT
```

Then navigate to explore and see the logs ;)
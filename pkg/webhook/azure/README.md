# Scaledown all databases and instances in the Azure subscription

To reduce monthly bill, this module can scale down Azure Database for MySQL servers and virtual mashines that linked to kubernetes namespace with evening namespace scaledown process.

Add to you kubernetes-manager config:

```yaml
webhooks: 
- provider: azure
  config:
    subscriptionid: "subscriptionid"
    clientid: "clientid"
    clientsecret: "clientsecret"
    tenantid: "tenantid"
  namespace: some-namespace
  cluster: some-cluster
```

this will scale down (and scale up) all databases and instances in the subscription with tags

```yaml
kubernetes-manager-cluster: some-cluster
kubernetes-manager-namespace: some-namespace
```

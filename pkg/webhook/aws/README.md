# Scaledown all databases and instances in the account

To reduce monthly bill, this module can scale down AWS RDS and EC2 instances that linked to kubernetes namespace with evening namespace scaledown process.

Add to you kubernetes-manager config:

```yaml
webhooks: 
- provider: aws
  config:
    accesskeyid: accesskeyid
    accesssecretkey: accesssecretkey
    region: us-east-1
  namespace: some-namespace
  cluster: some-cluster
```

this will scale down (and scale up) all databases and instances in the account with tags

```yaml
kubernetes-manager/cluster: some-cluster
kubernetes-manager/namespace: some-namespace
```

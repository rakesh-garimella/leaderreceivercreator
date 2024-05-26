# Leader Receiver Creator

This is a proof of concept for https://github.com/kyma-project/telemetry-manager/blob/main/docs/contributor/arch/012-leader-receiver-creator.md.

Leader Receiver Creator is a OTel Collector receiver that instantiates another receiver based on the leader election status. It is useful when you want to have a single instance of a receiver running in a cluster.

## How to test

1. Run the following command to deploy the application:

```bash
kubectl apply -f deploy/kube/rbac.yaml
kubectl apply -f deploy/kube/collectors
```

2. Run the following command to check the status of the deployment:

```bash
stern collector -n default
```

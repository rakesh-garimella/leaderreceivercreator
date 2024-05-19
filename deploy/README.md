# How to test the deployment

1. Run the following command to deploy the application:

```bash
kubectl apply -f kube/
```

2. Run the following command to check the status of the deployment:

```bash
stern collector -n default
```

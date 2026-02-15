# Kubernetes Deployment

This directory contains Kubernetes manifests for deploying Cassandra for the Newcord API.

## Architecture

- **Namespace**: `newcord` - Isolates all resources
- **StatefulSet**: `cassandra` - Manages Cassandra pod with persistent storage
- **Service** (Headless): `cassandra` - For StatefulSet pod discovery
- **Service** (NodePort): `cassandra-external` - Exposes Cassandra externally on port 30042

## Quick Start

### Deploy Everything
```bash
kubectl apply -k .
```

Or from the root directory:
```bash
make k8s-deploy
```

### Verify Deployment
```bash
kubectl get all -n newcord
```

You should see:
- Pod: `cassandra-0` (Status: Running, Ready: 1/1)
- Service: `cassandra` (ClusterIP: None)
- Service: `cassandra-external` (Type: NodePort)
- StatefulSet: `cassandra` (Ready: 1/1)

### Wait for Cassandra to be Ready
```bash
kubectl wait --for=condition=ready pod/cassandra-0 -n newcord --timeout=300s
```

## Accessing Cassandra

### Option 1: Port Forward (Recommended for local development)
```bash
kubectl port-forward -n newcord svc/cassandra 9042:9042
```

Then connect from your Go app using:
```
CASSANDRA_HOSTS=localhost
```

### Option 2: NodePort
The Cassandra service is exposed via NodePort on port 30042.

Connect from your Go app using:
```
CASSANDRA_HOSTS=localhost:30042
```

### Option 3: From within the cluster
If running the API inside Kubernetes, use the internal service name:
```
CASSANDRA_HOSTS=cassandra.newcord.svc.cluster.local
```

## Connect to Cassandra Shell (cqlsh)

```bash
kubectl exec -it -n newcord cassandra-0 -- cqlsh
```

Or:
```bash
make k8s-connect
```

## View Logs

```bash
kubectl logs -n newcord cassandra-0 -f
```

Or:
```bash
make k8s-logs
```

## Scaling

To scale Cassandra (not recommended for local dev, but useful for production):

```bash
kubectl scale statefulset cassandra -n newcord --replicas=3
```

Note: You'll need to update the `CASSANDRA_SEEDS` environment variable to include all nodes.

## Storage

The StatefulSet uses a PersistentVolumeClaim template that requests 5Gi of storage.
- Storage class: Uses the default storage class for your cluster
- Access mode: ReadWriteOnce
- Data is persisted even if the pod is deleted

To view persistent volume claims:
```bash
kubectl get pvc -n newcord
```

## Cleanup

### Delete all resources
```bash
kubectl delete -k .
```

Or:
```bash
make k8s-delete
```

Note: This will delete the namespace and all resources, including PersistentVolumeClaims and data.

### Delete only pods (keeps data)
```bash
kubectl delete pod cassandra-0 -n newcord
# StatefulSet will recreate the pod with the same PVC
```

## Troubleshooting

### Pod not starting
```bash
kubectl describe pod cassandra-0 -n newcord
kubectl logs cassandra-0 -n newcord
```

### Check Cassandra status
```bash
kubectl exec -n newcord cassandra-0 -- nodetool status
```

### Check cluster health
```bash
kubectl exec -n newcord cassandra-0 -- nodetool describecluster
```

### PVC issues
```bash
kubectl get pvc -n newcord
kubectl describe pvc cassandra-data-cassandra-0 -n newcord
```

## Resource Limits

Current resource configuration:
- **Requests**: 500m CPU, 1Gi memory
- **Limits**: 1000m CPU, 2Gi memory

Adjust in `cassandra-statefulset.yaml` based on your workload needs.

## Production Considerations

For production deployments, consider:

1. **High Availability**: Deploy at least 3 Cassandra nodes
2. **Storage**: Use a production-grade storage class with proper backup
3. **Resources**: Increase CPU and memory based on load
4. **Monitoring**: Add Prometheus exporters for metrics
5. **Security**: Enable authentication and encryption
6. **Backup**: Implement regular backup strategy
7. **Network Policies**: Restrict network access between pods

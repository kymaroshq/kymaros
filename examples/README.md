# Kymaros Examples

Real-world restore validation scenarios you can deploy on your cluster.

Each example includes:
- The application to back up (Deployment, StatefulSet, etc.)
- A Velero backup schedule
- A Kymaros RestoreTest that validates the restore
- A HealthCheckPolicy tuned to the workload

## Prerequisites

- Kubernetes 1.28+
- Velero installed with a working BackupStorageLocation
- Kymaros installed:
  ```bash
  helm repo add kymaros https://charts.kymaros.io
  helm install kymaros kymaros/kymaros
  ```

## Available examples

| Example | Complexity | What it validates |
|---------|-----------|-------------------|
| [wordpress-mysql](./wordpress-mysql/) | Beginner | Simple web app with a database, PVC binding, HTTP health |
| [postgresql-ha](./postgresql-ha/) | Intermediate | StatefulSet with replication, TCP connectivity, secrets |
| [redis-sentinel](./redis-sentinel/) | Intermediate | Redis with Sentinel HA, multiple pod roles |
| [microservices-app](./microservices-app/) | Advanced | Multi-namespace application with cross-namespace deps |
| [compliance-dora](./compliance-dora/) | Advanced | DORA Article 11 compliance evidence generation |

## Running an example

The flow is the same for every example:

1. Deploy the application
2. Set up a Velero backup schedule
3. Wait for the first backup to complete
4. Deploy the Kymaros HealthCheckPolicy and RestoreTest
5. Trigger a manual run (or wait for the cron)
6. Watch the RestoreReport

See each example's README for the exact commands.

## Contributing examples

Have a real-world scenario that would help others? Open a PR with your example in a new folder following the same structure.

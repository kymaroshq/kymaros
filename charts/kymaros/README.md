# Kymaros

Kubernetes operator for continuous backup restore validation. Kymaros automatically restores backups into isolated sandbox namespaces, runs configurable health checks, scores each restore on a 0-100 scale across six validation layers, and enforces SLA targets for RTO compliance.

## Components

- **Controller** — Go operator that drives the restore test state machine
- **API Server** — REST API exposing test results and triggering ad-hoc runs
- **Dashboard** — React frontend for real-time visibility and compliance reporting

## Quick Install

```bash
helm install kymaros oci://ghcr.io/kymaroshq/kymaros \
  --version 0.6.0 \
  --namespace kymaros-system \
  --create-namespace \
  --set ingress.enabled=true \
  --set ingress.dashboard.host=kymaros.example.com
```

## Key Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `controller.image.repository` | Controller image repository | `ghcr.io/kymaroshq/kymaros` |
| `controller.image.tag` | Controller image tag | `v0.6.0` |
| `controller.replicas` | Controller replica count | `1` |
| `controller.leaderElection.enabled` | Enable leader election | `true` |
| `api.enabled` | Deploy the API server | `true` |
| `api.image.repository` | API server image repository | `ghcr.io/kymaroshq/kymaros-api` |
| `api.image.tag` | API server image tag | `v0.6.0` |
| `dashboard.enabled` | Deploy the React dashboard | `true` |
| `dashboard.image.repository` | Dashboard image repository | `ghcr.io/kymaroshq/kymaros-frontend` |
| `dashboard.image.tag` | Dashboard image tag | `v0.6.0` |
| `ingress.enabled` | Create ingress resources | `false` |
| `ingress.className` | Ingress class name | `nginx` |
| `ingress.dashboard.host` | Dashboard hostname | `kymaros.example.com` |
| `ingress.dashboard.tls.enabled` | Enable TLS for dashboard | `false` |
| `rbac.create` | Create RBAC resources | `true` |
| `rbac.allowExec` | Allow pods/exec (needed for exec health checks) | `true` |
| `metrics.enabled` | Expose Prometheus metrics on :8443 | `true` |
| `metrics.serviceMonitor.enabled` | Create ServiceMonitor | `false` |
| `metrics.prometheusRule.enabled` | Create alerting PrometheusRule | `false` |
| `metrics.grafanaDashboard.enabled` | Create Grafana dashboard ConfigMap | `false` |
| `sandbox.namespacePrefix` | Prefix for sandbox namespaces | `rp-test` |
| `sandbox.defaultTTL` | Default sandbox TTL | `30m` |
| `sandbox.networkIsolation` | Default network isolation mode | `strict` |
| `adapters.velero.namespace` | Namespace where Velero is installed | `velero` |
| `sla.defaultRTOTarget` | Default RTO target | `15m` |
| `sla.alertScoreThreshold` | Alert score threshold | `70` |

## Prometheus Metrics

| Metric | Type | Labels |
|--------|------|--------|
| `kymaros_tests_total` | Counter | `test`, `result` |
| `kymaros_score` | Gauge | `test` |
| `kymaros_rto_seconds` | Gauge | `test` |
| `kymaros_test_duration_seconds` | Histogram | `test` |
| `kymaros_backup_age_seconds` | Gauge | `test` |

## Uninstall

```bash
helm uninstall kymaros -n kymaros-system
kubectl delete namespace kymaros-system
```

CRDs are not deleted by `helm uninstall`. To remove them:

```bash
kubectl delete crd restoretests.restore.kymaros.io
kubectl delete crd restorereports.restore.kymaros.io
kubectl delete crd healthcheckpolicies.restore.kymaros.io
```

## Links

- Documentation: https://docs.kymaros.io
- Website: https://kymaros.io
- Source: https://github.com/kymaroshq/kymaros
- Issues: https://github.com/kymaroshq/kymaros/issues

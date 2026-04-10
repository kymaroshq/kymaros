# Kymaros

Kubernetes operator for continuous backup restore validation. Kymaros automatically restores backups into isolated sandbox namespaces, runs configurable health checks, scores each restore on a 0-100 scale across six validation layers, and enforces SLA targets for RTO compliance.

Single binary: controller, API server, and React dashboard run in one pod.

## Quick Install

```bash
helm repo add kymaros https://charts.kymaros.io
helm install kymaros kymaros/kymaros \
  --version 0.6.7 \
  --namespace kymaros-system \
  --create-namespace
```

## Key Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Container image repository | `ghcr.io/kymaroshq/kymaros` |
| `image.tag` | Image tag (defaults to Chart.appVersion) | `""` |
| `replicas` | Pod replica count | `1` |
| `leaderElection.enabled` | Enable leader election for HA | `true` |
| `ingress.enabled` | Create legacy Ingress resource | `false` |
| `ingress.className` | Ingress class name | `nginx` |
| `gatewayAPI.enabled` | Create HTTPRoute (Gateway API) | `false` |
| `gatewayAPI.hostnames` | Hostnames matched by the route | `[kymaros.example.com]` |
| `gatewayAPI.parentRefs` | Parent Gateway references | `[{name: kymaros-gateway}]` |
| `rbac.create` | Create RBAC resources | `true` |
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

## Exposing the dashboard

Pick one of the two options:

**Gateway API (recommended)** — requires Gateway API CRDs and an existing Gateway:
```yaml
gatewayAPI:
  enabled: true
  hostnames:
    - kymaros.example.com
  parentRefs:
    - name: my-gateway
      namespace: gateway-system
```

**Legacy Ingress** — for clusters without Gateway API:
```yaml
ingress:
  enabled: true
  className: nginx
  host: kymaros.example.com
  tls:
    enabled: true
```

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

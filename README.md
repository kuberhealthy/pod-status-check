# pod-status-check

The `pod-status-check` reports pods that are older than a configured skip duration and are in unhealthy phases (`Pending`, `Failed`, or `Unknown`).

## Configuration

Set these environment variables in the `HealthCheck` spec:

- `SKIP_DURATION` (required): duration to skip newly created pods (for example, `10m`).
- `TARGET_NAMESPACE` (optional): namespace to inspect. Defaults to all namespaces.
- `KUBECONFIG` (optional): explicit kubeconfig path for local development.

When targeting all namespaces, the service account needs cluster-wide permissions for pods.

## Build

- `just build` builds the container image locally.
- `just test` runs unit tests.
- `just binary` builds the binary in `bin/`.

## Example HealthCheck

Apply the example below or the provided `healthcheck.yaml`:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: pod-status-check
  namespace: kuberhealthy
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: pod-status-check
  namespace: kube-system
rules:
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: pod-status-check
  namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: pod-status-check
subjects:
  - kind: ServiceAccount
    name: pod-status-check
    namespace: kuberhealthy
---
apiVersion: kuberhealthy.github.io/v2
kind: HealthCheck
metadata:
  name: pod-status-check
  namespace: kuberhealthy
spec:
  runInterval: 5m
  timeout: 10m
  podSpec:
    spec:
      serviceAccountName: pod-status-check
      containers:
        - name: pod-status-check
          image: kuberhealthy/pod-status-check:sha-<short-sha>
          imagePullPolicy: IfNotPresent
          env:
            - name: TARGET_NAMESPACE
              value: "kube-system"
            - name: SKIP_DURATION
              value: "10m"
          resources:
            requests:
              cpu: 10m
              memory: 50Mi
      restartPolicy: Never
```

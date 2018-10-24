# Charts Check PR Title

This little app was designed specifically for the [charts repo](https://github.com/helm/charts).
It asks people to put their pull request titles in a certain format if they didn't
already do it. It's a little helper for chart maintainers.

## Configuration

The following table lists the configurable parameters of the chart and their default values.

| Parameter                                 | Description                                  | Default                                          |
|-------------------------------------------|----------------------------------------------|--------------------------------------------------|
| `image.repository`                        | Image name                                   | `quay.io/helmpack/charts-check-pr-title`         |
| `image.tag`                               | Image tag                                    | `latest`                                         |
| `image.pullPolicy`                        | Image pull policy                            | `Always`                                         |
| `secrets.existingSecret`                  | Name of existing secret to use if one exists |                                                  |
| `secrets.sharedSecret`                    | Shared secret for the webhook                |                                                  |
| `secrets.repoName`                        | Name of the repo. Used for access control    |                                                  |
| `secrets.ghToken`                         | GitHub token for bot/user to make comment    |                                                  |
| `service.type`                            | Type of Service                              | `ClusterIP`                                      |
| `service.port`                            | Port for kubernetes service                  | `80`                                             |
| `ingress.enabled`                         | Enables Ingress                              | `false`                                          |
| `ingress.annotations`                     | Ingress annotations                          | `{}`                                             |
| `ingress.path`                            | Ingress path                                 | `/`                                              |
| `ingress.hosts`                           | Ingress accepted hostnames                   | `nil`                                            |
| `ingress.tls`                             | Ingress TLS configuration                    | `[]`                                             |
| `resources.requests.cpu`                  | CPU resource requests                        |                                                  |
| `resources.limits.cpu`                    | CPU resource limits                          |                                                  |
| `resources.requests.memory`               | Memory resource requests                     |                                                  |
| `resources.limits.memory`                 | Memory resource limits                       |                                                  |
| `nodeSelector`                            | Settings for nodeselector                    | `{}`                                             |
| `tolerations`                             | Settings for toleration                      | `[]`                                             |
| `affinity`                                | Settings for affinity                        | `{}`                                             |

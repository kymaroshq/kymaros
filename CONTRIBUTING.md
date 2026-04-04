# Contributing to Kymaros

We welcome contributions! Here's how to get started.

## Development Setup

### Prerequisites

- Go 1.25+
- Docker
- A Kubernetes cluster (kind, minikube, or remote)
- Velero installed on the cluster

### Build

```bash
make build        # Build the operator binary
make test         # Run unit tests
make lint         # Run golangci-lint
make manifests    # Regenerate CRD manifests
make generate     # Regenerate deep copy functions
```

### Run locally

```bash
make install      # Install CRDs into the cluster
make run          # Run the controller locally (outside the cluster)
```

## Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Write tests for your changes
4. Ensure `make test` and `make lint` pass
5. Submit a pull request

## Code Style

- Go: follow `gofmt` and the project's `.golangci.yml` configuration
- TypeScript (dashboard): follow the ESLint and Prettier configuration
- Commit messages: use imperative mood ("Add feature" not "Added feature")

## Reporting Bugs

Open an issue on GitHub with:
- Kymaros version (`kubectl get deployment -n kymaros-system -o yaml | grep image:`)
- Kubernetes version (`kubectl version`)
- Velero version (`velero version`)
- Steps to reproduce
- Expected vs actual behavior

## Security Issues

See [SECURITY.md](SECURITY.md) for reporting security vulnerabilities.

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.

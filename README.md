# OpenShift CRC Demo Project 🚀

This project is a Go-based application designed to demonstrate the core features of OpenShift (CodeReady Containers). It provides a hands-on experience with ConfigMaps, Secrets, Persistent Volumes, and Networking.

## 📁 Project Structure

- `app/`: Contains the Go source code.
- `manifests/`: Kubernetes/OpenShift resource definitions (all-in-one).
- `docs/`: Detailed guide and walkthrough for the demo.
- `Dockerfile`: Multi-stage build for the application.

## 🛠️ Features Demonstrated

1. **Source-to-Image (S2I) / Docker Build**: Build automation within OpenShift.
2. **ConfigMaps**: Application configuration management.
3. **Secrets**: Secure management of sensitive data.
4. **Persistent Storage (PVC)**: Data persistence across pod restarts.
5. **Networking (Service & Route)**: Internal and external access management.
6. **Observability**: Downward API for pod metadata.

## 🚀 Getting Started

### 1. Prerequisites
- OpenShift CRC running (`crc status`)
- `oc` CLI installed and logged in (`oc login`)

### 2. Create Project
```bash
oc new-project demo
```

### 3. Build the Image
```bash
oc new-build --name=demo-app --binary --strategy=docker
oc start-build demo-app --from-dir=. --follow
```

### 4. Deploy Resources
```bash
oc apply -f manifests/all-in-one.yaml
```

### 5. Access the App
Get the Route URL:
```bash
oc get route demo-app -o jsonpath='{.spec.host}'
```

## 📄 Documentation

For a step-by-step walkthrough, check out the [Detailed Guide](docs/guide.md).

---
*Created for OpenShift Learning and Demonstration purposes.*

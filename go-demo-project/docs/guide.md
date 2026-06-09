# OpenShift CRC — Core Concepts with Go Demo App

> **Goal:** To learn OpenShift's core services step-by-step via CLI by deploying a Go application from scratch in a local CRC environment.

---

## Table of Contents

1. [Overview](#1-overview)
2. [Prerequisites](#2-prerequisites)
3. [Project Structure](#3-project-structure)
4. [Step 1 — Introduction and Namespace](#4-step-1--introduction-and-namespace)
5. [Step 2 — Image Registry & BuildConfig](#5-step-2--image-registry--buildconfig)
6. [Step 3 — Secret & ConfigMap](#6-step-3--secret--configmap)
7. [Step 4 — PV / PVC / Storage](#7-step-4--pv--pvc--storage)
8. [Step 5 — Deployment & DeploymentConfig](#8-step-5--deployment--deploymentconfig)
9. [Step 6 — Service & Route (Network)](#9-step-6--service--route-network)
10. [Step 7 — Rollout & Updates](#10-step-7--rollout--updates)
11. [Step 8 — Cleanup](#11-step-8--cleanup)
12. [Summary — Concept Map](#12-summary--concept-map)
13. [Quick Reference Table](#13-quick-reference-table)

---

## 1. Overview

This demo application makes the following OpenShift concepts **observable** within a single Go binary:

| Concept | How is it demonstrated? |
|--------|------------------|
| **Image Registry** | Push to internal registry via BuildConfig |
| **BuildConfig** | Source-to-image flow with `oc new-build` |
| **ConfigMap** | `APP_ENV`, `APP_COLOR`, `APP_MESSAGE` env values |
| **Secret** | `DB_USER`, `DB_PASSWORD`, `API_KEY` — hidden in UI |
| **PVC / Storage** | Writes a file under `/data/app/` on every request |
| **Deployment** | Pod count, rollout strategy |
| **Service** | Access to pods via ClusterIP |
| **Route** | External access with Edge TLS |

---

## 2. Prerequisites

```bash
# Verify CRC is running
crc status

# Verify oc CLI is installed
oc version

# Login with CRC developer user
eval $(crc oc-env)
oc login -u developer -p developer https://api.crc.testing:6443
```

> **Note:** The `developer` user in CRC is sufficient for most demos.  
> For admin operations: `oc login -u kubeadmin -p $(cat ~/.crc/machines/crc/kubeadmin-password)`

---

## 3. Project Structure

```
openshift-demo/
├── app/
│   ├── main.go          # Go application
│   └── go.mod
├── manifests/
│   └── all-in-one.yaml  # All K8s/OpenShift resources
└── Dockerfile           # Multi-stage build
```

---

## 4. Step 1 — Introduction and Namespace

### Namespace = Project in OpenShift

In Kubernetes it's called `Namespace`, in OpenShift it's referred to as `Project` — essentially they are the same resource.

```bash
# Create a new project
oc new-project demo

# List existing projects
oc get projects

# Check active project
oc project
```

**What did we learn?**
- `oc new-project` creates both the namespace and the OpenShift Project resource.
- Every resource (Pod, Service, Secret…) belongs to a namespace.
- You can switch active context with `oc project <name>`.

---

## 5. Step 2 — Image Registry & BuildConfig

### 5.1 CRC Internal Image Registry

OpenShift includes its own internal container image registry:  
`image-registry.openshift-image-registry.svc:5000`

```bash
# View the Registry Service
oc get svc -n openshift-image-registry

# ImageStream list — represents images in the registry
oc get imagestream -n demo
```

### 5.2 Image Build with BuildConfig

`BuildConfig` is an OpenShift-specific resource used to generate images from source code (Git, local directory, Dockerfile).

```bash
# Create BuildConfig with Dockerfile in current directory
oc new-build --name=demo-app \
  --binary \
  --strategy=docker \
  -n demo

# Start build from root folder (uses Dockerfile)
oc start-build demo-app \
  --from-dir=. \
  --follow \
  -n demo
```

```bash
# Watch build logs
oc logs -f bc/demo-app -n demo

# View the generated image
oc get imagestream demo-app -n demo

# ImageStream tag details
oc describe imagestream demo-app -n demo
```

**What did we learn?**
- `BuildConfig (bc)` → defines the build recipe.
- `Build (build)` → every `oc start-build` command creates a new Build object.
- `ImageStream (is)` → a pointer to images within OpenShift; actual images are stored in the registry.
- Local files are transferred to CRC and built using `--from-dir`.

---

## 6. Step 3 — Secret & ConfigMap

### What is the difference?

| | ConfigMap | Secret |
|---|---|---|
| **Purpose** | Non-sensitive configuration | Passwords, tokens, certificates |
| **Storage** | Plain text | Base64 encoded (encrypted in etcd) |
| **Example** | `APP_ENV=production` | `DB_PASSWORD=S3cur3P@ss!` |

### 6.1 ConfigMap

```bash
# Create from YAML
oc apply -f manifests/all-in-one.yaml -n demo

# — OR — Create directly via CLI
oc create configmap demo-app-config \
  --from-literal=APP_ENV=production \
  --from-literal=APP_COLOR=red \
  --from-literal=APP_MESSAGE="Hello OpenShift CRC!" \
  --from-literal=ROUTE_URL="http://demo-app-demo.apps-crc.testing" \
  -n demo

# Read ConfigMap content
oc get configmap demo-app-config -o yaml -n demo

# Live change (for hot-reload demo)
oc patch configmap demo-app-config \
  --type merge \
  -p '{"data":{"APP_COLOR":"blue"}}' \
  -n demo
```

### 6.2 Secret

```bash
# Create Secret
oc create secret generic demo-app-secret \
  --from-literal=DB_USER=adminuser \
  --from-literal=DB_PASSWORD='S3cur3P@ss!' \
  --from-literal=API_KEY=myapikey123 \
  -n demo

# List Secret (values are hidden)
oc get secret demo-app-secret -o yaml -n demo

# Decode value (base64)
oc get secret demo-app-secret \
  -o jsonpath='{.data.DB_USER}' -n demo | base64 -d

# Verify Secret as env from within a pod
oc exec deploy/demo-app -n demo -- env | grep DB_
```

**What did we learn?**
- `envFrom.configMapRef` and `envFrom.secretRef` in Pod YAML automatically convert all keys to env vars.
- Secret values require base64 decoding even if shown via `oc get secret`.
- ConfigMap changes do not automatically restart the pod — this requires tools like `Reloader` or manual rollout.

---

## 7. Step 4 — PV / PVC / Storage

### Concepts

```
StorageClass (SC)
    └── PersistentVolume (PV)      ← cluster-level physical disk
            └── PersistentVolumeClaim (PVC)  ← namespace-level request
                        └── Pod Volume Mount  ← /data/app
```

| Resource | Managed by? | Description |
|--------|-------------|----------|
| **StorageClass** | Admin | Defines dynamic provisioner |
| **PV** | Admin / Auto | Actual storage unit |
| **PVC** | Developer | "I want X GB storage" request |

### 7.1 StorageClass and PVC

```bash
# View CRC's default StorageClass
oc get storageclass

# Create PVC (exists in all-in-one.yaml, or via CLI)
oc apply -f manifests/all-in-one.yaml -n demo

# Watch PVC status — wait for it to be Bound
oc get pvc -n demo -w

# View the automatically created PV
oc get pv
```

### 7.2 Demonstrating Storage

```bash
# Look inside /data/app while app is running
oc exec deploy/demo-app -n demo -- ls -la /data/app/

# Test data persistence even if pod is deleted
oc delete pod -l app=demo-app -n demo   # pod deleted, new one comes up
oc exec deploy/demo-app -n demo -- ls -la /data/app/  # files are still there!
```

**What did we learn?**
- `ReadWriteOnce (RWO)`: Pods on a single node can read and write.
- `ReadWriteMany (RWX)`: Pods on multiple nodes can access (like NFS, CephFS).
- Data is preserved even if the pod is deleted, as long as PVC is not deleted.
- `crc-csi-hostpath-provisioner` StorageClass in CRC provides dynamic PV.

---

## 8. Step 5 — Deployment & DeploymentConfig

### Deployment vs DeploymentConfig

| | Deployment (K8s standard) | DeploymentConfig (OpenShift) |
|---|---|---|
| **API** | `apps/v1` | `apps.openshift.io/v1` |
| **Trigger** | Manual / HPA | ImageStream change, ConfigChange |
| **Strategy** | RollingUpdate, Recreate | Rolling, Recreate, Custom |
| **Recommendation** | Preferred for new projects | Legacy, seen in older docs |

```bash
# Apply YAML
oc apply -f manifests/all-in-one.yaml -n demo

# Watch Deployment status
oc rollout status deploy/demo-app -n demo

# View pods
oc get pods -n demo -o wide

# Watch pod logs
oc logs -f deploy/demo-app -n demo

# Enter pod (debug)
oc exec -it deploy/demo-app -n demo -- sh

# Increase replica count
oc scale deploy/demo-app --replicas=3 -n demo
oc get pods -n demo   # 3 pods should appear

# Scale back to 1
oc scale deploy/demo-app --replicas=1 -n demo
```

**What did we learn?**
- `Deployment` defines the desired state; ReplicaSet manages the actual pods.
- `oc rollout status` shows whether deployment is complete.
- `oc scale` provides instant replica change; HPA (HorizontalPodAutoscaler) is used in production.

---

## 9. Step 6 — Service & Route (Network)

### Network Layers

```
Internet / Browser
       │
    [Route]          ← OpenShift-specific, external access
       │ HTTPS
  [HAProxy LB]       ← Automatic in CRC
       │ HTTP
    [Service]        ← ClusterIP, load-balancer
       │
  [Pod : 8080]       ← Application
```

### 9.1 Service

```bash
# Create Service
oc apply -f manifests/all-in-one.yaml -n demo

# Service details
oc get svc demo-app -n demo
oc describe svc demo-app -n demo

# Endpoints — which pod IPs are available?
oc get endpoints demo-app -n demo

# Test from within cluster (from another pod)
oc run test-pod --image=busybox --restart=Never --rm -it \
  -- wget -qO- http://demo-app:8080/healthz
```

### 9.2 Route

```bash
# Create Route (from YAML or CLI)
oc expose svc demo-app \
  --hostname=demo-app-demo.apps-crc.testing \
  -n demo

# List all Routes
oc get routes -n demo

# Get Route URL
oc get route demo-app -o jsonpath='{.spec.host}' -n demo

# Open in browser
open http://$(oc get route demo-app -o jsonpath='{.spec.host}' -n demo)
```

**Add TLS (Edge Termination):**

```bash
oc patch route demo-app \
  --type merge \
  -p '{"spec":{"tls":{"termination":"edge","insecureEdgeTerminationPolicy":"Redirect"}}}' \
  -n demo
```

**What did we learn?**
- `Service (ClusterIP)` makes pods accessible via a stable IP/DNS within the cluster.
- `Route` is OpenShift-specific; equivalent to `Ingress` in Kubernetes.
- `Edge TLS`: SSL ends at CRC's router, pod receives plain HTTP.
- `Passthrough TLS`: SSL is forwarded to the pod, pod manages its own certificate.

---

## 10. Step 7 — Rollout & Updates

### 10.1 Image Update (New Build)

```bash
# Make a change in main.go (e.g., version number)
# Start new build
oc start-build demo-app --from-dir=. --follow -n demo

# Is ImageStream updated?
oc get imagestream demo-app -n demo

# Watch new pods coming up
oc rollout status deploy/demo-app -n demo
oc get pods -n demo
```

### 10.2 ConfigMap Change → Trigger Rollout

```bash
# Change environment variable
oc set env deploy/demo-app APP_COLOR=green -n demo
# This command changes Deployment env, not the ConfigMap directly
# Rollout is automatically triggered because Deployment changed

# — OR — Change ConfigMap + manual rollout
oc patch configmap demo-app-config \
  --type merge \
  -p '{"data":{"APP_COLOR":"purple"}}' \
  -n demo
oc rollout restart deploy/demo-app -n demo
```

### 10.3 Rollback

```bash
# Rollout history
oc rollout history deploy/demo-app -n demo

# Rollback to previous version
oc rollout undo deploy/demo-app -n demo

# Rollback to specific revision
oc rollout undo deploy/demo-app --to-revision=1 -n demo
```

**What did we learn?**
- `oc rollout restart` restarts pods without changing the deployment (ideal for ConfigMap updates).
- `oc rollout undo` provides quick recovery after a faulty deploy.
- `RollingUpdate` strategy ensures zero downtime: old pod is not deleted until new pod is ready.

---

## 11. Step 8 — Cleanup

```bash
# Delete only the deployment
oc delete deploy demo-app -n demo

# Delete all resources (except namespace)
oc delete all -l app=demo-app -n demo

# PVC and Secret must be deleted separately (intentional protection)
oc delete pvc demo-app-pvc -n demo
oc delete secret demo-app-secret -n demo
oc delete configmap demo-app-config -n demo

# Delete the entire project (cleanest way)
oc delete project demo
```

---

## 12. Summary — Concept Map

```
oc new-project demo
        │
        ├── [Image Registry]
        │       └── BuildConfig → Build → ImageStream
        │
        ├── [Configuration]
        │       ├── ConfigMap  (non-sensitive: APP_ENV, APP_COLOR)
        │       └── Secret     (sensitive: DB_PASSWORD, API_KEY)
        │
        ├── [Storage]
        │       └── StorageClass → PV → PVC → Pod Volume Mount
        │
        ├── [Workload]
        │       └── Deployment → ReplicaSet → Pod(s)
        │               ├── env ← ConfigMap + Secret
        │               └── volume ← PVC
        │
        └── [Network]
                ├── Service (ClusterIP) → Pod(s)
                └── Route (TLS Edge) → Service
```

---

## 13. Quick Reference Table

| Command | Description |
|-------|----------|
| `oc new-project <name>` | Create new project/namespace |
| `oc project <name>` | Switch active project |
| `oc new-build --binary --name=X` | Create binary build config |
| `oc start-build X --from-dir=.` | Start build from local directory |
| `oc get all -n <ns>` | List all resources in namespace |
| `oc apply -f file.yaml` | Create/update resource from YAML |
| `oc get pods -w` | Watch pods live |
| `oc logs -f deploy/X` | Stream deployment logs |
| `oc exec -it deploy/X -- sh` | Open shell inside pod |
| `oc rollout restart deploy/X` | Restart pods |
| `oc rollout undo deploy/X` | Rollback to previous version |
| `oc scale deploy/X --replicas=N` | Change replica count |
| `oc get route` | List routes |
| `oc expose svc X` | Create Route from Service |
| `oc describe <resource> <name>` | View resource details |
| `oc delete project <name>` | Delete project entirely |

---

## 14. Deployment via Web Console (Portal)

While CLI provides more control, the OpenShift Web Console offers a visual and automated way to deploy applications directly from Git.

### Steps for Deployment

1.  **Switch to Developer Perspective:** Ensure you are in the "Developer" view in the top-left corner of the console.
2.  **Create/Select Project:** Choose your `demo` project.
3.  **Click "+Add":** Select the **"Import from Git"** option.
4.  **Configure Git:**
    *   **Git Repo URL:** `https://github.com/mehmetdenizli/openshift-demo.git`
    *   **Strategy:** OpenShift will automatically detect the `Dockerfile`.
5.  **Create:** Click the **"Create"** button at the bottom.

### What Happens Automatically?

When using "Import from Git", OpenShift creates several resources for you:
*   **BuildConfig & ImageStream:** To handle the build from GitHub.
*   **Deployment:** To run the application.
*   **Service:** To expose the application internally.
*   **Route:** To provide an external URL.

### Advantages of Portal Deployment

*   **Automation (Webhooks):** You can set up a GitHub Webhook so that every `git push` automatically triggers a new build and deployment.
*   **Topology View:** Provides a clear visual representation of your application's status (as seen in the screenshot).
*   **Easy Scaling:** You can scale pod replicas by simply clicking the up/down arrows on the deployment circle.

---

> **💡 Tip:** `oc explain <resource>` command documents fields of any resource.  
> Example: `oc explain deployment.spec.strategy`

# OpenShift SCC (Security Context Constraints) Demo

This demo illustrates how Security Context Constraints (SCC) in OpenShift affect container execution and demonstrates when and how to use the `anyuid` privilege.

## Scenario
The official Nginx image (`nginx:latest`) attempts to run as the **root** user (UID 0) and access system directories like `/var/cache/nginx`. By default, OpenShift's `restricted-v2` policy prevents this.

---

### Step 1: Create a New Project
```bash
oc new-project scc-demo
```

### Step 2: Run the Pod & Observe Failure
```bash
oc run nginx-test --image=nginx:latest
```

**Analyze the Status:**
`oc get pod nginx-test` will show `CrashLoopBackOff` or `Error`. 
`oc describe pod nginx-test` will show the pod is assigned to `restricted-v2` SCC, but might not show the exact permission error.

**Check the Logs (The "Smoking Gun"):**
```bash
oc logs nginx-test
```
*Crucial Output:* `nginx: [emerg] mkdir() "/var/cache/nginx/client_temp" failed (13: Permission denied)`
> **Note:** This proves the container is trying to perform root-level operations that `restricted-v2` blocks.

### Step 3: Verify Assigned SCC
```bash
oc get pod nginx-test -o jsonpath='{.metadata.annotations.openshift\.io/scc}'
```
*Output:* `restricted-v2`

### Step 4: Solution - Grant `anyuid` Privilege
> [!IMPORTANT]
> **Prerequisite:** This command requires **cluster-admin** privileges. If you are logged in as `developer`, you will get a `Forbidden` error. Switch to a privileged user first (e.g., `oc login -u kubeadmin`).

**The Command:**
```bash
oc adm policy add-scc-to-user anyuid -z default
```

**What does this command do?**
- `add-scc-to-user anyuid`: It grants the permission to use the `anyuid` SCC.
- `-z default`: It targets the `default` **ServiceAccount** in the **current namespace**.
- **Scope:** This is project-specific. It allows any pod running with the `default` service account *in this project* to bypass the UID restrictions and run as root (or any other UID).

### Step 5: Redeploy and Verify Success
```bash
oc delete pod nginx-test
oc run nginx-test --image=nginx:latest
```

**Final Verification:**
```bash
oc get pod nginx-test
oc exec nginx-test -- id
```
*Output:* `uid=0(root) gid=0(root)...` ✅

**Check the SCC Annotation again:**
```bash
oc get pod nginx-test -o jsonpath='{.metadata.annotations.openshift\.io/scc}'
```
*Output:* `anyuid` ✅

---

## Proof for Presentation (Pro Tip)
If you were using a default SCC, you would see a high UID like `1000670000`. By using `anyuid`, the image's requested `root` (0) user is honored.

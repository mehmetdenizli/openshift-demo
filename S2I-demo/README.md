# OpenShift S2I (Source-to-Image) & Build Automation Demo

This demo illustrates the core OpenShift developer experience: converting source code into a running container image without writing a Dockerfile, using **Source-to-Image (S2I)**, **BuildConfig**, and **ImageStream**.

We will be using our own Python application located in the `S2I-demo/python-app/` directory of this repository.

## Core Concepts

- **BuildConfig (BC):** The "recipe" for how to build your application. It defines the source code location and the builder image to use.
- **ImageStream (IS):** An abstraction layer for container images. It acts as a tag that tracks image versions and can trigger automatic deployments when a new version is built.
- **S2I (Source-to-Image):** A tool/process that takes source code, injects it into a builder image, and produces a new, ready-to-run container image.

---

### Step 1: Project Setup

Create a clean namespace for the demo.

```bash
oc new-project s2i-demo
```

### Step 2: Create the Application from our Repo

We use the `oc new-app` command, specifying the Python builder image, our repository URL, and the directory where our code lives (`--context-dir`).

```bash
oc new-app python:3.9-ubi8~https://github.com/mehmetdenizli/openshift-demo.git --context-dir=S2I-demo/python-app --name=python-s2i
```

**What happened here?**

- `python:3.9-ubi8`: The **Builder ImageStream** (Input).
- `https://...`: **Our Source Code** (Input).
- `--context-dir=S2I-demo/python-app`: Tells OpenShift that the code is in a specific folder.
- OpenShift automatically created a **BuildConfig** and an **ImageStream** for the output image.

> [!NOTE]
> **Understanding `python:3.9-ubi8`**
> This is a **Builder Image** that contains more than just Python:
>
> - **python:3.9**: The specific language runtime.
> - **ubi8 (Universal Base Image)**: A secure, RHEL-based Linux foundation from Red Hat.
> - **S2I Scripts**: The image contains built-in intelligence to detect your `requirements.txt`, run `pip install`, and configure the app to start automatically.
> *In short: You don't need a Dockerfile because this "kitchen" already knows how to cook your Python code.*

### Step 3: Inspecting Created Objects (The Behind-the-Scenes)

OpenShift doesn't just build your code; it creates a full set of Kubernetes/OpenShift objects. You can inspect all of them to see the "magic" YAML files.

**Check Deployment (The Pod manager):**

```bash
oc get deployment python-s2i -o yaml
```

> **Focus points:** Notice the `triggers` and `image` fields. It's waiting for the ImageStream to provide the container.

**Check Service (The Networking):**

```bash
oc get svc python-s2i
```

**Check BuildConfig (The Build recipe):**

```bash
oc get bc python-s2i -o yaml
```

> **Focus points:** Look for `strategy: type: Source` and `output: to: kind: ImageStreamTag`. This tells OpenShift: "Take the source from `S2I-demo/python-app/`, build it, and push the result to the ImageStream."

**Check ImageStream (The Version tracker):**

```bash
oc get is python-s2i
```

### Step 4: Follow the Build Process

Watch your code being converted into a container:

```bash
oc logs -f bc/python-s2i
```

> **What to see:** You will see OpenShift downloading your code, running `pip install flask`, and pushing the final image to the internal registry.

### Step 5: Expose the Application

Once the build is complete, expose the service to get a public URL:

```bash
oc expose svc/python-s2i
oc get route python-s2i
```

> **Action:** Open the URL in your browser. You should see: **"Hello from Python S2I!"** along with the hostname and version.

### Step 6: Automation & Triggering (The Finale)

Demonstrate how OpenShift reacts to new builds.

Trigger a new build manually:

```bash
oc start-build python-s2i --follow
```

**The Chain Reaction:**

1. **Build** finishes -> Updates **ImageStream**.
2. **ImageStream** update -> Triggers **Deployment**.
3. **Deployment** starts -> **Rolling Update** replaces the old pod with the new one.

### How to Verify the Trigger? (Proof of Automation)
Use these 4 commands to prove the automation worked:

**1. Check Build History:**
```bash
oc get builds
```
> See `python-s2i-2` in `Complete` status.

**2. Check Rollout History:**
```bash
oc rollout history deployment/python-s2i
```
> You should see `REVISION 2`. This proves the Deployment was automatically updated.

**3. Watch Pods:**
```bash
oc get pods -w
```
> Observe the old pod terminating and the new one starting.

**4. Check Events:**
```bash
oc describe deployment python-s2i
```
> Look at the events for "Scaled up replica set" messages.

### Step 7: Adding Automated Tests (Quality Gate)
One of the most powerful features of OpenShift builds is the **postCommit** hook. It allows you to run tests *inside* the newly built image before it is pushed to the registry. If tests fail, the build fails.

**1. Apply the Test Hook to BuildConfig:**
```bash
oc patch bc/python-s2i --patch '{"spec":{"postCommit":{"command":["pytest"]}}}'
```

**2. Trigger a Build and Watch the Tests:**
```bash
oc start-build python-s2i --follow
```
> **What to see:** In the logs, you will now see `pytest` running after the dependencies are installed. If the tests pass, the image is pushed.

**3. Test the Quality Gate (Bonus Demo):**
If you break the code (e.g., change the expected text in `app.py` without updating `test_app.py`) and push, the build will fail during the `pytest` stage, and the old version of the app will remain running safely.

---

## Technical Summary for Presentation

- **BuildConfig:** The "How-to" guide. It fetches code from our repo and builds the image.
- **ImageStream:** The "Label/Tag". It stores the build result and triggers deployments.
- **Deployment:** Watches the ImageStream. When a new image is pushed, it updates the app automatically.

This workflow showcases how OpenShift handles the entire lifecycle from Git to a running URL without needing any Docker knowledge from the developer.

# OpenShift S2I (Source-to-Image) & Build Automation Demo

This demo illustrates the core OpenShift developer experience: converting source code into a running container image without writing a Dockerfile, using **Source-to-Image (S2I)**, **BuildConfig**, and **ImageStream**.

We will be using our own Python application located in the `python-app/` directory of this repository.

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
oc new-app python:3.9-ubi8~https://github.com/mehmetdenizli/openshift-demo.git --context-dir=python-app --name=python-s2i
```

**What happened here?**
- `python:3.9-ubi8`: The **Builder ImageStream** (Input).
- `https://...`: **Our Source Code** (Input).
- `--context-dir=python-app`: Tells OpenShift that the code is in a specific folder.
- OpenShift automatically created a **BuildConfig** and an **ImageStream** for the output image.

### Step 3: Inspect BuildConfig and ImageStream
Verify the automation objects created by OpenShift.

**Check BuildConfig:**
```bash
oc get bc python-s2i -o yaml
```
> **Focus points:** Look for `strategy: type: Source` and `output: to: kind: ImageStreamTag`. This tells OpenShift: "Take the source from `python-app/`, build it, and push the result to the ImageStream."

**Check ImageStream:**
```bash
oc get is python-s2i
```
> **Focus points:** Notice the internal image name. This IS will track every new version of your build.

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

---

## Technical Summary for Presentation
- **BuildConfig:** The "How-to" guide. It fetches code from our repo and builds the image.
- **ImageStream:** The "Label/Tag". It stores the build result and triggers deployments.
- **Deployment:** Watches the ImageStream. When a new image is pushed, it updates the app automatically.

This workflow showcases how OpenShift handles the entire lifecycle from Git to a running URL without needing any Docker knowledge from the developer.

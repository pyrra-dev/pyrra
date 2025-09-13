# Pyrra Dev Environment Setup Summary

This document outlines the steps taken to configure a local development environment for Pyrra on Windows 10 using Minikube, Hyper-V, and VS Code with a Git Bash terminal.

## 1. Initial Goal

The primary objective is to set up a local Kubernetes cluster to build, deploy, and test modifications to the Pyrra project, specifically focusing on its dynamic threshold alerting logic.

## 2. Minikube Installation and Troubleshooting

We started by setting up a Minikube cluster, which required resolving several platform-specific issues on Windows.

### Issue 1: Hypervisor Conflict (VirtualBox vs. Hyper-V)

*   **Problem**: The default `minikube start` command failed because it tried to use the VirtualBox driver while the native Windows Hyper-V hypervisor was active. Only one hypervisor can be active at a time.
*   **Solution**: Explicitly instructed Minikube to use the active hypervisor by adding the `--driver=hyperv` flag.
    ```sh
    minikube start --driver=hyperv --cpus=4 --memory=8g
    ```

### Issue 2: Hyper-V Administrator Permissions

*   **Problem**: Even when running from an Administrator PowerShell, the command failed with a `PROVIDER_HYPERV_NOT_RUNNING` error. This was because the user account was not a member of the required **"Hyper-V Administrators"** security group.
*   **Solution**: Added the user account to the "Hyper-V Administrators" group and rebooted the system for the new permissions to take effect.

## 3. Configuring the Docker Environment for `ko`

The next challenge was connecting the local shell environment to the Docker daemon running *inside* the Minikube virtual machine. This is required for the `ko` tool to build and push images.

### Issue 3: Shell Permissions and VS Code Integration

*   **Problem**: Commands like `minikube status` or `minikube docker-env` failed in a standard Git Bash terminal with an `Access is denied` error when trying to read the SSH key at `~/.minikube/machines/minikube/id_rsa`. This key was created by an administrator process and was inaccessible to the standard user's shell.
*   **Attempted Solutions**:
    1.  **Running VS Code as Administrator**: This failed with a `launch-failed` error, a known issue with VS Code on some Windows configurations.
    2.  **Custom Admin Terminal Profile**: This approach launched a new, external PowerShell window instead of elevating the integrated terminal, making it unsuitable for our workflow.

### Final Solution: The Environment Script Workflow

We settled on a robust, secure workflow that does not require running VS Code with elevated privileges.

1.  **One-Time Setup**: Use a standalone **Administrator PowerShell** to generate a `bash`-compatible environment script. This captures the necessary Docker environment variables.
    ```powershell
    # Run this once in an Admin PowerShell
    minikube -p minikube docker-env --shell bash | Out-File -FilePath "$HOME\minikube-env.sh"
    ```

2.  **Daily Workflow**: In any **standard Git Bash terminal** inside VS Code, run the following commands to configure the session:
    ```bash
    # Source the script to load Docker environment variables
    source ~/minikube-env.sh

    # Set the repository variable required by ko for local development
    export KO_DOCKER_REPO='ko.local'
    ```

This final approach provides a stable and repeatable method for preparing the development environment.

## 4. Automating the Environment with `direnv`

To avoid manually sourcing the environment script in every new terminal, we automated the process on a per-project basis using `direnv`.

1.  **Project-Scoped Script**: The `minikube-env.sh` script was moved from the home directory to a project-specific `.dev/` folder for better encapsulation.

2.  **`direnv` Configuration**: A `.envrc` file was created in the project root to automatically source the script whenever we `cd` into the directory.
    ```bash
    # c:/Users/Yair/Code/pyrra/.envrc
    source .dev/minikube-env.sh
    ```

3.  **Git Ignore**: The `.dev/` directory and the `.envrc` file were added to `.gitignore` to keep local environment configuration out of version control.

4.  **Verification**: The setup is verified by running `docker ps` inside the project directory and confirming it lists the Kubernetes containers from Minikube, not from Docker Desktop.

## 5. Installing kube-prometheus Backend

With the environment configured, we need to install the complete monitoring stack using `kube-prometheus` (jsonnet-based) to serve as the monitoring backend for Pyrra. This provides better integration with Pyrra than the Helm-based `kube-prometheus-stack`.

**Note**: After initial development with `kube-prometheus-stack` (Helm), we migrated to `kube-prometheus` (jsonnet) based on Pyrra's official recommendations for better compatibility and integration.

1.  **Clone kube-prometheus**: Get the official jsonnet-based monitoring stack.
    ```bash
    git clone https://github.com/prometheus-operator/kube-prometheus.git
    cd kube-prometheus
    ```

2.  **Deploy the monitoring stack**: Install CRDs first, then the complete stack.
    ```bash
    # Deploy the CRDs and the Prometheus Operator
    kubectl apply -f ./manifests/setup
    
    # Deploy all resources like Prometheus, StatefulSets, and Deployments
    kubectl apply -f ./manifests/
    ```

3.  **Verify deployment**: Check that all pods are running in the monitoring namespace.
    ```bash
    kubectl get pods -n monitoring
    ```

This complete stack provides:
- **Prometheus**: Metrics collection and querying with proper label selectors
- **Grafana**: Visualization and dashboards  
- **AlertManager**: Alert routing and notification
- **Various Exporters**: Node exporter, kube-state-metrics, etc.
- **Proper Integration**: Native support for Pyrra's PrometheusRule generation

**Advantages over kube-prometheus-stack**:
- ✅ Native Pyrra integration with matching label selectors
- ✅ Upstream-recommended setup for Pyrra deployments
- ✅ Better PrometheusRule discovery and processing
- ✅ Consistent with official Pyrra documentation and examples

For accessing services locally, use port-forwarding:
```bash
# Prometheus (note the different service name)
kubectl port-forward svc/prometheus-k8s 9090:9090 -n monitoring

# Grafana (admin/admin)  
kubectl port-forward svc/grafana 3000:3000 -n monitoring

# AlertManager
kubectl port-forward svc/alertmanager-main 9093:9093 -n monitoring
```

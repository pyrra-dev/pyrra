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

## 5. Installing Prometheus Backend

With the environment configured, we installed Prometheus using Helm to serve as the monitoring backend for Pyrra.

1.  **Add Helm Repo**: Added the official Prometheus community chart repository.
    ```bash
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    ```

2.  **Install Chart**: Installed the `prometheus` chart into the `default` namespace for simplicity in our local development context.
    ```bash
    helm install prometheus prometheus-community/prometheus
    ```

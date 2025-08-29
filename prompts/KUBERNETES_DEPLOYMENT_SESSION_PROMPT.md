# Kubernetes Deployment Session Prompt

## Session Context
This session continues the dynamic burn rate feature implementation for Pyrra, specifically focusing on **Priority 2: Testing and Deployment**. The UI components have been updated to display dynamic vs static burn rate information, and a custom Docker image has been built with these changes.

## Current Status

### ✅ Completed Work
1. **UI Component Updates** - All dynamic burn rate display components have been implemented:
   - `burnrate.tsx`: Helper functions for burn rate type detection and display logic
   - `AlertsTable.tsx`: Dynamic-aware tooltips and threshold display with icons
   - `BurnrateGraph.tsx`: Context-aware threshold descriptions
   - `Detail.tsx`: Import organization fixes

2. **Custom Docker Image** - Built successfully:
   - Image name: `pyrra-with-burnrate:latest`
   - Built with Dockerfile.custom using Go 1.24.0-alpine
   - Includes embedded UI build with all burn rate changes
   - Fixed user permissions (using `nobody` user)

3. **Kubernetes Manifests** - Updated for custom image:
   - `examples/kubernetes/manifests/pyrra-kubernetesDeployment.yaml`
   - Changed image to `pyrra-with-burnrate:latest`
   - Added `imagePullPolicy: Never` for local testing

### 🎯 Immediate Objectives
The main goal is to **test the dynamic burn rate UI changes** by:

1. **Set up Kubernetes cluster** with kube-prometheus-stack (not just Prometheus)
2. **Load and deploy custom image** to test dynamic vs static burn rate display
3. **Create SLO configurations** to demonstrate both static and dynamic burn rates
4. **Verify UI functionality** shows correct badges, tooltips, and behavior

### 🔧 Required Infrastructure Setup

#### Minikube Cluster Setup
Based on previous session (see SETUP_SUMMARY.md), you'll need:
```bash
# Start minikube with proper resources
minikube start --driver=hyperv --cpus=4 --memory=8g

# Configure Docker environment for local images
minikube docker-env --shell bash > ~/.minikube-env.sh
source ~/.minikube-env.sh
export KO_DOCKER_REPO='ko.local'
```

#### Install kube-prometheus-stack (NOT just prometheus)
The SETUP_SUMMARY.md is **outdated** - it mentions installing only Prometheus, but we need the full kube-prometheus-stack for proper SLO monitoring:

```bash
# Add Prometheus community repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install complete monitoring stack
helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace \
  --set prometheus.service.type=NodePort \
  --set grafana.service.type=NodePort \
  --set alertmanager.service.type=NodePort
```

#### Load Custom Image
```bash
# Load our custom Docker image into minikube
minikube image load pyrra-with-burnrate:latest

# Verify image is available
minikube image ls | grep pyrra-with-burnrate
```

### 📋 Deployment Steps

1. **Apply Pyrra manifests** with custom image:
   ```bash
   kubectl apply -f examples/kubernetes/manifests/
   ```

2. **Create test SLO configurations** - both static and dynamic types:
   - Static SLO: Traditional fixed burn rate thresholds
   - Dynamic SLO: Traffic-aware adaptive thresholds
   - Use examples from `examples/` folder as templates

3. **Port forward services** for testing:
   ```bash
   # Pyrra API/UI
   kubectl port-forward svc/pyrra-kubernetes 9099:9099 -n monitoring
   
   # Prometheus (for data)
   kubectl port-forward svc/kube-prometheus-stack-prometheus 9090:9090 -n monitoring
   
   # Grafana (for visualization)
   kubectl port-forward svc/kube-prometheus-stack-grafana 3000:80 -n monitoring
   ```

4. **Access and test UI**:
   - Pyrra: http://localhost:9099
   - Prometheus: http://localhost:9090  
   - Grafana: http://localhost:3000 (admin/prom-operator)

### 🎯 Testing Checklist

#### UI Component Verification
- [ ] **Badge Display**: Dynamic SLOs show green "Dynamic" badges, static SLOs show gray "Static" badges
- [ ] **Tooltip Content**: Hover over thresholds shows appropriate dynamic vs static explanations
- [ ] **Threshold Display**: Dynamic thresholds show "Traffic-Aware" instead of multiplier factors
- [ ] **Graph Descriptions**: BurnrateGraph shows context-aware threshold descriptions
- [ ] **Icon Integration**: Dynamic burn rate entries show the dynamic icon

#### API Integration Testing  
- [ ] **Backend Connection**: UI successfully connects to Pyrra API on port 9099
- [ ] **Data Retrieval**: SLO objectives load with correct burn rate type information
- [ ] **Dynamic Detection**: Backend properly identifies and serves dynamic vs static burn rate types

#### End-to-End Functionality
- [ ] **SLO Creation**: Can create both static and dynamic SLOs via Kubernetes manifests
- [ ] **Real-time Updates**: UI updates reflect current burn rate calculations
- [ ] **Alert Behavior**: Dynamic thresholds adapt based on actual traffic patterns

### 🐛 Known Issues & Solutions

#### API Connectivity Issues
Previous session encountered API connectivity problems when running standalone. The backend requires proper Kubernetes deployment to expose APIs correctly.

**Solution**: Deploy to Kubernetes cluster instead of standalone operation.

#### Docker Image Management
The custom image `pyrra-with-burnrate:latest` contains all our UI changes and must be loaded into minikube for testing.

**Critical**: Always use `minikube image load` rather than trying to pull from registry.

#### Development Workflow
For UI development, you can still run `npm start` in the `ui/` directory for rapid iteration, but final testing must be done against the Kubernetes-deployed backend.

### 🎨 UI Development Context

#### File Overview
- **burnrate.tsx**: Core logic for burn rate type detection and helper functions
- **AlertsTable.tsx**: Main table showing SLO alerts with dynamic-aware tooltips
- **BurnrateGraph.tsx**: Graph component with context-sensitive descriptions  
- **Detail.tsx**: SLO detail page with proper imports

#### Key Functions Implemented
- `getBurnRateType(objective)`: Detects dynamic vs static from backend API
- `getBurnRateTooltip(objective, factor?)`: Context-aware tooltip content
- `getBurnRateDisplayText(objective, factor?)`: Display text for thresholds
- `getThresholdDescription(objective, threshold, shortWindow, longWindow)`: Graph descriptions

### 🚀 Success Criteria
Session is complete when:
1. ✅ Kubernetes cluster running with kube-prometheus-stack
2. ✅ Custom Pyrra image deployed and accessible  
3. ✅ Both static and dynamic SLOs configured and visible
4. ✅ UI correctly displays different badge colors and tooltips for each type
5. ✅ Dynamic burn rate thresholds show "Traffic-Aware" instead of static multipliers
6. ✅ All components integrate seamlessly without compilation errors

### 📚 Reference Files
- `SETUP_SUMMARY.md`: Initial setup process (needs kube-prometheus-stack update)
- `examples/kubernetes/manifests/`: Kubernetes deployment manifests
- `examples/`: Sample SLO configurations to adapt for testing
- `Dockerfile.custom`: Multi-stage build with UI integration

---

**Next Actions**: Start with minikube setup, install kube-prometheus-stack, load custom image, and deploy manifests for testing.

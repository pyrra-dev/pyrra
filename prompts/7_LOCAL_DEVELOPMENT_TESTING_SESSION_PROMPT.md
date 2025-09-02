# Local Development Testing Session Prompt

## Session Context  
This session continues dynamic burn rate feature testing using the **maintainer-recommended local development workflow** instead of Docker deployments. Previous session identified that CRD schema mismatch was blocking dynamic SLO creation, and the proper solution is using local binaries with current Go code.

## Current Status

### ✅ Completed in Previous Session
- **Root Cause Analysis**: CRD missing `burnRateType` field prevents dynamic SLO creation in Kubernetes
- **Development Approach**: Switched from Docker deployment to local development per CONTRIBUTING.md
- **UI Improvements**: Enhanced burnrate.tsx with better tooltip calculations and decimal formatting
- **Build Success**: `make all` completed successfully with latest code
- **Clean Commit**: Production changes committed (f299ce5), temporary test configs avoided

### 🎯 Immediate Objectives  
Follow CONTRIBUTING.md local development workflow to test dynamic burn rate UI:

1. **Run local backends**: `./pyrra kubernetes` (connects to k8s cluster)
2. **Run API server**: `./pyrra api` (serves on localhost:9099) 
3. **Test with existing SLOs**: Local backend handles missing CRD fields gracefully
4. **Verify UI functionality**: Dynamic badges, tooltips, threshold displays
5. **Document findings**: Results of local development testing approach

## 🔧 Required Setup

### Infrastructure Ready
- ✅ **Minikube cluster**: Running with kube-prometheus-stack
- ✅ **Local binaries**: Built via `make all` with latest Go code including burnRateType support
- ✅ **Existing SLOs**: Multiple SLOs available in monitoring namespace for testing

### Development Workflow (per CONTRIBUTING.md)
```bash
# 1. Start Kubernetes backend (connects to cluster)
./pyrra kubernetes

# 2. Start API server (in separate terminal)
./pyrra api

# 3. UI testing options:
# Option A: Use embedded UI at http://localhost:9099
# Option B: Development server at http://localhost:3000 (cd ui && npm start)
```

## 📋 Testing Plan

### Phase 1: Local Backend Testing
1. **Start `./pyrra kubernetes`**: Verify cluster connection and SLO discovery
2. **Check logs**: Ensure no errors reading existing SLOs despite CRD field mismatch
3. **Validate behavior**: Local backend should handle missing burnRateType gracefully

### Phase 2: API Integration Testing  
1. **Start `./pyrra api`**: Launch API server on localhost:9099
2. **Test API endpoints**: Verify SLO data serves correctly with burn rate type info
3. **Check API responses**: Confirm dynamic vs static burn rate detection logic works

### Phase 3: UI Functionality Verification
1. **Access UI**: Navigate to http://localhost:9099 
2. **SLO List Display**: Verify existing SLOs show with proper badges/tooltips
3. **Dynamic Detection**: Check if any SLOs are detected as dynamic type
4. **Tooltip Quality**: Test improved decimal formatting from burnrate.tsx changes
5. **Badge Behavior**: Verify static SLOs show gray "Static" badges

### Phase 4: Dynamic SLO Creation Test
1. **Create test SLO**: Apply SLO with burnRateType field via local API
2. **Verify processing**: Check if local backend accepts and processes dynamic config
3. **UI reflection**: Confirm dynamic SLO shows green "Dynamic" badge
4. **Tooltip content**: Validate dynamic-specific tooltip descriptions

## 🚀 Success Criteria

### ✅ **Minimum Success** (Local Development Validation)
- Local backends run without errors
- Existing SLOs display correctly with improved tooltips
- API serves burnRateType information properly
- UI shows appropriate badges based on burn rate detection

### 🎯 **Full Success** (Dynamic Feature Demonstration)  
- Can create and display dynamic SLOs via local backend
- Green "Dynamic" badges appear for dynamic SLOs
- Tooltips show "Traffic-Aware" instead of static multipliers  
- All UI improvements from burnrate.tsx changes visible and working

### 📊 **Bonus Success** (Production Readiness Assessment)
- Document what works with local development approach
- Identify what still needs CRD regeneration for production deployment
- Create clear guidance for production vs development testing workflows

## 🐛 Expected Challenges & Solutions

### Challenge 1: CRD Field Mismatch
- **Issue**: Existing SLOs don't have burnRateType field
- **Expected**: Local backend should default to "static" type gracefully
- **Solution**: Go code logic should handle missing fields with defaults

### Challenge 2: API Connectivity
- **Issue**: Previous sessions had API integration problems  
- **Expected**: Local development approach should resolve this
- **Solution**: Follow exact CONTRIBUTING.md workflow sequence

### Challenge 3: Dynamic SLO Testing
- **Issue**: May not be able to create dynamic SLOs due to CRD validation
- **Expected**: Local backend might bypass Kubernetes validation
- **Solution**: Test both approaches - Kubernetes apply vs API creation

## 📚 Reference Context

### Key Files Modified
- **burnrate.tsx**: Enhanced tooltip calculations, decimal formatting
- **FEATURE_IMPLEMENTATION_SUMMARY.md**: Added PR preparation guidelines

### Previous Session Insights
- Docker deployment approach was unnecessarily complex
- CRD regeneration needed for production but not for local testing
- Maintainer workflow designed for this exact scenario

### Code Status
- Current branch: `add-dynamic-burn-rate` 
- Latest commit: f299ce5 (production improvements)
- Go code includes full burnRateType support
- UI code ready for testing

## 🎯 Next Actions Sequence

1. **Terminal 1**: `./pyrra kubernetes` → Monitor logs for cluster connection
2. **Terminal 2**: `./pyrra api` → Start API server  
3. **Browser**: Navigate to http://localhost:9099 → Test basic UI functionality
4. **Testing**: Create dynamic SLO → Verify green badge appears
5. **Documentation**: Record results and production deployment requirements

---

**Expected Outcome**: Successful demonstration of dynamic burn rate UI functionality using local development workflow, with clear documentation of what works locally vs what needs CRD regeneration for production deployment.

# Dynamic Burn Rate Validation Session Prompt

## Session Context
Previous sessions successfully completed **basic UI functionality testing** and resolved **critical production deployment issues** for the dynamic burn rate feature. Static threshold display now works correctly in embedded UI (production). However, **critical data validation and real-world functionality issues** remain that need comprehensive investigation and resolution.

## ✅ Previous Session Achievements
- **UI Components Working**: Green/gray badges, tooltips, sorting, navigation all functional
- **API Integration**: burnRateType field transmission confirmed
- **Mixed Environment**: 7 SLOs (2 dynamic + 5 static) environment established
- **Windows Compatibility**: API architecture documented, CRD generation workarounds identified
- **🎉 PRODUCTION ISSUE RESOLVED**: Static SLO threshold display fixed in embedded UI (port 9099)
- **Critical Documentation Added**: Complete UI build workflow documented for future developers

## ✅ Major Production Issue Resolved (September 4, 2025)
**Issue**: Static SLOs showed "14x, 7x, 2x, 1x" in embedded UI while development UI showed correct calculated thresholds
**Solution**: Implemented complete UI build workflow (`npm run build` → `make build` → restart)
**Result**: Embedded UI now displays correct calculated thresholds (e.g., "0.700, 0.350, 0.100, 0.050" for 95% SLO)
**Documentation**: Added workflow guidance to CONTRIBUTING.md, ui/README.md, and FEATURE_IMPLEMENTATION_SUMMARY.md

## 🚨 Critical Issues Identified Requiring Investigation

### **Issue 1: Missing Availability/Budget Data** 
**Problem**: All SLOs show "No data" in Availability and Budget columns despite API integration working
- **Root Cause**: Likely no actual metric data available for prometheus_http_requests_total in test environment
- **Impact**: Cannot validate real-world SLO calculations, error budget consumption, or threshold behavior
- **Investigation Needed**: 
  - Check if prometheus_http_requests_total has actual data in our cluster
  - Investigate why GetStatus API endpoint returns empty responses
  - Consider using metrics with real data (API server metrics, etc.)
  - Potentially create synthetic load to generate test data

### **Issue 2: Data Correctness Validation Missing**
**Problem**: No validation of actual burn rate calculations, thresholds, or dynamic vs static behavior
- **Root Cause**: Testing focused on UI display, not mathematical correctness of underlying calculations
- **Impact**: Cannot confirm dynamic thresholds are actually calculated correctly
- **Investigation Needed**:
  - Validate dynamic threshold calculations match expected formula
  - Compare static vs dynamic burn rate values in real scenarios  
  - Verify short/long window burn rates are calculated properly
  - Test threshold adaptation based on traffic volume changes

### **Issue 3: Real-Time Dynamic Threshold Display Enhancement**
**Problem**: Dynamic SLOs show generic "Traffic-Aware" text instead of real-time calculated threshold values
- **Root Cause**: Implementation uses placeholder text rather than actual calculated values
- **Impact**: Users cannot see actual threshold values, reducing observability and debugging capability
- **Enhancement Needed**:
  - Display actual calculated dynamic threshold: `(N_SLO/N_long) × E_budget_percent × (1-SLO_target)`
  - Show current burn rate vs threshold in tooltip: `current_burn_rate × (1-slo_target)`
  - Provide real-time threshold updates based on traffic patterns

### **Issue 4: Prometheus Rules Generation Investigation**
**Problem**: Prometheus UI may not be showing generated rules correctly when both test and example SLOs are applied
- **Root Cause**: Potentially resolved in previous sessions but not documented in .dev-docs
- **Impact**: Cannot validate that dynamic burn rate expressions are actually being used by Prometheus
- **Investigation Needed**:
  - Verify Prometheus rules are generated with correct dynamic expressions
  - Check if rules are loading properly in Prometheus UI
  - Validate that dynamic threshold calculations are present in generated PromQL
  - Document the resolution if previously fixed

## 🎯 Session Objectives

### **Phase 1: Data Infrastructure Validation**
1. **Metric Data Investigation**:
   - Check what metrics actually have data in our test environment
   - Investigate prometheus_http_requests_total data availability
   - Consider switching to metrics with real data (apiserver_request_total, etc.)
   - Potentially generate synthetic load for testing

2. **API Data Flow Analysis**:
   - Debug why GetStatus endpoint returns empty responses
   - Investigate recording rules and metric availability
   - Validate that SLO calculations can access necessary data

### **Phase 2: Dynamic Threshold Calculation Validation**
1. **Mathematical Correctness**:
   - Verify dynamic threshold calculations match expected formula
   - Compare generated PromQL expressions with theoretical expectations
   - Test threshold values with different traffic patterns

2. **Real-Time Threshold Display**:
   - Replace "Traffic-Aware" placeholder with actual calculated values
   - Implement real-time threshold calculation display
   - Add detailed tooltip showing calculation breakdown

### **Phase 3: Prometheus Rules Deep Dive**
1. **Generated Rules Analysis**:
   - Examine actual PromQL expressions generated for dynamic SLOs
   - Verify that dynamic threshold calculations are present
   - Compare dynamic vs static rule generation

2. **Prometheus Integration Validation**:
   - Confirm rules load properly in Prometheus UI
   - Test rule evaluation and alerting behavior
   - Validate recording rules generate expected metrics

### **Phase 4: End-to-End Functional Testing**
1. **Real-World Scenario Testing**:
   - Test with metrics that have actual data
   - Validate SLO calculations produce meaningful results
   - Compare static vs dynamic behavior under different conditions

2. **Load Testing for Dynamic Behavior**:
   - Generate traffic to test threshold adaptation
   - Validate dynamic thresholds change based on traffic patterns
   - Confirm alerts fire correctly with dynamic vs static thresholds

## 🧪 Testing Methodology

### **Data-Driven Approach**
- Use metrics with confirmed data availability
- Generate synthetic load if necessary for testing
- Focus on mathematical correctness over UI aesthetics

### **Comparative Analysis**
- Side-by-side testing of static vs dynamic SLOs
- Threshold behavior validation under different traffic conditions
- Performance and accuracy comparison

### **Real-World Validation**
- Test with production-like metric patterns
- Validate alerting behavior under realistic conditions
- Confirm feature provides actual value over static approach

## 📋 Success Criteria

### **Minimum Success** (Data Infrastructure Working)
- SLO availability and budget data displays correctly
- GetStatus API returns meaningful data
- Basic threshold calculations work

### **Full Success** (Complete Dynamic Feature Validation)
- Real-time dynamic threshold display working
- Mathematical correctness confirmed through testing
- Dynamic vs static behavior clearly differentiated
- Prometheus rules generation and evaluation working correctly

### **Production Readiness** (Real-World Validation)
- Feature tested with realistic data and load patterns
- Performance and accuracy validated
- Documentation complete for deployment and troubleshooting

## 🔧 Expected Challenges

### **Data Availability**
- Test environment may lack realistic metric data
- May need to switch to different metrics or generate synthetic load
- Recording rules may need time to populate

### **Prometheus Integration Complexity**
- Dynamic PromQL expressions may be complex to validate
- Rule generation and loading may have timing issues
- Alertmanager integration may need specific configuration

### **Mathematical Validation**
- Dynamic threshold calculations involve multiple variables
- Real-time calculations may have performance implications
- Edge cases (zero traffic, traffic spikes) need special handling

## 📚 Reference Materials

### **Previous Session Insights**
- API architecture documented in FEATURE_IMPLEMENTATION_SUMMARY.md
- Windows compatibility issues and workarounds identified
- Basic UI integration confirmed working

### **Implementation Files**
- `slo/rules.go`: Dynamic PromQL generation logic
- `ui/src/burnrate.tsx`: UI threshold display logic
- Generated Prometheus rules in Kubernetes cluster

### **Testing Environment**
- Kubernetes cluster with kube-prometheus-stack
- Mixed SLO environment (2 dynamic + 5 static SLOs currently applied)
- Local development services running
- **Production UI deployment validated** (embedded UI working correctly)

## 🎯 Next Actions Sequence

1. **Data Investigation**: Check metric availability and API data flow
2. **Threshold Validation**: Verify mathematical correctness of dynamic calculations  
3. **Real-Time Display**: Implement actual threshold value display
4. **Prometheus Validation**: Confirm rules generation and evaluation
5. **End-to-End Testing**: Complete functional validation with realistic scenarios
6. **Documentation Update**: Record findings and deployment guidance

---

**Expected Outcome**: Comprehensive validation of dynamic burn rate feature functionality, mathematical correctness, and real-world applicability, with all data display and calculation issues resolved for production deployment.

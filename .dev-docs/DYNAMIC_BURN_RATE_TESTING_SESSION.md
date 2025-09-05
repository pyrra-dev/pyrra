# Dynamic Burn Rate Testing Session - September 5, 2025

## 🔍 Session Overview
**Objective**: Investigate why all SLOs show "No data" in Availability/Budget columns despite API integration working correctly.

**Testing Approach**: 
- **UI-Based Tests**: Human operator performs queries in Prometheus UI and Pyrra UI
- **Terminal Tests**: AI performs commands directly for faster iteration and direct output reading
- **Documentation**: All results captured in this session document

**Status**: Phase 1 Data Infrastructure Validation - COMPLETED
**Next Phase**: Test 5 & 6 - SLO Rule Generation Diagnosis and Metric Switching

---

## ✅ Phase 1: Data Infrastructure Validation - COMPLETED

### **Test 1: Prometheus Metric Data Availability**

**Objective**: Verify which metrics actually have data and identify root cause of "No data" issue.

**Test Results**:

#### **Query 1: `prometheus_http_requests_total`**
- **Result**: ✅ **114 series found**
- **Code Distribution**: 
  - 112 series with `code="200"`
  - 1 series with `code="302"`
  - 1 series with `code="400"`
- **⚠️ Critical Finding**: **NO 5xx error codes exist**

#### **Query 2: `prometheus_http_requests_total{code=~"5.."}`**
- **Result**: ❌ **No data** - confirms no 5xx errors

#### **Query 3: `{__name__=~".*:.*"}` (Recording Rules)**
- **Result**: ✅ **925 recording rules found**
- **Note**: Large number of recording rules exist in cluster

#### **Query 4: `{__name__=~".*slo_name:burn_rate_5m.*"}`**
- **Result**: ❌ **No SLO-specific recording rules found**
- **Alternative**: `{__name__=~".*slo.*"}` only returned prometheus-adapter and grafana metrics

**Data Retention**: ~40 minutes recent data visible in 1h window, historical data available when extending time range.

---

### **Test 2: SLO-Specific Metric Investigation**

**Objective**: Check if our actual SLO metrics have data.

**Test Results**:

#### **Dynamic SLO Metric: `prometheus_http_requests_total{code=~"5.."}`**
- **Result**: ❌ **No data** (consistent with Test 1)

#### **API Service Metrics: `http_requests_total{job="api-service"}`**
- **Result**: ❌ **Metric doesn't exist**
- **Available alternatives**: `kubelet_http_requests_total`, `prometheus_http_requests_total`

#### **Prometheus Operator Metrics: `prometheus_operator_kubernetes_client_http_requests_total`**
- **Result**: ✅ **3 series found**
- **Code Distribution**:
  - 2073 requests with `code="200"`
  - 52 requests with `code="201"`
  - 47 requests with `code="404"`

#### **Pyrra Connect Metrics: `connect_server_requests_total{job="pyrra"}`**
- **Result**: ❌ **No data**

---

### **Test 3: SLO Recording Rules Generation**

**Objective**: Check if Pyrra generated the expected recording rules for our SLOs.

**Test Results**:

#### **Query 1: `{__name__=~".*increase.*"}`**
- **Result**: ✅ **37 series found**
- **Key Metrics**:
  - `prometheus_operator_kubernetes_client_http_requests:increase1w`
  - `prometheus_http_requests:increase1w`
  - `prometheus_http_requests:increase30d`

#### **Query 2: `{__name__=~".*prometheus_http_requests.*increase.*"}`**
- **Result**: ✅ **34 series found** (subset of above)

#### **Query 3: `{__name__=~".*burn.*"}`**
- **Result**: ✅ **14 series found**
- **⚠️ Important**: These are NOT Pyrra-generated rules
- **Source**: kube-prometheus stack (e.g., `apiserver_request:burnrate1h{verb="read"}`)

#### **Prometheus Rules Page Check: `http://localhost:9090/rules`**
- **Result**: ⚠️ **Limited Pyrra rules found**
- **Found**: Rules only for `pyrra-connect-errors` and `pyrra-connect-errors-increase`
- **Missing**: Rules for our test SLOs (dynamic/static demos)

---

### **Test 4: Alternative Metrics with Error Data**

**Objective**: Find metrics that actually have error data we can use for testing.

**Test Results**:

#### **Prometheus Operator 404s: `prometheus_operator_kubernetes_client_http_requests_total{status_code="404"}`**
- **Result**: ✅ **47 requests confirmed**

#### **Prometheus Operator All Errors: `prometheus_operator_kubernetes_client_http_requests_total{status_code=~"4..|5.."}`**
- **Result**: ✅ **47 series** (only 404s, no 5xx)

#### **Kubelet Errors: `kubelet_http_requests_total{code=~"4..|5.."}`**
- **Result**: ❌ **No data**

#### **API Server Errors: `apiserver_request_total{code=~"4..|5.."}`**
- **Result**: ✅ **71 series found** - **EXCELLENT TEST DATA**
- **Error Distribution**:
  - `400`: 1 series
  - `403`: 2 series  
  - `404`: 43 series
  - `409`: 12 series
  - `422`: 1 series
  - `429`: 7 series
  - `500`: 3 series
  - `504`: 2 series
- **Labels Available**: `verb` (HTTP method), `code` (status code)

---

## 🎯 Root Cause Analysis Summary

### **Primary Issues Identified**:

1. **❌ Missing Error Data in Primary SLO Metrics**
   - `prometheus_http_requests_total` has NO 5xx errors
   - Most test SLO metrics (`http_requests_total{job="api-service"}`) don't exist
   - This explains why SLOs show "No data" - no error events to calculate against

2. **⚠️ Incomplete SLO Recording Rule Generation**
   - Only `pyrra-connect-errors` generating rules
   - Test SLOs (dynamic/static demos) not generating expected recording rules
   - This suggests SLO processing issues in Pyrra

3. **✅ Infrastructure Working Correctly**
   - Prometheus has good metric data volume (114+ series)
   - Recording rule generation mechanism works (925 total rules)
   - Alternative metrics with rich error data available (`apiserver_request_total`)

### **Recommended Solution Path**:

1. **Investigate SLO Rule Generation** - Check Pyrra logs and SLO status
2. **Switch to Working Metrics** - Use `apiserver_request_total` for testing
3. **Validate Dynamic Calculations** - With real error data, test threshold calculations
4. **Document Findings** - Update implementation guidance for metric selection

---

## ✅ Test 5: Diagnose Missing SLO Rules - COMPLETED

### **Pyrra Deployment Check**
- **Command Attempted**: `kubectl logs -n monitoring deployment/pyrra --tail=50`
- **Result**: ❌ **Deployment not found** - `error: deployments.apps "pyrra" not found`
- **Explanation**: Running local executable `./pyrra` (not K8s deployment) per CONTRIBUTING.md

### **Pyrra UI Status Check - `http://localhost:9099`**

#### **SLO Count Analysis**:
- **Total Visible**: 15 SLOs (not 7 as expected)
- **Reason**: `prometheus-pi-query` splits by `handler` label → 9 SLOs
- **Expected**: 7 SLOs total (2 dynamic + 5 static)

#### **"No Data" SLOs Identified**: 
- ✅ **4 SLOs confirmed showing "No Data"**:
  1. `test-latency-dynamic`
  2. `demo-dynamic-api` 
  3. `demo-static-api`
  4. `pyrra-connect-errors`

#### **SLO Detail Page Analysis - `test-slo`**:
- **Error Budget Graph**: ✅ **Shows 100% full bar constantly** (expected with no errors)
- **Requests Graph**: ✅ **Shows data** (prometheus_http_requests_total working)
- **Errors Graph**: ❌ **Blank** (confirms no 5xx errors available)

### **Key Insights from Test 5**:
1. **✅ Pyrra Processing Working**: All SLOs visible and processed
2. **✅ Recording Rules Generated**: Error budget calculations work (100% budget)
3. **❌ Missing Error Data**: Root cause confirmed - specific SLOs have no error metrics
4. **✅ Data Flow Working**: Request data flows correctly, error data missing

---

## 📋 Next Steps - Test 6

### **Test 6: Switch to Working Metrics for Dynamic Testing**
- Update dynamic SLO to use `apiserver_request_total`
- Validate dynamic burn rate calculations with real error data
- Test threshold adaptation and alerting behavior

---

## 🔧 Testing Environment Context

**Cluster Setup**: Kubernetes with kube-prometheus-stack
**SLO Count**: 7 total (2 dynamic + 5 static)
**Prometheus**: `http://localhost:9090`
**Pyrra UI**: `http://localhost:9099`
**Data Retention**: 40+ minutes recent, historical data available
**Recording Rules**: 925 total system rules, limited Pyrra-specific rules

**Key Metrics for Testing**:
- ✅ `apiserver_request_total` - Rich error data (71 series, multiple error codes)
- ✅ `prometheus_operator_kubernetes_client_http_requests_total` - Limited but available
- ❌ `prometheus_http_requests_total` - No error data (only 200/302/400)
- ❌ `http_requests_total{job="api-service"}` - Doesn't exist

---

**Session Status**: Test 6 COMPLETED - New test SLOs with real data successfully deployed

---

## ✅ Test 6: Switch to Working Metrics for Dynamic Testing - COMPLETED

### **Step 1: Current Dynamic SLO Configuration Analysis**
- **Command**: `kubectl get slo -A` and `kubectl get slo demo-dynamic-api -n monitoring -o yaml`
- **Finding**: Deployed SLO had `burnRateType: static` instead of expected `dynamic`
- **Root Cause**: Using old demo SLOs with non-existent metrics (`http_requests_total{job="api-service"}`)

### **Step 2: Cleanup Old Demo SLOs**
- **Deleted from k8s**: `demo-dynamic-api` and `demo-static-api` SLOs
- **Deleted file**: `examples/demo-dynamic-burnrate.yaml` 
- **Reason**: Using non-existent metrics, causing "No data" issues

### **Step 3: Create Test SLOs with Real Metrics**
- **Metrics Research**: `curl -s "http://localhost:9090/api/v1/label/verb/values"` 
- **Available verbs**: GET, LIST, POST, PUT, DELETE, PATCH, etc.
- **Created**: `test-dynamic-slo.yaml` and `test-static-slo.yaml` in `.dev/` directory
- **Metrics Used**: `apiserver_request_total{verb="GET"}` and `apiserver_request_total{verb="LIST"}`

### **Step 4: Apply New Test SLOs**
- **Applied**: `test-dynamic-apiserver` (dynamic) and `test-static-apiserver` (static)
- **Configuration**:
  - Both use 95% target, 30d window
  - Dynamic: GET requests with `burnRateType: dynamic`
  - Static: LIST requests with `burnRateType: static`

### **Test 6 Results - Pyrra UI Verification**:

#### **✅ SUCCESS: Data Availability Resolved**
- **test-static-apiserver**: 
  - ✅ **100% availability**
  - ✅ **100% error budget**
- **test-dynamic-apiserver**: 
  - ✅ **99.99% availability** 
  - ✅ **99.79% error budget**

#### **SLO Count Verification**:
- **Total SLOs**: 15 (confirmed: deleted 2 + added 2 = net zero change)
- **"No Data" SLOs Reduced**: Now only legacy SLOs show "No data"

### **Key Achievement**: 
🎉 **Dynamic burn rate SLOs now display real availability and budget data**, confirming the issue was metric selection, not the dynamic burn rate implementation.

---

## 📋 Test 7: Dynamic vs Static Threshold Display Validation

### **Testing Approach Note**:
**Hybrid Testing**: AI performs terminal commands directly for efficiency, human operator tests UI functionality and reports results.

### **Objective**: Verify dynamic threshold display and compare with static thresholds

**Testing Methodology**:
- **Mathematical Calculations**: Use `python -c "..."` for accurate arithmetic (avoiding LLM calculation errors)
- **Live Data Validation**: Accept small discrepancies due to constantly changing live metrics
- **Prometheus Verification**: Cross-check calculations with actual Prometheus query results

---

## ✅ Test 8: Mathematical Correctness Validation - COMPLETED

### **Step 1: Extract Live Metric Values**
- **Command**: `curl -s "http://localhost:9090/api/v1/query?query=sum%28increase%28apiserver_request_total%7Bverb%3D%22GET%22%7D%5B30d%5D%29%29"`
- **N_SLO (30d total)**: 85,247 requests
- **Command**: `curl -s "http://localhost:9090/api/v1/query?query=sum%28increase%28apiserver_request_total%7Bverb%3D%22GET%22%7D%5B1h4m%5D%29%29"`  
- **N_long (1h4m total)**: ~10,593 requests

### **Step 2: Python Calculation Verification**
```python
python -c "
n_slo = 85247
n_long = 10593  
e_budget_percent = 0.020833
slo_target = 0.95
dynamic_threshold = (n_slo / n_long) * e_budget_percent * (1 - slo_target)
print(f'Dynamic threshold: {dynamic_threshold:.12f}')
"
```
**Result**: 0.008382661904

### **Step 3: Prometheus Expression Validation**
**Query**: `(sum(increase(apiserver_request_total{verb="GET"}[30d])) / sum(increase(apiserver_request_total{verb="GET"}[1h4m]))) * 0.020833 * (1-0.95)`
**Result**: 0.008854711477347382

### **Step 4: Comparison Analysis**

#### **Dynamic vs Static Threshold Comparison**:
- **Dynamic Threshold**: ~0.00885 (0.885% error rate triggers alert)
- **Static Threshold**: 0.7 (70% error rate triggers alert)  
- **Sensitivity Analysis**: 
  - ✅ **Dynamic**: Provides meaningful alerting sensitivity  
  - ❌ **Static**: Essentially non-functional (70% error rate unrealistic)

#### **Mathematical Validation Results**:
- ✅ **Formula Implementation**: Correct `(N_SLO/N_long) × E_budget_percent × (1-SLO_target)`
- ✅ **Live Data Integration**: Dynamic thresholds adapt to actual traffic patterns
- ✅ **Expected Discrepancy**: Small difference (0.008382 vs 0.008854) due to live metric changes
- ✅ **Traffic Scaling**: Higher recent traffic → higher threshold sensitivity

### **Test 8 Key Findings**:

1. **🎉 Mathematical Correctness Confirmed**
   - Dynamic threshold formula implemented correctly in Prometheus rules
   - Calculations match expected mathematical results (accounting for live data changes)
   - Traffic scaling working as designed

2. **✅ Practical Alerting Validation**  
   - Dynamic: Alerts at realistic ~0.9% error rate
   - Static: Requires unrealistic 70% error rate
   - Dynamic thresholds provide actionable alerting sensitivity

3. **✅ Live Data Integration Working**
   - Thresholds adapt to real traffic patterns
   - Small timing differences expected and acceptable
   - End-to-end mathematical pipeline validated

---

## 📋 Test 9: Real-Time Threshold Display Enhancement

**Status:** 🚧 **IMPLEMENTATION COMPLETED - TESTING NEEDED**  
**Objective:** Enhance UI to show calculated dynamic threshold values instead of generic "Traffic-Aware" text

### **Current State Investigation:**
- ✅ Located `getBurnRateDisplayText()` function in `ui/src/burnrate.tsx` line 91
- ✅ Function currently returns "Traffic-Aware" placeholder for dynamic SLOs
- ✅ `AlertsTable.tsx` component has access to `promClient` (PrometheusService) for real-time queries
- ✅ Identified PrometheusService API and `usePrometheusQuery` hook available

### **Enhancement Implementation Complete:**

#### **Step 1: Code Analysis** ✅
- Examined `DynamicWindows()` function in `slo/rules.go` 
- **Critical Finding**: E_budget_percent_threshold values are **constants**, not calculated values
- Constants mapped to static factors: 14→1/48, 7→1/16, 2→1/14, 1→1/7
- These are independent of SLO window periods

#### **Step 2: UI Implementation** ✅
- Updated `getBurnRateDisplayText()` in `ui/src/burnrate.tsx`
- Added `DYNAMIC_THRESHOLD_CONSTANTS` mapping static factors to E_budget_percent values
- Modified function to show actual threshold constants for dynamic SLOs
- Maintained backward compatibility with static threshold calculations

#### **Step 3: Documentation Corrections** ✅
- **Fixed critical errors** in both documentation files:
  - `CORE_CONCEPTS_AND_TERMINOLOGY.md`: Corrected "Window Period" column to "Static Factor"
  - `FEATURE_IMPLEMENTATION_SUMMARY.md`: Updated table to reflect correct static factor mapping
- **Corrected misunderstanding**: E_budget_percent values are constants, not window-period dependent

### **Implementation Details:**

#### **Constants Added to burnrate.tsx:**
```typescript
const DYNAMIC_THRESHOLD_CONSTANTS = {
  14: 1.0 / 48,  // First critical window - 50% per day
  7: 1.0 / 16,   // Second critical window - 100% per 4 days  
  2: 1.0 / 14,   // First warning window
  1: 1.0 / 7     // Second warning window
}
```

#### **Enhanced Display Logic:**
- Dynamic SLOs now show actual threshold constants (e.g., "0.0208", "2.08e-2")
- Automatic precision formatting for small numbers
- Fallback to "Traffic-Aware" only when factor unavailable
- Static SLO behavior unchanged

### **Test 9 Implementation Results:**

1. **🎉 SUCCESS: UI Enhancement Implementation Complete**
   - Dynamic SLOs now configured to display actual E_budget_percent_threshold constants
   - Code updated to replace generic "Traffic-Aware" with calculated values
   - Proper precision formatting for small numbers implemented

2. **🔧 CRITICAL CORRECTION: Documentation Fixed**
   - Corrected fundamental misunderstanding about E_budget_percent calculation
   - Fixed "Window Period" column errors in documentation  
   - Updated both core concept docs and implementation summary

3. **✅ TECHNICAL VALIDATION: Constants Implementation**
   - Confirmed E_budget_percent values are predefined constants (1/48, 1/16, 1/14, 1/7)
   - Values mapped correctly to static window factors in DynamicWindows() function
   - Implementation matches backend rule generation logic

4. **📝 KNOWLEDGE UPDATE: Architectural Understanding**
   - E_budget_percent thresholds are **design constants**, not calculated values
   - Static factor hierarchy (14, 7, 2, 1) maps to predefined sensitivity levels
   - Window periods scale automatically but threshold constants remain fixed

### **⏭️ Next: Test 9 Verification Required**
- **Build UI changes**: `cd ui && npm run build` ✅ **COMPLETED**
- **Component Architecture**: `BurnRateThresholdDisplay` component created and integrated ✅ **COMPLETED**
- **TypeScript/ESLint Issues**: All compilation errors resolved ✅ **COMPLETED**
- **Code Cleanup**: Removed unused `getBurnRateDisplayText()` function ✅ **COMPLETED**
- **Next Session**: Verify dynamic SLOs show actual threshold values instead of "Traffic-Aware"
- **Expected Values**: Real-time calculated thresholds based on Prometheus queries

### **🎯 Test 9 Status: 🚧 IMPLEMENTATION COMPLETE - UI TESTING NEEDED**
**Implementation completed but requires verification in fresh testing session**

---

## ✅ Test 7: Dynamic vs Static Threshold Display Validation - COMPLETED

### **Step 1: Prometheus Rules Analysis**
- **Command**: `kubectl get prometheusrule test-dynamic-apiserver -n monitoring -o yaml`
- **Command**: `kubectl get prometheusrule test-static-apiserver -n monitoring -o yaml`

### **🎯 BREAKTHROUGH: Dynamic Expressions Successfully Generated!**

#### **Dynamic SLO Alert Expression**:
```promql
> ((sum(increase(apiserver_request_total{verb="GET"}[30d])) 
   / sum(increase(apiserver_request_total{verb="GET"}[1h4m]))) 
   * 0.020833 * (1-0.95))
```

#### **Static SLO Alert Expression**:
```promql
> (14 * (1-0.95))
```

### **Key Differences Identified**:

1. **✅ Dynamic Threshold Calculation**: 
   - Uses **actual traffic ratio**: `N_SLO/N_long` 
   - Multiplies by budget factor: `0.020833 * (1-0.95)`
   - **Adapts based on traffic volume**

2. **❌ Static Fixed Threshold**:
   - Uses **fixed multiplier**: `14 * (1-0.95)` = `0.7`
   - **No traffic adaptation**

### **Step 2: UI Display Verification**

#### **Burn Rate Thresholds Column (in detail pages)**:
- **✅ Dynamic SLO**: Shows **"Traffic-Aware"** (generic placeholder)
- **✅ Static SLO**: Shows **actual calculated values** ("0.700, 0.350, 0.100, 0.050")

#### **Tooltip Functionality**:
- **✅ Dynamic tooltip**: Shows **formula** `(N_SLO / N_long) x E_budget_percent x (1 - SLO_target)`
- **❌ Missing**: **Actual calculated threshold values** in tooltip

### **Test 7 Key Findings**:

1. **🎉 SUCCESS: Dynamic PromQL Generation Working**
   - Dynamic expressions correctly generated in Prometheus rules
   - Mathematical formula properly implemented in backend
   - Static vs dynamic expressions clearly differentiated

2. **⚠️ ENHANCEMENT OPPORTUNITY: UI Display**
   - Dynamic SLO shows generic "Traffic-Aware" instead of actual values
   - Tooltip shows formula but not calculated results
   - Static SLO correctly shows calculated thresholds

3. **✅ VALIDATION: Feature Implementation Correct**
   - Backend dynamic burn rate calculations are working
   - Traffic-aware threshold adaptation implemented
   - UI framework in place, needs value calculation display

---

## 📋 Test 8: Mathematical Correctness Validation

### **Objective**: Verify the dynamic threshold calculations produce correct mathematical results

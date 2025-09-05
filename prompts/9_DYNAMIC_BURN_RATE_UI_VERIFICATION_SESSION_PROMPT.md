# Dynamic Burn Rate UI Verification Session Prompt

## Session Context
Previous sessions successfully completed **comprehensive data infrastructure validation** and **mathematical correctness verification** for the dynamic burn rate feature. **Test 9 implementation is complete** but requires final UI verification to confirm the real-time threshold display enhancement is working correctly.

## ✅ Previous Session Achievements (September 5, 2025)
- **✅ Tests 1-8 COMPLETED**: Complete data infrastructure validation, metric switching, and mathematical correctness verification
- **✅ Real Data Integration**: Switched to `apiserver_request_total` metrics with actual error data (71 series, multiple error codes)
- **✅ Mathematical Validation**: Confirmed dynamic threshold formula `(N_SLO/N_long) × E_budget_percent × (1-SLO_target)` working correctly
- **✅ Live Data Confirmed**: Dynamic thresholds ~0.00885 (0.885% error rate) vs Static 0.7 (70% error rate) - dynamic provides meaningful sensitivity
- **✅ Test 9 Implementation**: `BurnRateThresholdDisplay` component created, integrated, and built successfully

## ✅ Test 9 Implementation Status - READY FOR VERIFICATION
**Component Architecture Complete**:
- **✅ BurnRateThresholdDisplay.tsx**: Component handles both static and dynamic threshold display
- **✅ React Hooks Fixed**: Proper conditional hook execution, no React Hooks rules violations
- **✅ TypeScript/ESLint Clean**: All compilation errors resolved, builds successfully
- **✅ Integration Complete**: Component integrated in AlertsTable.tsx with proper imports
- **✅ Code Cleanup**: Removed unused `getBurnRateDisplayText()` function and constants

**Implementation Details**:
```tsx
// Static SLOs: Shows calculated threshold using factor * (1 - target)
if (burnRateType === BurnRateType.Static && factor !== undefined) {
  const threshold = factor * (1 - targetDecimal)
  return <span>{threshold.toFixed(5)}</span>
}

// Dynamic SLOs: Makes real-time Prometheus queries for dynamic calculations
if (burnRateType === BurnRateType.Dynamic) {
  return <DynamicThresholdValue objective={objective} promClient={promClient} />
}
```

## 🎯 Session Objectives - UI Verification Focus

### **Primary Objective: Test 9 Completion**
**Verify Real-Time Dynamic Threshold Display Enhancement**

#### **Expected Behavior**:
1. **Static SLOs**: Should show calculated thresholds (e.g., "0.70000", "0.35000", "0.10000", "0.05000" for 95% SLO)
2. **Dynamic SLOs**: Should show **real-time calculated values** instead of "Traffic-Aware" placeholder
3. **Dynamic Calculation**: Uses live Prometheus queries with formula: `(shortErrorRate / longErrorRate) × errorBudgetPercent × (1 - target)`

#### **Test Scenarios**:
1. **UI Load Test**: Open `http://localhost:9099`, navigate to AlertsTable
2. **Static SLO Verification**: Confirm static thresholds display correctly (regression test)
3. **Dynamic SLO Verification**: Check if `test-dynamic-apiserver` shows calculated values or "Traffic-Aware"
4. **Tooltip Verification**: Confirm tooltips show calculation details for dynamic SLOs
5. **Real-Time Updates**: Verify if dynamic values update based on live Prometheus data

### **Secondary Objectives: Comprehensive Validation**

#### **Performance Validation**:
- Monitor browser console for errors during component rendering
- Verify Prometheus query performance doesn't impact UI responsiveness
- Check if multiple dynamic SLOs cause performance issues

#### **Edge Case Testing**:
- **No Data Scenarios**: How component behaves when Prometheus queries fail
- **Zero Traffic**: Behavior when `longErrorRate = 0` (division by zero protection)
- **Loading States**: Component behavior during async Prometheus queries

#### **Cross-Browser Compatibility**:
- Test in primary browser used for development
- Verify component renders consistently

## 🧪 Testing Methodology

### **UI-First Approach**
- **Primary Method**: Human operator interaction with actual UI at `http://localhost:9099`
- **Secondary**: Browser developer tools inspection for errors/network issues
- **Documentation**: Screenshots and detailed behavior descriptions

### **Expected Outcomes Matrix**

| Scenario | Expected Result | Success Criteria |
|----------|----------------|------------------|
| Static SLO Display | "0.70000, 0.35000, 0.10000, 0.05000" | ✅ Calculated values shown |
| Dynamic SLO Loading | Loading indicator or "Traffic-Aware" | ✅ No errors in console |
| Dynamic SLO Success | Real calculated values (e.g., "0.00885") | 🎉 **PRIMARY SUCCESS** |
| Dynamic SLO Failure | "Traffic-Aware" fallback | ⚠️ Acceptable but needs investigation |
| Tooltip Display | Calculation details on hover | ✅ Enhanced user experience |

### **Troubleshooting Scenarios**

#### **If Dynamic Values Don't Appear**:
1. **Browser Console Inspection**: Check for JavaScript errors or network failures
2. **Prometheus Query Validation**: Verify queries work in Prometheus UI
3. **Component State Debugging**: Check if `usePrometheusQuery` hook receives data
4. **Fallback Behavior**: Confirm "Traffic-Aware" appears instead of crashes

#### **If Static Values Break**:
1. **Regression Analysis**: Component should maintain backward compatibility
2. **Factor Calculation**: Verify `factor * (1 - target)` logic unchanged
3. **UI Integration**: Confirm AlertsTable still calls component correctly

## 📋 Success Criteria

### **Minimum Success** (Component Integration Working)
- ✅ UI loads without errors at `http://localhost:9099`
- ✅ Static SLO thresholds display correctly (no regression)
- ✅ Dynamic SLOs show some indication (even if "Traffic-Aware")
- ✅ No JavaScript console errors

### **Full Success** (Real-Time Dynamic Display Working)
- 🎉 **Dynamic SLOs show calculated threshold values** instead of "Traffic-Aware"
- ✅ Values update based on real Prometheus data
- ✅ Tooltips provide calculation details
- ✅ Performance remains acceptable

### **Production Readiness** (Complete Feature Validation)
- ✅ Multiple dynamic SLOs work correctly
- ✅ Edge cases handled gracefully
- ✅ Component provides actual value over static approach
- ✅ Documentation complete for deployment

## 🔧 Expected Challenges & Solutions

### **Challenge 1: Prometheus Query Complexity**
- **Issue**: Real-time queries may be complex or slow
- **Detection**: Long loading times or "Traffic-Aware" fallbacks
- **Investigation**: Test queries directly in Prometheus UI
- **Solution**: Query optimization or simplified calculation approach

### **Challenge 2: React Hook Dependencies**
- **Issue**: `usePrometheusQuery` may have complex dependency requirements
- **Detection**: Console errors about hook rules or dependency arrays
- **Investigation**: Component state debugging with React DevTools
- **Solution**: Proper dependency management and error boundaries

### **Challenge 3: Metric Data Availability**
- **Issue**: Live Prometheus data may be insufficient for calculations
- **Detection**: "Traffic-Aware" fallbacks even with working queries
- **Investigation**: Check if `apiserver_request_total` has sufficient error rate data
- **Solution**: Metric selection adjustment or longer time windows

### **Challenge 4: Mathematical Edge Cases**
- **Issue**: Division by zero or extreme values
- **Detection**: NaN values, Infinity, or calculation errors in tooltips
- **Investigation**: Console logging of intermediate calculation values
- **Solution**: Input validation and boundary condition handling

## 🎯 Testing Environment Requirements

### **Prerequisites**
- **Services Running**: `./pyrra kubernetes` (port 9444) + `./pyrra api` (port 9099)
- **Test SLOs Deployed**: `test-dynamic-apiserver` and `test-static-apiserver` with real data
- **UI Build Current**: Recent `npm run build` completed successfully
- **Browser Ready**: Developer tools open for console monitoring

### **Test Data Context**
- **Dynamic SLO**: Uses `apiserver_request_total{verb="GET"}` with confirmed error data
- **Static SLO**: Uses `apiserver_request_total{verb="LIST"}` for comparison
- **Expected Traffic Ratio**: ~8:1 (85,247 / 10,593) based on previous calculations
- **Expected Dynamic Threshold**: ~0.00885 (0.885% error rate)

## 📚 Reference Materials

### **Implementation Context**
- **Previous Session**: DYNAMIC_BURN_RATE_TESTING_SESSION.md (Tests 1-8 complete)
- **Mathematical Validation**: Dynamic threshold formula confirmed working in Test 8
- **Component Architecture**: BurnRateThresholdDisplay.tsx with real-time Prometheus integration

### **Expected Queries in Component**
```typescript
// Short window query (1 hour)
const shortQuery = `rate(${errorMetric}[1h]) / rate(${totalMetric}[1h])`

// Long window query (SLO window, e.g., 30d)  
const longQuery = `rate(${errorMetric}[${windowSeconds}s]) / rate(${totalMetric}[${windowSeconds}s])`

// Dynamic threshold calculation
const dynamicThreshold = (shortErrorRate / longErrorRate) * errorBudgetPercent * (1 - target)
```

### **UI Navigation Path**
1. Open `http://localhost:9099`
2. Navigate to SLO list
3. Click on `test-dynamic-apiserver` or similar dynamic SLO
4. Examine AlertsTable threshold display column
5. Check tooltips and interactive elements

## 🎯 Next Actions Sequence

1. **Start Fresh Session**: Clear terminal, restart services if needed
2. **UI Verification**: Open Pyrra UI and navigate to dynamic SLO
3. **Threshold Display Check**: Verify actual calculated values appear
4. **Regression Testing**: Confirm static SLOs still work correctly
5. **Documentation Update**: Record final results and mark Test 9 complete
6. **Success Assessment**: Determine if feature is ready for production deployment

## 📋 Session Completion Criteria

### **Test 9 Complete**: 
- ✅ Dynamic threshold display verified working or fallback behavior documented
- ✅ Static SLO functionality confirmed unchanged (no regression)
- ✅ Browser console free of critical errors
- ✅ Component performance acceptable

### **Feature Complete**:
- 🎉 **Dynamic burn rate feature fully validated** from mathematical correctness through UI display
- ✅ Production deployment guidance updated
- ✅ All documentation current and accurate

---

**Expected Outcome**: Final verification of dynamic burn rate feature UI integration, completing the comprehensive validation process started in previous sessions. Feature should be ready for production deployment or have clear guidance for remaining work.

# Browser Compatibility Testing Guide

## Overview

This guide provides step-by-step instructions for testing the dynamic burn rate feature across different browsers to ensure cross-browser compatibility.

## Test Environment Setup

### Prerequisites
- Pyrra services running (`./pyrra api` on port 9099)
- Test SLOs deployed (mix of static and dynamic)
- Access to multiple browsers for testing

### Test URLs
- **Embedded UI**: http://localhost:9099
- **Development UI**: http://localhost:3000 (if running `npm start`)

## Browser Test Matrix

| Browser | Version | Platform | Priority | Status |
|---------|---------|----------|----------|--------|
| Chrome | Latest | Windows | HIGH | ⬜ Not Tested |
| Firefox | Latest | Windows | HIGH | ⬜ Not Tested |
| Edge | Latest | Windows | MEDIUM | ⬜ Not Tested |
| Safari | Latest | macOS | MEDIUM | ⬜ Not Tested |
| Chrome | Latest | Linux | LOW | ⬜ Not Tested |

## Test Scenarios

### Test 1: Visual Rendering

**Objective**: Verify all UI components render correctly

**Steps**:
1. Open http://localhost:9099 in browser
2. Navigate to SLO list page
3. Check burn rate column displays correctly
4. Verify badge colors (green for dynamic, gray for static)
5. Check icons display correctly (eye for dynamic, lock for static)
6. Verify table layout is not broken

**Expected Results**:
- ✅ Burn rate column visible and properly aligned
- ✅ Dynamic badges are green with eye icons
- ✅ Static badges are gray with lock icons
- ✅ No visual artifacts or layout issues
- ✅ Text is readable and properly sized

**Browser-Specific Notes**:
- **Chrome**: Reference browser, should work perfectly
- **Firefox**: May have slight font rendering differences
- **Edge**: Should match Chrome (Chromium-based)
- **Safari**: May have different tooltip positioning

### Test 2: Interactive Tooltips

**Objective**: Verify tooltips work correctly on hover

**Steps**:
1. Hover over dynamic SLO badge
2. Verify tooltip appears with traffic-aware description
3. Hover over static SLO badge
4. Verify tooltip appears with static description
5. Move mouse away and verify tooltip disappears
6. Test tooltip positioning near screen edges

**Expected Results**:
- ✅ Tooltips appear smoothly on hover
- ✅ Tooltips disappear when mouse moves away
- ✅ Tooltip content is readable and properly formatted
- ✅ Tooltips don't overflow screen boundaries
- ✅ No flickering or positioning issues

**Browser-Specific Notes**:
- **Chrome**: Bootstrap tooltips should work perfectly
- **Firefox**: May have slight animation differences
- **Edge**: Should match Chrome behavior
- **Safari**: May need webkit-specific CSS adjustments

### Test 3: Column Sorting

**Objective**: Verify table sorting works correctly

**Steps**:
1. Click on "Burn Rate" column header
2. Verify table re-sorts (dynamic first or static first)
3. Click again to reverse sort order
4. Verify sorting arrow indicator appears
5. Test sorting other columns to ensure no conflicts

**Expected Results**:
- ✅ Clicking column header triggers sort
- ✅ Sort arrow indicator appears
- ✅ Table re-orders correctly
- ✅ No JavaScript errors in console
- ✅ Sorting is stable and consistent

**Browser-Specific Notes**:
- **All browsers**: Should use react-table sorting, no browser-specific issues expected

### Test 4: Column Visibility Toggle

**Objective**: Verify column visibility controls work

**Steps**:
1. Click "Columns" dropdown button
2. Verify "Burn Rate" checkbox is present
3. Uncheck "Burn Rate" checkbox
4. Verify column disappears from table
5. Check "Burn Rate" checkbox again
6. Verify column reappears

**Expected Results**:
- ✅ Dropdown opens and closes correctly
- ✅ Checkbox state reflects column visibility
- ✅ Column hides/shows smoothly
- ✅ Table layout adjusts properly
- ✅ No visual glitches during transition

**Browser-Specific Notes**:
- **All browsers**: Bootstrap dropdown should work consistently

### Test 5: Navigation and Detail Pages

**Objective**: Verify navigation between pages works

**Steps**:
1. Click on a dynamic SLO name
2. Verify detail page loads
3. Check burn rate type badge displays on detail page
4. Verify threshold display component works
5. Navigate back to list page
6. Verify state is preserved

**Expected Results**:
- ✅ Navigation works without errors
- ✅ Detail page loads completely
- ✅ Dynamic burn rate information displays correctly
- ✅ Back navigation works
- ✅ No memory leaks or performance issues

**Browser-Specific Notes**:
- **All browsers**: React Router should handle navigation consistently

### Test 6: API Communication

**Objective**: Verify API communication works correctly

**Steps**:
1. Open browser developer tools (F12)
2. Go to Network tab
3. Refresh SLO list page
4. Verify API request to `/objectives.v1alpha1.ObjectiveService/List`
5. Check response includes `burnRateType` field
6. Verify no CORS errors
7. Check console for JavaScript errors

**Expected Results**:
- ✅ API request completes successfully (200 OK)
- ✅ Response includes burnRateType field
- ✅ No CORS errors
- ✅ No JavaScript errors in console
- ✅ Connect/gRPC-Web protocol works

**Browser-Specific Notes**:
- **Chrome**: DevTools most comprehensive
- **Firefox**: Similar DevTools functionality
- **Edge**: Chromium-based, similar to Chrome
- **Safari**: Different DevTools interface but same functionality

### Test 7: Performance and Memory

**Objective**: Verify browser performance is acceptable

**Steps**:
1. Open browser developer tools
2. Go to Performance/Memory tab
3. Record performance profile while navigating
4. Check memory usage over time
5. Navigate between pages multiple times
6. Verify no memory leaks

**Expected Results**:
- ✅ Page loads within 3 seconds
- ✅ Smooth scrolling and interactions
- ✅ Memory usage stable (no continuous growth)
- ✅ No performance warnings in console
- ✅ CPU usage reasonable

**Browser-Specific Notes**:
- **Chrome**: Best performance profiling tools
- **Firefox**: Good memory profiling
- **Edge**: Similar to Chrome
- **Safari**: Different profiling interface

### Test 8: Error Handling

**Objective**: Verify error states display correctly

**Steps**:
1. Stop Pyrra API service
2. Refresh browser page
3. Verify error message displays
4. Restart Pyrra API service
5. Refresh page
6. Verify recovery works

**Expected Results**:
- ✅ Error message displays when API unavailable
- ✅ No blank pages or crashes
- ✅ Recovery works when service restored
- ✅ Appropriate loading states shown
- ✅ User-friendly error messages

**Browser-Specific Notes**:
- **All browsers**: Error handling should be consistent

## Testing Checklist

Use this checklist to track testing progress:

### Chrome (Windows)
- [ ] Test 1: Visual Rendering
- [ ] Test 2: Interactive Tooltips
- [ ] Test 3: Column Sorting
- [ ] Test 4: Column Visibility Toggle
- [ ] Test 5: Navigation and Detail Pages
- [ ] Test 6: API Communication
- [ ] Test 7: Performance and Memory
- [ ] Test 8: Error Handling

### Firefox (Windows)
- [ ] Test 1: Visual Rendering
- [ ] Test 2: Interactive Tooltips
- [ ] Test 3: Column Sorting
- [ ] Test 4: Column Visibility Toggle
- [ ] Test 5: Navigation and Detail Pages
- [ ] Test 6: API Communication
- [ ] Test 7: Performance and Memory
- [ ] Test 8: Error Handling

### Edge (Windows)
- [ ] Test 1: Visual Rendering
- [ ] Test 2: Interactive Tooltips
- [ ] Test 3: Column Sorting
- [ ] Test 4: Column Visibility Toggle
- [ ] Test 5: Navigation and Detail Pages
- [ ] Test 6: API Communication
- [ ] Test 7: Performance and Memory
- [ ] Test 8: Error Handling

### Safari (macOS) - If Available
- [ ] Test 1: Visual Rendering
- [ ] Test 2: Interactive Tooltips
- [ ] Test 3: Column Sorting
- [ ] Test 4: Column Visibility Toggle
- [ ] Test 5: Navigation and Detail Pages
- [ ] Test 6: API Communication
- [ ] Test 7: Performance and Memory
- [ ] Test 8: Error Handling

## Known Issues and Workarounds

### Issue 1: Tooltip Positioning in Safari
**Description**: Tooltips may position incorrectly near screen edges in Safari
**Workaround**: Bootstrap tooltips should auto-adjust, but may need webkit-specific CSS
**Status**: To be tested

### Issue 2: Font Rendering in Firefox
**Description**: Slight font rendering differences compared to Chrome
**Workaround**: Acceptable difference, no action needed
**Status**: Expected behavior

### Issue 3: DevTools Differences
**Description**: Different browsers have different DevTools interfaces
**Workaround**: Adapt testing approach to each browser's tools
**Status**: Not an issue, just different interfaces

## Reporting Issues

When reporting browser-specific issues, include:

1. **Browser and Version**: e.g., "Firefox 120.0 on Windows 11"
2. **Test Scenario**: Which test failed
3. **Expected Behavior**: What should happen
4. **Actual Behavior**: What actually happened
5. **Screenshots**: Visual evidence of the issue
6. **Console Errors**: Any JavaScript errors from console
7. **Network Errors**: Any API communication issues
8. **Reproducibility**: Steps to reproduce the issue

## Success Criteria

**Minimum Success**:
- ✅ All tests pass in Chrome and Firefox
- ✅ No critical visual issues in Edge
- ✅ No JavaScript errors in any browser

**Full Success**:
- ✅ All tests pass in Chrome, Firefox, and Edge
- ✅ Safari testing completed (if available)
- ✅ Performance acceptable in all browsers
- ✅ Comprehensive issue documentation

**Production Ready**:
- ✅ All browsers tested and documented
- ✅ Known issues have workarounds
- ✅ Browser compatibility matrix complete
- ✅ Recommendations documented for users

## Next Steps

1. **Start with Chrome** - Establish baseline behavior
2. **Test Firefox** - Identify any cross-browser issues
3. **Test Edge** - Should match Chrome (Chromium-based)
4. **Test Safari** - If macOS available, test webkit-specific issues
5. **Document findings** - Update this guide with results
6. **Create issue list** - Document any browser-specific problems
7. **Implement fixes** - Address critical issues if found

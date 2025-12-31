# Testing Workflow

Complete testing checklist for the Story Engine Obsidian plugin.

## Prerequisites

Before testing, ensure:

- [x] Backend running on :8080
  ```bash
  cd main-service
  make run-http
  ```

- [ ] Database with test tenant created
  ```bash
  curl -X POST http://localhost:8080/api/v1/tenants \
    -H "Content-Type: application/json" \
    -d '{"name": "Test Workspace"}'
  ```
  Save the tenant ID from response.
  
  1c645fd8-95ca-48c4-8dfd-11003f55fd18

- [x] Plugin built and linked
  ```bash
  cd obsidian-plugin
  npm run build
  # Link to vault (see DEVELOPMENT.md)
  ```

- [ ] Obsidian loaded with plugin enabled
  - Settings → Community plugins → Story Engine → ON

## Test Cases

### 1. Settings Configuration

**Test:** Can configure plugin settings

- [ ] Open Settings → Story Engine
- [ ] Enter API URL: `http://localhost:8080`
- [ ] Enter API Key: (can be empty for MVP)
- [ ] Enter Tenant ID: (from curl test)
- [ ] Click "Test Connection"
- [ ] Should see "✓ Connected" in green
- [ ] Close settings
- [ ] Reopen settings
- [ ] All values should persist

**Expected:** Settings save and persist correctly

---

### 2. List Stories

**Test:** Can list all stories

- [ ] Press `Cmd+P` / `Ctrl+P`
- [ ] Type "Story Engine: List Stories"
- [ ] Press Enter
- [ ] Modal should open
- [ ] If no stories exist:
  - [ ] Should show "No stories found"
  - [ ] Should show "Create Story" button
- [ ] If stories exist:
  - [ ] Should show list of stories
  - [ ] Each story shows title, version, status
  - [ ] Can click on story to view details

**Expected:** Stories load from API and display correctly

**Error Cases:**
- [ ] Test with invalid tenant ID → Should show error
- [ ] Test with backend down → Should show error
- [ ] Test with network error → Should handle gracefully

---

### 3. Create Story

**Test:** Can create a new story

- [ ] Press `Cmd+P` / `Ctrl+P`
- [ ] Type "Story Engine: Create Story"
- [ ] Press Enter
- [ ] Prompt appears asking for title
- [ ] Enter title: "Test Story"
- [ ] Press OK
- [ ] Story should be created
- [ ] Details modal should open automatically
- [ ] Notification should appear: "Story 'Test Story' created successfully"
- [ ] List stories again → New story should appear

**Expected:** Story created successfully and appears in list

**Error Cases:**
- [ ] Cancel prompt → Nothing happens
- [ ] Empty title → Should show validation error
- [ ] Invalid tenant ID → Should show error notification
- [ ] Backend error → Should show error notification

---

### 4. View Story Details

**Test:** Can view story information

- [ ] From story list, click on a story
- [ ] Details modal opens
- [ ] Verify all fields displayed:
  - [ ] Title matches
  - [ ] Status displayed
  - [ ] Version number displayed
  - [ ] Created date formatted correctly
  - [ ] Updated date formatted correctly
  - [ ] Story ID displayed
- [ ] "Clone Story" button visible
- [ ] "Copy ID" button visible

**Expected:** All story information displayed correctly

---

### 5. Clone Story

**Test:** Can clone a story to create new version

- [ ] Open story details
- [ ] Note the current version number
- [ ] Click "Clone Story" button
- [ ] Button should show "Cloning..." and be disabled
- [ ] New version should be created
- [ ] Details modal for new story should open
- [ ] Verify:
  - [ ] Version number incremented
  - [ ] Root story ID matches original
  - [ ] Previous story ID matches original story ID
- [ ] List stories → Both versions should appear

**Expected:** New version created with incremented version number

**Error Cases:**
- [ ] Clone non-existent story → Should show error
- [ ] Backend error → Should show error message

---

### 6. Copy Story ID

**Test:** Can copy story ID to clipboard

- [ ] Open story details
- [ ] Click "Copy ID" button
- [ ] Button should change to "Copied!"
- [ ] After 2 seconds, should revert to "Copy ID"
- [ ] Paste clipboard → Should contain story ID

**Expected:** Story ID copied to clipboard

---

### 7. Test Connection Button

**Test:** Connection test works correctly

- [ ] Open Settings → Story Engine
- [ ] Enter valid API URL
- [ ] Click "Test Connection"
- [ ] Should show "✓ Connected" in green
- [ ] Change to invalid URL
- [ ] Click "Test Connection"
- [ ] Should show "✗ Failed" in red
- [ ] After 3 seconds, button resets

**Expected:** Connection test accurately reflects backend status

---

## End-to-End Flow Test

**Complete workflow:**

1. [ ] Configure plugin with valid settings
2. [ ] Create a new story via plugin
3. [ ] View story details
4. [ ] Clone the story
5. [ ] Verify both versions appear in list
6. [ ] View details of cloned story
7. [ ] Verify version number incremented
8. [ ] Copy story ID
9. [ ] Use curl to verify story exists in backend

**Expected:** Complete workflow works without errors

---

## Performance Testing

- [ ] List stories with 10+ stories → Should load quickly
- [ ] Create story → Should complete in < 2 seconds
- [ ] Clone story → Should complete in < 5 seconds
- [ ] Multiple rapid clicks → Should handle gracefully

---

## Browser Compatibility

Since Obsidian uses Electron:

- [ ] Test on macOS
- [ ] Test on Windows
- [ ] Test on Linux
- [ ] Verify fetch API works (should work in Electron)

---

## Regression Testing

After making changes:

1. [ ] Run all test cases above
2. [ ] Check console for errors
3. [ ] Check network tab for failed requests
4. [ ] Verify no console warnings

---

## Reporting Issues

When reporting issues, include:

1. **Steps to reproduce**
2. **Expected behavior**
3. **Actual behavior**
4. **Console errors** (from developer console)
5. **Network errors** (from Network tab)
6. **Backend logs** (if relevant)
7. **Obsidian version**
8. **Plugin version** (from manifest.json)

---

## Success Criteria

All tests pass when:

- ✅ All checkboxes above are checked
- ✅ No console errors
- ✅ No network errors
- ✅ All features work as expected
- ✅ Error handling works correctly
- ✅ UI is responsive and intuitive


# Debugging Guide

Troubleshooting common issues when developing the Story Engine Obsidian plugin.

## Opening Developer Console

The Obsidian developer console is essential for debugging:

1. Open Obsidian
2. Go to View → Toggle Developer Tools
   - Mac: `Cmd+Shift+I`
   - Windows/Linux: `Ctrl+Shift+I`
3. The console opens showing:
   - **Console tab:** All logs and errors
   - **Network tab:** HTTP requests/responses
   - **Sources tab:** Source code debugging

## Common Issues

### Plugin Not Loading

**Symptoms:**
- Plugin doesn't appear in Installed plugins list
- Plugin appears but can't be enabled
- Error message when enabling

**Solutions:**

1. **Check files exist:**
   ```bash
   ls -la /path/to/TestVault/.obsidian/plugins/story-engine/
   ```
   Should show: `main.js`, `manifest.json`, `styles.css`

2. **Check manifest.json is valid:**
   ```bash
   cat /path/to/TestVault/.obsidian/plugins/story-engine/manifest.json | python -m json.tool
   ```
   Should parse without errors

3. **Check main.js exists and is built:**
   ```bash
   ls -lh /path/to/TestVault/.obsidian/plugins/story-engine/main.js
   ```
   File should exist and have a recent modification time

4. **Check console for errors:**
   - Open developer console
   - Look for red error messages
   - Common: "Cannot find module", "Syntax error"

5. **Restart Obsidian:**
   - Close Obsidian completely
   - Reopen
   - Try enabling plugin again

### API Errors

**Symptoms:**
- "Failed to load stories"
- Network errors in console
- Connection test fails

**Solutions:**

1. **Check backend is running:**
   ```bash
   curl http://localhost:8080/health
   ```
   Should return `{"status":"ok"}`

2. **Check CORS headers:**
   - Open Network tab in developer console
   - Make a request (e.g., list stories)
   - Check response headers
   - Should include `Access-Control-Allow-Origin: app://obsidian.md`

3. **Check request/response in console:**
   - Network tab shows all requests
   - Click on request to see:
     - Request URL
     - Request headers
     - Response status
     - Response body

4. **Verify API key and tenant ID:**
   - Settings → Story Engine
   - Check API URL is correct
   - Check Tenant ID matches one from backend
   - Try "Test Connection" button

5. **Check console logs:**
   - Look for error messages
   - Common: "Failed to fetch", "NetworkError", "CORS error"

### Build Errors

**Symptoms:**
- `npm run build` fails
- TypeScript errors
- esbuild errors

**Solutions:**

1. **Check TypeScript errors:**
   ```bash
   npm run build
   ```
   Look for TypeScript compilation errors

2. **Check tsconfig.json:**
   - Ensure it's valid JSON
   - Check compiler options

3. **Clear and reinstall:**
   ```bash
   rm -rf node_modules package-lock.json
   npm install
   ```

4. **Check Node version:**
   ```bash
   node --version
   ```
   Should be Node 18+ (check Obsidian requirements)

### Hot Reload Not Working

**Symptoms:**
- Changes not reflected after reload
- Need to restart Obsidian for changes

**Solutions:**

1. **Ensure watch mode is running:**
   ```bash
   npm run dev
   ```
   Should show "watching..." message

2. **Check file is being rebuilt:**
   - Save a TypeScript file
   - Check terminal for build output
   - Check `main.js` modification time

3. **Reload Obsidian properly:**
   - `Cmd+R` / `Ctrl+R` to reload
   - Or close and reopen Obsidian

4. **Check for build errors:**
   - Watch terminal output
   - Fix any TypeScript errors

### Settings Not Persisting

**Symptoms:**
- Settings reset after reload
- Changes don't save

**Solutions:**

1. **Check saveSettings is called:**
   - Look in `src/settings.ts`
   - Ensure `onChange` calls `saveSettings()`

2. **Check console for errors:**
   - Look for save errors
   - Check file permissions

3. **Manually check settings file:**
   ```bash
   cat /path/to/TestVault/.obsidian/plugins/story-engine/data.json
   ```
   Should contain your settings

## Debugging Tips

### Add Console Logs

Add logging to understand flow:

```typescript
console.log("Loading stories...", tenantId);
const stories = await this.apiClient.listStories(tenantId);
console.log("Stories loaded:", stories);
```

### Check Network Requests

1. Open Network tab
2. Filter by "Fetch/XHR"
3. Click on request to see:
   - Request details
   - Response data
   - Headers
   - Timing

### Use Breakpoints

1. Open Sources tab
2. Find your file (may be under "webpack://")
3. Click line number to set breakpoint
4. Trigger action
5. Step through code

### Test API Directly

Before debugging plugin, test API works:

```bash
curl http://localhost:8080/api/v1/stories?tenant_id=<TENANT_ID>
```

If this fails, fix backend first.

## Getting Help

1. Check console for specific error messages
2. Check Network tab for failed requests
3. Verify backend is working with curl
4. Check all files are in correct locations
5. Restart Obsidian and try again


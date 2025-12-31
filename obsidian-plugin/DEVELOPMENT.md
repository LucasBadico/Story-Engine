# Obsidian Plugin Development Guide

Complete guide for setting up and testing the Story Engine Obsidian plugin.

## Installing Obsidian

1. Download Obsidian from https://obsidian.md
2. Install and launch Obsidian
3. Create a test vault:
   - File → New vault
   - Choose a location (e.g., `/Users/you/Documents/TestVault`)
   - Note the vault path - you'll need it later

## Setting up Development Environment

### 1. Navigate to Plugin Directory

```bash
# From repository root
cd obsidian-plugin
```

### 2. Install Dependencies

```bash
npm install
```

This installs:
- TypeScript
- esbuild
- Obsidian types

### 3. Build the Plugin

**One-time build:**
```bash
npm run build
```

**Watch mode (recommended for development):**
```bash
npm run dev
```

This watches for file changes and rebuilds automatically.

### 4. Link Plugin to Obsidian Vault

**Option A: Symbolic Link (Recommended)**

```bash
# Create plugins directory in vault
mkdir -p /path/to/TestVault/.obsidian/plugins/story-engine

# Create symbolic links (from obsidian-plugin directory)
ln -s $(pwd)/main.js /Users/badico/Library/Mobile Documents/iCloud~md~obsidian/Documents/Writing/.obsidian/plugins/story-engine/
ln -s $(pwd)/manifest.json /Users/badico/Library/Mobile Documents/iCloud~md~obsidian/Documents/Writing/.obsidian/plugins/story-engine/
ln -s $(pwd)/styles.css /Users/badico/Library/Mobile Documents/iCloud~md~obsidian/Documents/Writing/.obsidian/plugins/story-engine/ 2>/dev/null || touch styles.css && ln -s $(pwd)/styles.css /path/to/TestVault/.obsidian/plugins/story-engine/
```

**Option B: Copy Files**

```bash
# Create plugins directory
mkdir -p /path/to/TestVault/.obsidian/plugins/story-engine

# Copy files
cp main.js manifest.json /path/to/TestVault/.obsidian/plugins/story-engine/
touch /path/to/TestVault/.obsidian/plugins/story-engine/styles.css
```

**Important:** Replace `/path/to/TestVault` with your actual vault path.

## Loading the Plugin

1. Open Obsidian
2. Open your test vault
3. Go to Settings → Community plugins
4. Turn off "Safe mode" (if enabled)
5. Go to "Installed plugins"
6. Find "Story Engine" in the list
7. Toggle it ON

If the plugin doesn't appear:
- Check that `main.js` and `manifest.json` exist in `.obsidian/plugins/story-engine/`
- Check that `manifest.json` is valid JSON
- Restart Obsidian

## Configuring the Plugin

1. Go to Settings → Story Engine (should appear in the left sidebar)
2. Enter API URL: `http://localhost:8080` (or your backend URL)
3. Enter API Key: (leave empty for MVP, or enter a test key)
4. Enter Tenant ID: (get this from creating a tenant via curl - see Testing_REST_API.md)
5. Click "Test Connection"
6. If successful, you should see "✓ Connected" in green

## Using the Plugin

### List Stories

- Press `Cmd+P` (Mac) or `Ctrl+P` (Windows/Linux)
- Type "Story Engine: List Stories"
- Press Enter
- A modal should open showing all stories

### Create Story

- Press `Cmd+P` / `Ctrl+P`
- Type "Story Engine: Create Story"
- Press Enter
- Enter a title when prompted
- Story is created and details modal opens

### View Story Details

- From the story list, click on any story
- Details modal opens showing:
  - Title, status, version number
  - Created/updated dates
  - Story ID
  - Clone button
  - Copy ID button

### Clone Story

- Open story details
- Click "Clone Story" button
- New version is created
- Details modal for new version opens automatically

## Development Workflow

1. **Start backend:**
   ```bash
   cd ../main-service
   make run-http
   ```

2. **Start plugin watch mode:**
   ```bash
   cd obsidian-plugin
   npm run dev
   ```

3. **Make changes** to TypeScript files in `src/`

4. **Reload Obsidian:**
   - `Cmd+R` (Mac) or `Ctrl+R` (Windows/Linux)
   - Or close and reopen Obsidian

5. **Test your changes**

## File Structure

```
obsidian-plugin/
├── src/
│   ├── main.ts              # Main plugin class
│   ├── types.ts             # TypeScript type definitions
│   ├── settings.ts          # Settings tab UI
│   ├── commands.ts          # Command registration
│   ├── api/
│   │   └── client.ts        # REST API client
│   └── views/
│       ├── StoryListModal.ts      # Story list modal
│       └── StoryDetailsModal.ts   # Story details modal
├── package.json
├── tsconfig.json
├── manifest.json
├── esbuild.config.mjs
└── main.js                  # Generated (don't edit)
```

## Troubleshooting

See DEBUGGING.md for common issues and solutions.


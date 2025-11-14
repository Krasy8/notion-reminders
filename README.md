# Notion Desktop Reminder (Go Version)

A lightweight Go application that checks your Notion database for incomplete reminders and shows desktop notifications on your Arch Linux machine.

## Why Go?

- **Single binary** - No dependencies or runtime needed
- **Fast startup** - Typically <10ms
- **Small footprint** - Binary is only ~2-3MB
- **Easy deployment** - Just copy one file
- **Native performance** - Compiled to machine code

## Features

- ✅ Add reminders from your iPhone using the Notion app
- 📱 Cross-platform (iPhone for input, Linux for notifications)
- ⏰ Shows desktop notifications for pending tasks
- 📅 Displays when each reminder was created
- 🎯 Optional priority levels
- 🔄 Can run on system startup automatically
- 🚀 Fast and efficient (written in Go)

## Prerequisites

- Arch Linux (or any Linux with systemd)
- Go 1.21+ (for building)
- Notion account
- Internet connection

## Setup Guide

### Part 1: Notion Setup

#### 1. Create the Database

1. Open Notion (web or app)
2. Create a new page called "Desktop Reminders"
3. Add a **Table** database
4. Set up these columns:
   - **Name** (Title) - This is created automatically
   - **Status** (Checkbox) - Click "+", choose "Checkbox"
   - **Created** (Created time) - Click "+", choose "Created time"
   - **Priority** (Select - optional) - Click "+", choose "Select", add options: Low, Medium, High

#### 2. Create Notion Integration

1. Go to https://www.notion.so/my-integrations
2. Click **"+ New integration"**
3. Give it a name: "Desktop Reminder Bot"
4. Select your workspace
5. Under "Capabilities", ensure "Read content" is checked
6. Click **Submit**
7. **COPY THE TOKEN** - you'll need this! It looks like: `secret_abc123...`

#### 3. Connect Database to Integration

1. Open your "Desktop Reminders" database
2. Click the **"..."** menu in the top right
3. Scroll down to "Connections"
4. Click **"+ Add connections"**
5. Find and select "Desktop Reminder Bot"

#### 4. Get Database ID

1. Open your database as a full page (click the title)
2. Look at the URL in your browser:
   ```
   https://www.notion.so/xxxxxxxxxxxxxxxxxxxxxxxxxxxxx?v=yyy...
   ```
3. Copy the `xxxxxxxxxxxxxxxxxxxxxxxxxxxxx` part (32 characters, mix of letters and numbers)
4. This is your **DATABASE_ID**

### Part 2: Linux Setup

#### 1. Download the Files

Save these files to a folder on your Arch machine:
- `main.go` - Main Go source code
- `go.mod` - Go module file
- `setup-go.sh` - Installation script
- `README-GO.md` - This file

#### 2. Run the Setup Script

```bash
cd /path/to/downloaded/files
chmod +x setup-go.sh
./setup-go.sh
```

This will:
- Check Go installation
- Install `libnotify` if needed
- Build the Go binary
- Install the binary to `~/.local/bin/notion-reminder`
- Create config directory at `~/.config/notion-reminder/`
- Create a systemd service for auto-start

#### 3. Configure Your Credentials

Edit the config file:

```bash
nano ~/.config/notion-reminder/config.conf
```

Replace the placeholder values:

```conf
NOTION_TOKEN=secret_your_actual_token_here
DATABASE_ID=your_actual_32_character_database_id
```

Save and exit (Ctrl+X, then Y, then Enter)

#### 4. Test It!

Add a test reminder in Notion (from your phone or computer), then run:

```bash
notion-reminder
```

You should see a desktop notification! ✨

## Usage

### Adding Reminders

**From iPhone:**
1. Open Notion app
2. Navigate to your "Desktop Reminders" database
3. Tap the "+ New" button
4. Type your reminder
5. That's it! Leave "Status" unchecked

**From Desktop:**
Same process, but in the Notion app or web browser.

### Checking Reminders

**Manual check:**
```bash
notion-reminder
```

**Automatic on login:**
```bash
systemctl --user enable notion-reminder.service
systemctl --user start notion-reminder.service
```

Now it will check every time you log in!

### Completing Reminders

When you've done a task:
1. Open Notion (phone or computer)
2. Check the "Status" checkbox next to the reminder
3. Done! It won't appear in notifications anymore

## Customization

### Run on a Schedule

To check reminders every hour:

```bash
nano ~/.config/systemd/user/notion-reminder.timer
```

Add this content:

```ini
[Unit]
Description=Notion Reminder Timer
Requires=notion-reminder.service

[Timer]
OnBootSec=5min
OnUnitActiveSec=1h
Unit=notion-reminder.service

[Install]
WantedBy=timers.target
```

Enable it:

```bash
systemctl --user enable notion-reminder.timer
systemctl --user start notion-reminder.timer
```

### Rebuilding After Changes

If you modify the Go code:

```bash
cd /path/to/source
go build -o notion-reminder main.go
mv notion-reminder ~/.local/bin/
```

### Change Notification Limit

Edit `main.go` and change this line:

```go
maxNotifications := 5  // Change 5 to your preferred number
```

Then rebuild.

## Advantages of Go Version

Compared to the Python version:

1. **No dependencies** - Python version needs `notion-client` package
2. **Single binary** - No pip, no virtual environments
3. **Faster startup** - Go binary starts in milliseconds
4. **Smaller footprint** - One ~2MB file vs Python + packages
5. **Easy updates** - Just replace one binary file
6. **Cross-compilation** - Can build for other systems easily

## Building for Other Systems

You can compile for different systems from your Arch machine:

```bash
# For another Linux system
GOOS=linux GOARCH=amd64 go build -o notion-reminder-linux main.go

# For Raspberry Pi
GOOS=linux GOARCH=arm64 go build -o notion-reminder-pi main.go

# For macOS (if you have a Mac too)
GOOS=darwin GOARCH=amd64 go build -o notion-reminder-mac main.go
```

## Troubleshooting

### "Config file not found"
Run the setup script again: `./setup-go.sh`

### "notify-send not found"
Install libnotify: `sudo pacman -S libnotify`

### "Error fetching reminders"
- Check your NOTION_TOKEN is correct
- Check your DATABASE_ID is correct
- Make sure you connected the integration to your database
- Verify you have internet connection

### "Go is not installed"
Install Go: `sudo pacman -S go`

### Binary not in PATH
Add to your `~/.bashrc` or `~/.zshrc`:
```bash
export PATH="$HOME/.local/bin:$PATH"
```

Then reload: `source ~/.bashrc`

### No notifications appearing
1. Check if the program runs without errors: `notion-reminder`
2. Test notify-send: `notify-send "Test" "Hello"`
3. Make sure you're in a graphical session (not SSH)

## Performance

Typical metrics on modern hardware:
- Binary size: ~2-3MB
- Memory usage: ~10-15MB when running
- Startup time: <10ms
- API request time: ~200-500ms (depends on internet)
- Total execution time: Usually <1 second

## File Locations

- Binary: `~/.local/bin/notion-reminder`
- Config: `~/.config/notion-reminder/config.conf`
- Service: `~/.config/systemd/user/notion-reminder.service`
- Source: Keep wherever you downloaded it

## How It Works

1. Reads config from `~/.config/notion-reminder/config.conf`
2. Makes HTTP POST request to Notion API
3. Queries database for unchecked Status items
4. Parses JSON response
5. Formats reminder data
6. Calls `notify-send` for each reminder (max 5)
7. Exits (no daemon needed)

## Privacy & Security

- Your Notion token is stored locally in `~/.config/notion-reminder/config.conf`
- Keep this file secure (readable only by your user by default)
- The program only reads from Notion, it doesn't modify anything
- No data is sent anywhere except to Notion's API
- All code is in `main.go` - you can review exactly what it does

## Code Structure

The Go program is organized into these functions:

- `main()` - Entry point, orchestrates everything
- `loadConfig()` - Reads config file
- `getIncompleteReminders()` - Queries Notion API
- `formatReminder()` - Extracts data from Notion pages
- `showNotification()` - Displays desktop notifications

Total: ~280 lines of clear, readable Go code with no external dependencies.

## License

Free to use and modify as you wish!

## Development

Want to contribute or modify?

```bash
# Format code
go fmt main.go

# Check for issues
go vet main.go

# Build with optimizations
go build -ldflags="-s -w" -o notion-reminder main.go

# Run without installing
go run main.go
```

## Questions?

The code is straightforward Go - feel free to modify it to your needs!

# D&D Shopkeeper Bot

A Discord bot quartermaster for D&D one-shot campaigns. Players can browse and purchase items between sessions via Discord without taking table time. Powered by a local Ollama LLM for immersive, in-character interactions.

## Features

- **Slash Commands**: `/shop`, `/buy`, `/inventory`, `/history`
- **AI-Powered Shopkeeper**: Grash Ironledger, a sassy half-orc quartermaster with attitude
- **Character-Aware**: Knows player backstories for thematic item recommendations
- **Monthly Rotation**: Seed-based uncommon item rotation (same month = same items)
- **Purchase Logging**: Tracks all purchases for GM review at session start
- **Local-Only**: Uses Ollama, no external API calls

## Architecture

```
dnd-shopkeep/
├── main.go                    # Entry point, loads .env, starts bot
├── bot/
│   ├── bot.go                 # Discord session, system prompt, event handlers
│   ├── commands.go            # Slash command definitions
│   ├── handlers.go            # /shop, /buy, /inventory, /history handlers
│   └── messaging.go           # Legacy chat support, username mapping
├── ai/
│   └── ai.go                  # Ollama API client, conversation history
├── shop/
│   ├── catalog.go             # Load/query catalog, fuzzy item search
│   ├── character.go           # Character profile loading, user mapping
│   ├── history.go             # Purchase history read/append
│   └── rotation.go            # Monthly uncommon item rotation algorithm
├── config/
│   └── config.go              # Configuration constants
└── data/
    ├── catalog.json           # Shop inventory (30 PHB items)
    ├── characters/            # Character profile JSONs
    │   ├── tim_paladin.json
    │   ├── eric_wizard.json
    │   ├── dieter_rogue.json
    │   └── guest_fighter.json
    └── history/               # Purchase log JSONs
        ├── tim_paladin.json
        ├── eric_wizard.json
        ├── dieter_rogue.json
        └── guest_fighter.json
```

## Data Flow

```
Discord slash command
    ↓
bot/handlers.go (command handler)
    ↓
shop/ package (load catalog, character, history)
    ↓
ai/ai.go (send to Ollama for flavor text)
    ↓
Response posted to Discord
```

## Requirements

- Go 1.21+
- [Ollama](https://ollama.ai/) running locally
- Llama 3.1 8B model
- Discord bot token with slash command permissions

## Setup

### 1. Install Ollama and pull the model

```bash
# Install Ollama (see https://ollama.ai/)
curl -fsSL https://ollama.ai/install.sh | sh

# Pull the model
ollama pull llama3.1:8b
```

### 2. Create Discord Bot

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Create a new application
3. Go to Bot → Add Bot
4. Copy the bot token
5. Go to OAuth2 → URL Generator
   - Scopes: `bot`, `applications.commands`
   - Bot Permissions: `Send Messages`, `Use Slash Commands`
6. Use generated URL to invite bot to your server

### 3. Configure Environment

Create a `.env` file in the project root:

```bash
BOT_TOKEN=your_discord_bot_token_here
GUILD_ID=your_guild_id_here  # Optional: for faster command registration during dev
```

### 4. Build and Run

```bash
# Build
go build -o dnd-shopkeep

# Start Ollama (in separate terminal)
ollama serve

# Run the bot
./dnd-shopkeep
```

## Commands

### `/shop [category]`

Browse the quartermaster's wares.

**Categories:**
- `all` - All items (default)
- `weapons` - Swords, bows, etc.
- `armor` - Armor and shields
- `potions` - Healing potions and consumables
- `gear` - Adventuring equipment
- `wondrous` - Magical items
- `monthly` - This month's special rotation

### `/buy <item> [quantity]`

Purchase an item from the shop.

- Fuzzy matches item names (e.g., "longsword" finds "Longsword")
- Logs purchase to character's history
- Reminds player to deduct gold from character sheet

### `/inventory`

View your character's current inventory plus any pending purchases.

### `/history`

View your complete purchase history with dates and totals.

## Configuration

### Adding New Players

Edit `shop/character.go` to add the Discord username → character file mapping:

```go
var UserCharacterMap = map[string]string{
    "discord_username": "character_filename",
    // ...
}
```

Then create the character profile in `data/characters/character_filename.json`:

```json
{
  "name": "Character Display Name",
  "class_level": "Fighter 5",
  "current_inventory": ["Longsword", "Shield"],
  "backstory_summary": "Brief backstory for AI recommendations..."
}
```

And initialize their history in `data/history/character_filename.json`:

```json
{
  "character": "character_filename",
  "purchases": []
}
```

### Adding Items to Catalog

Edit `data/catalog.json` to add items:

```json
{
  "name": "Item Name",
  "category": "weapons",
  "price": 100,
  "description": "Item description",
  "rarity": "common"
}
```

### Monthly Rotation

The rotation algorithm in `shop/rotation.go` uses the month string (e.g., "2026-01") as a seed, so the same month always produces the same 5 uncommon items. Edit `UncommonItems` in that file to change the rotation pool.

## The Shopkeeper

**Grash Ironledger** is a grizzled female half-orc quartermaster with:
- Gray-streaked dark hair, prominent tusks, weathered green skin with old scars
- A desk buried in requisition forms and inventory ledgers
- A perpetually raised eyebrow and zero tolerance for excuses
- Dry, cutting wit and heavy sighs

She responds to mentions of "grash" or "quartermaster" in regular chat, in addition to slash commands.

## Workflow

### Between Sessions
1. Player uses `/shop` to browse items
2. Player uses `/buy` to purchase (logged to history JSON)
3. Player manually deducts gold from their character sheet

### At Session Start
1. GM checks `data/history/*.json` files
2. Confirms purchases were added to character sheets
3. Clears history files if desired

## Development

### Testing AI Standalone

```bash
go run ai/ai.go
```

This launches an interactive CLI to test the Ollama conversation engine.

### Guild vs Global Commands

- With `GUILD_ID` set: Commands register instantly but only work in that server
- Without `GUILD_ID`: Commands register globally (can take up to 1 hour to propagate)

For development, use `GUILD_ID` for instant updates.

## License

MIT

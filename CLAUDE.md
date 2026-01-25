# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

### Discord Shop Bot

#### Goal
Discord bot quartermaster for our monthly D&D one-shots. Players browse/purchase items between sessions via Discord without taking table time.

#### Simplified Architecture
```
Mission ends → Claude generates catalog JSON
    ↓
Players interact with Discord bot (Ollama 8b)
    ↓
Bot logs purchases to character's history
    ↓
Next session: Review purchase histories
```

#### Database Structure

##### Character Profile (one JSON per character)
```json
{
  "name": "Jack 'Hammer'",
  "class_level": "Paladin 5",
  "current_inventory": [
    "Cloak of Protection",
    "Longsword",
    "Plate Armor"
  ],
  "backstory_summary": "Sworn vengeance against necromancer who killed family. Former city guard."
}
```

##### Purchase History (one JSON per character)
```json
{
  "character": "Jack 'Hammer'",
  "purchases": [
    {
      "date": "2026-01-24",
      "item": "+1 Longsword",
      "price": 500,
      "session": "Between Wild Sheep Chase and Wolves of Welton"
    }
  ]
}
```

##### Catalog (single shared JSON)
```json
{
  "common_items": [...],
  "monthly_rotation": {
    "month": "2026-01",
    "uncommon_items": [...]
  }
}
```

#### How It Works

##### Between Sessions:
1. Player: `/shop magic items`
2. Bot reads catalog, responds as quartermaster
3. Player: `/buy +1 longsword`
4. Bot logs purchase to character's history JSON
5. **Player manually deducts gold from character sheet**

##### At Session Start:
GM checks purchase history JSONs, confirms items added to character sheets

#### What I Need for Claude Code Session

1. **Sample catalog JSON** (20-30 items, PHB pricing)
2. **Character profile JSON templates** (for party members)
3. **Purchase history JSON schema**
4. **Ollama system prompt** (quartermaster personality + backstory-aware recommendations)
5. **Monthly rotation algorithm** (seed-based uncommon item selection)
6. **Discord bot commands** - `/shop`, `/buy`, `/inventory`, `/history`

#### Key Features
- Bot uses character backstory to offer thematic item suggestions
- No gold tracking (manual on sheets)
- Simple append-only purchase logs
- Local-only (Ollama, no internet)
- Matches D&D 5e PHB pricing

**Total docs needed: ~6-8 JSONs (4 characters + catalog + 4 purchase histories)**

## Build and Run

```bash
# Build
go build -o dnd-shopkeep

# Run (requires .env with BOT_TOKEN and Ollama running locally)
./dnd-shopkeep

# Test AI conversation engine standalone (without Discord)
go run ai/ai.go
```

## External Requirements

- Discord bot token in `.env` file as `BOT_TOKEN`
- Ollama running on `localhost:11434`
- Llama 3.1 8B model: `ollama pull llama3.1:8b`

## Architecture

```
main.go          → Entry point, loads .env, starts bot
bot/bot.go       → Discord session, event handlers, system prompt (bot personality)
bot/messaging.go → Message processing, username mapping, response posting
ai/ai.go         → Ollama API client, conversation history management
config/config.go → Configuration (placeholder)
```

**Data Flow:** Discord message → `newMessage()` handler → augment with username → `SendToOllama()` → post response to channel

**Conversation Context:** The bot maintains full conversation history in memory via the `Conversation` struct in `ai/ai.go`, sending the complete context to Ollama on each request.

**Username Mapping:** `augmentMessageWithUsername()` in `messaging.go` maps Discord usernames to display names (hardcoded for the 3 community members).

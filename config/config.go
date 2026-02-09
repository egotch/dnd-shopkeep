package config

// DataPaths contains paths to data directories
var DataPaths = struct {
	Weapons         string
	Armor           string
	Potions         string
	AdventuringGear string
	SessionSpecials string
	Characters      string
	History         string
	MagicWeapons    string
	MagicArmor      string
	MagicPotions    string
	WondrousItems   string
}{
	Weapons:         "data/weapons.json",
	Armor:           "data/armor.json",
	Potions:         "data/potions.json",
	AdventuringGear: "data/adventuring_gear.json",
	SessionSpecials: "data/session_specials.json",
	Characters:      "data/characters",
	History:         "data/history",
	MagicWeapons:    "data/magic_weapons.json",
	MagicArmor:      "data/magic_armor.json",
	MagicPotions:    "data/magic_potions.json",
	WondrousItems:   "data/wondrous_items.json",
}

// ShopkeeperName is the name of the quartermaster NPC
var ShopkeeperName = "Grash Ironledger"

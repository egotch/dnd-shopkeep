package config

// DataPaths contains paths to data directories
var DataPaths = struct {
	Catalog    string
	Characters string
	History    string
}{
	Catalog:    "data/catalog.json",
	Characters: "data/characters",
	History:    "data/history",
}

// UserCharacterMap maps Discord usernames to character file names (without .json)
// This is the central mapping used by the shop package
var UserCharacterMap = map[string]string{
	"timmehhey": "tim_paladin",
	"egotch":    "eric_wizard",
	"dhrudolp":  "dieter_rogue",
}

// ShopkeeperName is the name of the quartermaster NPC
var ShopkeeperName = "Grash Ironledger"

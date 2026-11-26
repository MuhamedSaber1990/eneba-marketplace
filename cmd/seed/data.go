package main

// The invented catalog. Every title, seller and description here is made up
// for this project.

type game struct {
	title       string
	genre       string
	released    string // YYYY-MM-DD
	basePrice   int    // cents; sellers price a little above or below this
	description string
}

var games = []game{
	{"Ashen Crown", "Action RPG", "2024-03-14", 4999, "Reclaim a burned kingdom one fortress at a time. Fast melee combat, branching skill trees and a world that remembers which towns you saved."},
	{"Starfall Odyssey", "Adventure", "2023-09-01", 5999, "Chart a broken star map in a ship you build yourself. Land on hand-made planets, trade with drifting stations and follow a signal nobody else can hear."},
	{"Neon Drift 2089", "Racing", "2022-06-21", 2999, "Arcade street racing through a city that never switches off. Tune your car, chain drifts for boost and outrun the corporate patrols."},
	{"Hollow Tide", "Horror", "2024-10-18", 3999, "A lighthouse keeper, a missing crew and a sea that rises a little higher every night. First-person survival horror with almost no weapons."},
	{"Ironvale Tactics", "Strategy", "2021-11-05", 3499, "Turn-based battles for a mountain republic. Position matters, weather matters, and every soldier you lose stays lost."},
	{"Pocket Kingdoms", "Strategy", "2023-02-09", 1999, "Build a kingdom that fits on a hilltop. A relaxed city builder with short sessions and neighbours who visit, trade and occasionally invade."},
	{"Skybound Couriers", "Simulation", "2022-04-12", 2499, "Fly parcels between floating islands in a patched-up seaplane. Plan routes around storms and upgrade your hangar with what you earn."},
	{"Emberwatch", "Shooter", "2025-01-23", 5999, "Four-player co-op shooter set in a city under permanent wildfire. Hold the line, rescue civilians and get out before the wind turns."},
	{"Lanternwood", "Simulation", "2023-05-30", 2999, "Inherit an overgrown orchard at the edge of an old forest. Farm, fish, befriend the villagers and find out why the lanterns light themselves."},
	{"Circuit Breakers", "Platformer", "2021-08-19", 1499, "A precise 2D platformer about two robots sharing one battery. Swap power mid-jump to get through two hundred hand-built rooms."},
	{"Frostline Rally", "Racing", "2024-12-02", 3999, "Point-to-point rally across frozen passes and logging roads. Realistic handling, a full co-driver and stages that change with the weather."},
	{"Dune Runners", "Adventure", "2022-10-27", 4499, "Cross an open desert on a sand skiff you repair with whatever you find. Convoys, sandstorms and ruins that only appear at dusk."},
	{"Mythforge Legends", "Action RPG", "2025-06-05", 5499, "Forge weapons from the myths you defeat. Co-op dungeon crawling with loot that carries the abilities of the monster it came from."},
	{"Quiet Harbor", "Puzzle", "2023-11-16", 1799, "A calm puzzle adventure about restoring a fishing town. Fix boats, redirect canals and listen to the stories people tell while you work."},
	{"Bladeborn Arena", "Fighting", "2024-07-11", 3999, "One-on-one weapon fighting with a parry system that rewards patience. Sixteen fighters, full rollback netcode and a deep training mode."},
	{"Verdant Protocol", "Shooter", "2022-02-24", 4499, "A sci-fi shooter on a terraforming station where the plants have started winning. Fight through greenhouses the size of cities."},
	{"Tiny Titans League", "Sports", "2023-07-06", 2499, "Five-a-side football with oversized heads and simple controls. Local multiplayer for four and a season mode with transfers."},
	{"Obsidian Vow", "Stealth", "2025-03-27", 4999, "Stealth action in a city built inside a dormant volcano. Every contract can be finished without being seen, or very loudly."},
	{"Gearheart Garage", "Simulation", "2021-05-13", 1999, "Buy wrecks, strip them down and rebuild them bolt by bolt. A mechanic simulator with a used-car market that reacts to what you sell."},
	{"Witchlight Academy", "Action RPG", "2024-09-05", 4499, "Enroll at a school for apprentice witches. Brew, duel and explore a castle whose corridors rearrange themselves every term."},
	{"Rogue Meridian", "Roguelike", "2022-12-08", 1999, "A roguelike deck-builder aboard a train that never stops. Draft cards between stations and decide which carriage to defend."},
	{"Saltwind Pirates", "Adventure", "2023-04-20", 3499, "Sail a hand-drawn archipelago with a crew you recruit island by island. Ship battles, buried maps and a rival who is always one port ahead."},
	{"Paper Planets", "Puzzle", "2021-09-30", 999, "Fold tiny paper worlds into shape. A short, clever puzzle game where each planet unfolds into the next."},
	{"Thunder Circuit Kart", "Racing", "2025-08-14", 3999, "Kart racing with weather you can throw at your friends. Twenty-four tracks, split-screen for four and online cups."},
}

type seller struct {
	name   string
	rating float64
	sales  int
}

var sellers = []seller{
	{"KeyHarbor", 4.9, 184210},
	{"PixelVault", 4.8, 96540},
	{"NorthGate Games", 4.7, 71300},
	{"LootLantern", 4.6, 45870},
	{"GG Depot", 4.5, 38120},
	{"Arcadia Keys", 4.4, 22950},
	{"ByteBazaar", 4.2, 12480},
	{"CodeCrate", 3.9, 5310},
}

// An edition is one platform and region a game can be sold for.
type edition struct {
	platform string
	region   string
}

var editions = []edition{
	{"Steam", "Global"},
	{"Steam", "Europe"},
	{"PlayStation", "Europe"},
	{"PlayStation", "United States"},
	{"Xbox", "Global"},
	{"Xbox", "United States"},
	{"Nintendo Switch", "Europe"},
	{"Nintendo Switch", "United States"},
}

// Pairs of colors for the generated covers.
var coverColors = [][2]string{
	{"#ff5f6d", "#ffc371"},
	{"#7f00ff", "#e100ff"},
	{"#00c6ff", "#0072ff"},
	{"#11998e", "#38ef7d"},
	{"#fc466b", "#3f5efb"},
	{"#f7971e", "#ffd200"},
	{"#1d2b64", "#f8cdda"},
	{"#0f2027", "#2c5364"},
	{"#c31432", "#240b36"},
	{"#134e5e", "#71b280"},
	{"#8e2de2", "#4a00e0"},
	{"#ee0979", "#ff6a00"},
}

package main

import (
	"math/rand"
	"strings"
	"time"
)

var Usernames = []string{
	"EpicGamer42",
	"PixelMaster",
	"StreamNinja",
	"LevelUpLuna",
	"ShadowPlay",
	"QuantumQuest",
	"NoobSlayer",
	"MemeStream",
	"ChaosController",
	"NeonNinja",
	"GameWizard",
	"ElectricEcho",
	"WanderingPixel",
	"ChillStream",
	"ArcadeAlpha",
	"CyberSurfer",
	"LootLegend",
	"VirtualVoyager",
	"LaserFocus",
	"EchoReplay",
	"PrismPulse",
}

var Slugs = []string{
	"R4ver",
	"ACPixel",
	"Lorylie",
	"iGypc",
	"LucidSilver",
	"Lyrrah",
	"Moofypoof",
	"Kizime",
	"Kevlantis",
	"Tiru",
}

var FakeMessages = []string{
	// Game-specific comments
	"That boss fight was insane!",
	"Wow, didn't see that plot twist coming!",
	"How did you manage that jump?!",
	"This game's graphics are next level.",
	"Been waiting to see this level for weeks!",

	// User interactions
	"Can't argue with that, {username}. Great point!",
	"@{username} Nice play! How long have you been streaming this game?",
	"Lol, {username} is totally carrying the team right now.",
	"Yo {username}, what's your best strategy here?",
	"Someone clip that moment, {username}!",

	// Supportive chat messages
	"First time watching - this stream is awesome!",
	"Sub goal getting closer! Let's go!",
	"Chat, we're the best community ever!",
	"Anyone else loving this game?",
	"Seriously can't believe what just happened!",

	// Gaming banter
	"RNG gods are not on your side today!",
	"That was 200 IQ right there!",
	"Biggest fail of the century, lmao",
	"Clutch or kick, am I right?",
	"Chat, should we call that a pro gamer move?",
}

type BadgeInfo struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Badge types and their possible text variations
var badgeTypes = []struct {
	Type string `json:"type"`
	Text string `json:"text"`
}{
	// Subscriber badges
	{Type: "subscriber", Text: "1 Month"},

	// VIP badges
	{Type: "vip", Text: "VIP"},

	// OG badges
	{Type: "og", Text: "OG"},

	// Occasional empty badge (no badge)
	{Type: "", Text: ""},
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// GetRandomUsername returns a random username from the list
func GetRandomUsername() string {
	return Usernames[rng.Intn(len(Usernames))]
}

func GetRandomSlug() string {
	return Slugs[rng.Intn(len(Slugs))]
}

// GetRandomMessage returns a random message from the list
// This version replaces {username} placeholders with actual usernames
func GetRandomMessage() string {
	message := FakeMessages[rng.Intn(len(FakeMessages))]

	// If the message contains {username}, replace it with a random username
	if strings.Contains(message, "{username}") {
		randomUsername := GetRandomSlug()
		message = strings.Replace(message, "{username}", randomUsername, -1)
	}

	return message
}

// GetRandomBadges generates a slice of badges
// By default, it can return 0-2 badges
func GetRandomBadges() []struct {
	Type string `json:"type"`
	Text string `json:"text"`
} {
	// Randomly decide how many badges to give (0-2)
	numBadges := rng.Intn(3)

	// If no badges, return empty slice
	if numBadges == 0 {
		return []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{}
	}

	// Create a copy of badge types to avoid modifying the original
	availableBadges := make([]struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}, len(badgeTypes))
	copy(availableBadges, badgeTypes)

	// Shuffle the badges
	rng.Shuffle(len(availableBadges), func(i, j int) {
		availableBadges[i], availableBadges[j] = availableBadges[j], availableBadges[i]
	})

	// Return the first numBadges, filtering out empty badges
	var selectedBadges []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	for _, badge := range availableBadges {
		if badge.Type != "" {

			if len(selectedBadges) > 1 && badge.Type != selectedBadges[len(selectedBadges)-1].Type {
				break
			}

			selectedBadges = append(selectedBadges, badge)
			if len(selectedBadges) == numBadges {
				break
			}
		}
	}

	return selectedBadges
}

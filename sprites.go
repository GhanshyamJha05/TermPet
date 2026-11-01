package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

type Direction int

const (
	DirLeft Direction = iota
	DirRight
)

type StateType string

const (
	StateIdle     StateType = "idle"
	StateWalking  StateType = "walking"
	StateSleeping StateType = "sleeping"
	StateHappy    StateType = "happy"
	StateHungry   StateType = "hungry"
	StateCurious  StateType = "curious"
	StateAngry    StateType = "angry"
	StateVictory  StateType = "victory"
)

type SpriteFrames struct {
	Left  []string `json:"left"`
	Right []string `json:"right"`
}

type SpeciesSprites struct {
	Idle     SpriteFrames `json:"idle"`
	Walking  SpriteFrames `json:"walking"`
	Sleeping SpriteFrames `json:"sleeping"`
	Happy    SpriteFrames `json:"happy"`
	Hungry   SpriteFrames `json:"hungry"`
	Curious  SpriteFrames `json:"curious"`
	Angry    SpriteFrames `json:"angry"`
	Victory  SpriteFrames `json:"victory"`
}

type SpeciesDefinition struct {
	Name    string         `json:"name"`
	Normal  SpeciesSprites `json:"normal"`
	Compact SpeciesSprites `json:"compact"`
	// Evolutionary stages: optional custom sprites for higher stages
	Juvenile *SpeciesSprites `json:"juvenile,omitempty"`
	Ancient  *SpeciesSprites `json:"ancient,omitempty"`
}

// ColorTheme holds Lip Gloss color styling for pet and UI elements
type ColorTheme struct {
	Name         string
	PetPrimary   string
	PetSecondary string
	PetAccent    string
	StatusBarBg  string
	PromptColor  string
	MoodColor    string
}

// Built-in color themes
var Themes = map[string]ColorTheme{
	"dracula": {
		Name:         "dracula",
		PetPrimary:   "#BD93F9", // purple
		PetSecondary: "#FFB86C", // orange
		PetAccent:    "#FF79C6", // pink
		StatusBarBg:  "#282A36",
		PromptColor:  "#50FA7B",
		MoodColor:    "#FF79C6",
	},
	"fire": {
		Name:         "fire",
		PetPrimary:   "#FF5555", // red
		PetSecondary: "#FFB86C", // amber
		PetAccent:    "#F1FA8C", // yellow
		StatusBarBg:  "#3A1515",
		PromptColor:  "#FFB86C",
		MoodColor:    "#FF5555",
	},
	"emerald": {
		Name:         "emerald",
		PetPrimary:   "#50FA7B", // green
		PetSecondary: "#8BE9FD", // cyan
		PetAccent:    "#F1FA8C", // lime
		StatusBarBg:  "#1A2E22",
		PromptColor:  "#50FA7B",
		MoodColor:    "#50FA7B",
	},
	"cyber": {
		Name:         "cyber",
		PetPrimary:   "#00FFFF", // electric cyan
		PetSecondary: "#FF007F", // neon pink
		PetAccent:    "#FFE600", // neon yellow
		StatusBarBg:  "#0A0F1F",
		PromptColor:  "#00FFFF",
		MoodColor:    "#FF007F",
	},
	"sakura": {
		Name:         "sakura",
		PetPrimary:   "#FFB7C5", // cherry blossom
		PetSecondary: "#FF69B4", // hot pink
		PetAccent:    "#FFFFFF", // white
		StatusBarBg:  "#2E1F27",
		PromptColor:  "#FFB7C5",
		MoodColor:    "#FF69B4",
	},
}

var SpeciesMap = map[string]SpeciesDefinition{
	"dragon": {
		Name: "dragon",
		Normal: SpeciesSprites{
			Idle: SpriteFrames{
				Left: []string{
					"    /\\_/\\     \n   ( o.o ) <3 \n   /| W |\\  ^  \n  /  -  \\/    \n =\"==\"=\"=     ",
					"    /\\_/\\     \n   ( -.- )    \n   /| W |\\  ^  \n  /  -  \\/    \n =\"==\"=\"=     ",
				},
				Right: []string{
					"     /\\_/\\    \n 3> ( o.o )   \n  ^  /| W |\\   \n   \\/  -  \\   \n     =\"==\"=\"= ",
					"     /\\_/\\    \n    ( -.- )   \n  ^  /| W |\\   \n   \\/  -  \\   \n     =\"==\"=\"= ",
				},
			},
			Walking: SpriteFrames{
				Left: []string{
					"    /\\_/\\ ~   \n   ( ^.^ )/   \n  ~/| W |\\     \n  /  /  \\/    \n =\"==\"        ",
					"    /\\_/\\     \n   ( ^.^ )\\ ~ \n   /| W |\\~    \n  /  \\  \\/    \n     =\"==\"    ",
				},
				Right: []string{
					"   ~ /\\_/\\    \n    \\( ^.^ )  \n     /| W |\\~  \n   \\/  \\  \\   \n        =\"==\" ",
					"     /\\_/\\    \n   ~ /( ^.^ ) \n    ~/| W |\\   \n   \\/  /  \\   \n    =\"==\"     ",
				},
			},
			Sleeping: SpriteFrames{
				Left: []string{
					"    /\\_/\\   z \n   ( -.- ) z  \n  ~/|   |\\~   \n  (___m_m)    ",
					"    /\\_/\\   Z \n   ( -.- )  z \n  ~/|   |\\~   \n  (___m_m)    ",
				},
				Right: []string{
					" z   /\\_/\\    \n  z ( -.- )   \n   ~/|   |\\~  \n   (m_m___)   ",
					" Z   /\\_/\\    \n  z  ( -.- )  \n   ~/|   |\\~  \n   (m_m___)   ",
				},
			},
			Happy: SpriteFrames{
				Left: []string{
					"    /\\_/\\  *! \n   ( >.< )/   \n ~(<3|W|<3)~  \n  /  -  \\/    \n =\"==\"=\"=     ",
					"  * /\\_/\\ !*  \n   ( ^o^ )/   \n ~(<3|W|<3)~  \n  /  -  \\/    \n =\"==\"=\"=     ",
				},
				Right: []string{
					" !* /\\_/\\    \n    \\( >.< )  \n   ~(<3|W|<3)~ \n   \\/  -  \\   \n     =\"==\"=\"= ",
					" *! /\\_/\\ *  \n    \\( ^o^ )  \n   ~(<3|W|<3)~ \n   \\/  -  \\   \n     =\"==\"=\"= ",
				},
			},
			Hungry: SpriteFrames{
				Left: []string{
					"    /\\_/\\   feed me!\n   ( ;o; ) / (fire)\n   /| W |\\ /   \n  /  -  \\/    \n =\"==\"=\"=     ",
					"    /\\_/\\  feed me!\n   ( ;.; )    \n   /| W |\\     \n  /  -  \\/    \n =\"==\"=\"=     ",
				},
				Right: []string{
					"feed me!  /\\_/\\   \n (fire)\\ ( ;o; )  \n        \\ /| W |\\ \n        \\/  -  \\  \n          =\"==\"=\"=",
					" feed me!  /\\_/\\   \n          ( ;.; )  \n           /| W |\\ \n          \\/  -  \\  \n            =\"==\"=\"=",
				},
			},
			Curious: SpriteFrames{
				Left: []string{
					"    /\\_/\\  ?  \n   ( @.@ )/   \n  ~/| W |\\    \n  /  -  \\/    \n =\"==\"=\"=     ",
					"    /\\_/\\  ?? \n   ( O.O )/   \n  ~/| W |\\    \n  /  -  \\/    \n =\"==\"=\"=     ",
				},
				Right: []string{
					"  ? /\\_/\\     \n    \\( @.@ )  \n    /| W |\\~  \n   \\/  -  \\   \n     =\"==\"=\"= ",
					" ?? /\\_/\\     \n    \\( O.O )  \n    /| W |\\~  \n   \\/  -  \\   \n     =\"==\"=\"= ",
				},
			},
			Angry: SpriteFrames{
				Left: []string{
					"    /\\_/\\  GRRR!\n   ( >o< ) / (fire)\n   /| W |\\/   \n  /  -  \\/    \n =\"==\"=\"=     ",
					"    /\\_/\\  GRRR!\n   ( `o` ) / (fire)\n   /| W |\\/   \n  /  -  \\/    \n =\"==\"=\"=     ",
				},
				Right: []string{
					"GRRR!  /\\_/\\  \n(fire)\\ ( >o< )\n      \\ /| W |\\ \n      \\/  -  \\  \n        =\"==\"=\"=",
					" GRRR! /\\_/\\  \n (fire)\\ ( `o` )\n       \\ /| W |\\ \n       \\/  -  \\  \n         =\"==\"=\"=",
				},
			},
			Victory: SpriteFrames{
				Left: []string{
					"  * /\\_/\\ *\\o/\n   ( ^o^ )/   \n ~(<3|W|<3)~  \n  /  -  \\/    \n =\"==\"=\"=     ",
					" \\o/*\\_/\\ *   \n   ( ^o^ )\\   \n ~(<3|W|<3)~  \n  /  -  \\/    \n =\"==\"=\"=     ",
				},
				Right: []string{
					"\\o/* /\\_/\\ *  \n    \\( ^o^ )  \n   ~(<3|W|<3)~ \n   \\/  -  \\   \n     =\"==\"=\"= ",
					"   * /\\_/*\\o/ \n    /( ^o^ )  \n   ~(<3|W|<3)~ \n   \\/  -  \\   \n     =\"==\"=\"= ",
				},
			},
		},
		Juvenile: &SpeciesSprites{
			Idle: SpriteFrames{
				Left: []string{
					"  __   /\\_^_/\\   __  \n /  \\ ( O.O ) /  \\ \n \\_//\\ /|WW|\\ /\\_//\n     ~/ - - \\~     \n    ==\"====\"==     ",
					"  __   /\\_^_/\\   __  \n /  \\ ( -.- ) /  \\ \n \\_//\\ /|WW|\\ /\\_//\n     ~/ - - \\~     \n    ==\"====\"==     ",
				},
				Right: []string{
					"  __   /\\_^_/\\   __  \n /  \\ ( O.O ) /  \\ \n \\_//\\ /|WW|\\ /\\_//\n     ~/ - - \\~     \n    ==\"====\"==     ",
					"  __   /\\_^_/\\   __  \n /  \\ ( -.- ) /  \\ \n \\_//\\ /|WW|\\ /\\_//\n     ~/ - - \\~     \n    ==\"====\"==     ",
				},
			},
			Walking: SpriteFrames{
				Left: []string{
					"  \\\\   /\\_^_/\\   //  \n   \\\\ ( ^.^ ) //   \n \\_//\\ /|WW|\\ /\\_//\n     ~/  / \\~      \n    ==\"==          ",
					"  //   /\\_^_/\\   \\\\  \n   // ( ^.^ ) \\\\   \n \\_//\\ /|WW|\\ /\\_//\n     ~/ \\  \\~      \n        ==\"==      ",
				},
				Right: []string{
					"  //   /\\_^_/\\   \\\\  \n   // ( ^.^ ) \\\\   \n \\_//\\ /|WW|\\ /\\_//\n     ~/ \\  \\~      \n        ==\"==      ",
					"  \\\\   /\\_^_/\\   //  \n   \\\\ ( ^.^ ) //   \n \\_//\\ /|WW|\\ /\\_//\n     ~/  / \\~      \n    ==\"==          ",
				},
			},
			Sleeping: SpriteFrames{
				Left: []string{
					"   /\\_^_/\\   __  z \n  ( -.- )   /  \\  Z\n ~(|WW| )   \\_//   \n  (___m_m_m)       ",
					"   /\\_^_/\\   __  Z \n  ( -.- )   /  \\ z \n ~(|WW| )   \\_//   \n  (___m_m_m)       ",
				},
				Right: []string{
					" z  __   /\\_^_/\\   \n  Z/  \\ ( -.- )    \n   \\_// (|WW|)~    \n        (m_m_m___) ",
					" Z  __   /\\_^_/\\   \n  z/  \\ ( -.- )    \n   \\_// (|WW|)~    \n        (m_m_m___) ",
				},
			},
			Happy: SpriteFrames{
				Left: []string{
					" \\\\//  /\\_^_/\\  \\\\// \n (★)  ( ^o^ )  (★) \n \\_//\\ /|WW|\\ /\\_//\n     ~/ - - \\~     \n    ==\"====\"==     ",
					" //\\\\  /\\_^_/\\  //\\\\ \n (★)  ( ^▽^ )  (★) \n \\_//\\ /|WW|\\ /\\_//\n     ~/ - - \\~     \n    ==\"====\"==     ",
				},
				Right: []string{
					" \\\\//  /\\_^_/\\  \\\\// \n (★)  ( ^o^ )  (★) \n \\_//\\ /|WW|\\ /\\_//\n     ~/ - - \\~     \n    ==\"====\"==     ",
					" //\\\\  /\\_^_/\\  //\\\\ \n (★)  ( ^▽^ )  (★) \n \\_//\\ /|WW|\\ /\\_//\n     ~/ - - \\~     \n    ==\"====\"==     ",
				},
			},
			Hungry: SpriteFrames{
				Left: []string{
					"  __   /\\_^_/\\   __  FEED ME!\n /  \\ ( ;o; ) /  \\ /(fire)  \n \\_//\\ /|WW|\\ /\\_//         \n     ~/ - - \\~              \n    ==\"====\"==              ",
				},
				Right: []string{
					"FEED ME! __   /\\_^_/\\   __  \n (fire)\\ /  \\ ( ;o; ) /  \\ \n        \\_//\\ /|WW|\\ /\\_// \n            ~/ - - \\~      \n           ==\"====\"==      ",
				},
			},
			Curious: SpriteFrames{
				Left: []string{
					"  __   /\\_^_/\\   __  ?? \n /  \\ ( @.@ ) /  \\/  \n \\_//\\ /|WW|\\ /\\_//  \n     ~/ - - \\~       \n    ==\"====\"==       ",
				},
				Right: []string{
					" ??  __   /\\_^_/\\   __  \n   \\/  \\ ( @.@ ) /  \\ \n    \\_//\\ /|WW|\\ /\\_// \n        ~/ - - \\~      \n       ==\"====\"==      ",
				},
			},
			Angry: SpriteFrames{
				Left: []string{
					"  \\\\   /\\_^_/\\   //  ROAR!\n  \\\\\\ ( >O< ) ///  /(fire)\n \\_//\\ /|WW|\\ /\\_//        \n     ~/ - - \\~             \n    ==\"====\"==             ",
				},
				Right: []string{
					"ROAR!  //   /\\_^_/\\   \\\\  \n(fire)\\/// ( >O< ) ///   \n      \\_//\\ /|WW|\\ /\\_//  \n          ~/ - - \\~       \n         ==\"====\"==       ",
				},
			},
			Victory: SpriteFrames{
				Left: []string{
					" \\\\//  /\\_^_/\\  \\\\// ★\n (\\o/)( ^▽^ )(\\o/)    \n \\_//\\ /|WW|\\ /\\_//   \n     ~/ - - \\~        \n    ==\"====\"==        ",
				},
				Right: []string{
					"★ \\\\//  /\\_^_/\\  \\\\// \n   (\\o/)( ^▽^ )(\\o/)  \n   \\_//\\ /|WW|\\ /\\_// \n       ~/ - - \\~      \n      ==\"====\"==      ",
				},
			},
		},
		Ancient: &SpeciesSprites{
			Idle: SpriteFrames{
				Left: []string{
					" <\\\\//>  /\\_★_★_/\\  <\\\\//> \n  \\\\//  (  -   -  )  \\\\//  \n  //\\\\  /||WWWW||\\  //\\\\  \n  \\//\\\\~/   ---   \\~//\\\\/ \n       ===\"=====\"===      ",
					" <\\\\//>  /\\_★_★_/\\  <\\\\//> \n  \\\\//  (  O   O  )  \\\\//  \n  //\\\\  /||WWWW||\\  //\\\\  \n  \\//\\\\~/   ---   \\~//\\\\/ \n       ===\"=====\"===      ",
				},
				Right: []string{
					" <\\\\//>  /\\_★_★_/\\  <\\\\//> \n  \\\\//  (  -   -  )  \\\\//  \n  //\\\\  /||WWWW||\\  //\\\\  \n  \\//\\\\~/   ---   \\~//\\\\/ \n       ===\"=====\"===      ",
					" <\\\\//>  /\\_★_★_/\\  <\\\\//> \n  \\\\//  (  O   O  )  \\\\//  \n  //\\\\  /||WWWW||\\  //\\\\  \n  \\//\\\\~/   ---   \\~//\\\\/ \n       ===\"=====\"===      ",
				},
			},
			Walking: SpriteFrames{
				Left: []string{
					" <<\\\\>>  /\\_★_★_/\\  <<//>> \n  \\\\//  (  ^   ^  )  \\\\//  \n  //\\\\  /||WWWW||\\  //\\\\  \n  \\//\\\\~/   / \\   \\~//\\\\/ \n       ===\"=              ",
				},
				Right: []string{
					" <<//>>  /\\_★_★_/\\  <<\\\\>> \n  \\\\//  (  ^   ^  )  \\\\//  \n  //\\\\  /||WWWW||\\  //\\\\  \n  \\//\\\\~/   / \\   \\~//\\\\/ \n             ===\"=        ",
				},
			},
			Sleeping: SpriteFrames{
				Left: []string{
					" <\\\\//>  /\\_★_★_/\\   ZZZ  \n  \\\\//  (  -   -  ) z     \n  //\\\\  /||WWWW||\\        \n       (_____m_m_m_m)     ",
				},
				Right: []string{
					"  ZZZ   /\\_★_★_/\\  <\\\\//> \n     z (  -   -  )  \\\\//  \n        /||WWWW||\\  //\\\\  \n       (m_m_m_m_____)     ",
				},
			},
			Happy: SpriteFrames{
				Left: []string{
					" ★<<>>  /\\_★_★_/\\  <<>>★  \n  \\\\//  (  ^   ^  )  \\\\//  \n  //\\\\ ~(||WWWW||)~ //\\\\  \n  \\//\\\\~/   ---   \\~//\\\\/ \n       ===\"=====\"===      ",
				},
				Right: []string{
					" ★<<>>  /\\_★_★_/\\  <<>>★  \n  \\\\//  (  ^   ^  )  \\\\//  \n  //\\\\ ~(||WWWW||)~ //\\\\  \n  \\//\\\\~/   ---   \\~//\\\\/ \n       ===\"=====\"===      ",
				},
			},
			Hungry: SpriteFrames{
				Left: []string{
					" <\\\\//>  /\\_★_★_/\\  <\\\\//> (FIRE!)\n  \\\\//  (  ;   ;  )  \\\\// /       \n  //\\\\  /||WWWW||\\  //\\\\/         \n  \\//\\\\~/   ---   \\~//\\\\/         \n       ===\"=====\"===              ",
				},
				Right: []string{
					"(FIRE!) <\\\\//>  /\\_★_★_/\\  <\\\\//> \n      \\  \\\\//  (  ;   ;  )  \\\\//  \n       \\//\\\\  /||WWWW||\\  //\\\\   \n       \\//\\\\~/   ---   \\~//\\\\/    \n            ===\"=====\"===         ",
				},
			},
			Curious: SpriteFrames{
				Left: []string{
					" <\\\\//>  /\\_★_★_/\\  <\\\\//> ??? \n  \\\\//  (  @   @  )  \\\\// /    \n  //\\\\  /||WWWW||\\  //\\\\/      \n  \\//\\\\~/   ---   \\~//\\\\/      \n       ===\"=====\"===           ",
				},
				Right: []string{
					" ???   <\\\\//>  /\\_★_★_/\\  <\\\\//> \n    \\   \\\\//  (  @   @  )  \\\\//  \n     \\  //\\\\  /||WWWW||\\  //\\\\   \n      \\//\\\\~/   ---   \\~//\\\\/    \n           ===\"=====\"===         ",
				},
			},
			Angry: SpriteFrames{
				Left: []string{
					" <\\\\//>  /\\_★_★_/\\  <\\\\//> (INFERNO!)\n  \\\\//  (  >   <  )  \\\\// /          \n  //\\\\  /||WWWW||\\  //\\\\/            \n  \\//\\\\~/   ---   \\~//\\\\/            \n       ===\"=====\"===                 ",
				},
				Right: []string{
					"(INFERNO!) <\\\\//>  /\\_★_★_/\\  <\\\\//> \n         \\  \\\\//  (  >   <  )  \\\\//  \n          \\ //\\\\  /||WWWW||\\  //\\\\   \n          \\//\\\\~/   ---   \\~//\\\\/    \n               ===\"=====\"===         ",
				},
			},
			Victory: SpriteFrames{
				Left: []string{
					"★<<\\//>> /\\_★_★_/\\ <<\\//>>★\n  \\\\//  (  ★   ★  )  \\\\//  \n  //\\\\ ~(||WWWW||)~ //\\\\  \n  \\//\\\\~/   ---   \\~//\\\\/ \n       ===\"=====\"===      ",
				},
				Right: []string{
					"★<<\\//>> /\\_★_★_/\\ <<\\//>>★\n  \\\\//  (  ★   ★  )  \\\\//  \n  //\\\\ ~(||WWWW||)~ //\\\\  \n  \\//\\\\~/   ---   \\~//\\\\/ \n       ===\"=====\"===      ",
				},
			},
		},
		Compact: SpeciesSprites{
			Idle: SpriteFrames{
				Left:  []string{"<~( o.o)>"},
				Right: []string{"<(o.o )~>"},
			},
			Walking: SpriteFrames{
				Left:  []string{"<~( ^.^)~", "<~( ^.^) "},
				Right: []string{"~(^.^ )~>", " (^.^ )~>"},
			},
			Sleeping: SpriteFrames{
				Left:  []string{"<~(-.-)zzZ"},
				Right: []string{"Zzz(-.- )~>"},
			},
			Happy: SpriteFrames{
				Left:  []string{"<~(^o^)<3"},
				Right: []string{"<3(^o^ )~>"},
			},
			Hungry: SpriteFrames{
				Left:  []string{"<~(;o;)[feed!]"},
				Right: []string{"[feed!](;o;)~>"},
			},
			Curious: SpriteFrames{
				Left:  []string{"<~(@.@)?"},
				Right: []string{"?(o.@ )~>"},
			},
			Angry: SpriteFrames{
				Left:  []string{"<~(>o<)!"},
				Right: []string{"!(>o< )~>"},
			},
			Victory: SpriteFrames{
				Left:  []string{"<~(^o^)*"},
				Right: []string{"*(^o^ )~>"},
			},
		},
	},
}

// LoadCustomSprites scans ~/.termpet/sprites/*.json and loads custom species definitions
func LoadCustomSprites() {
	dir, err := CustomSpritesDir()
	if err != nil {
		return
	}
	_ = os.MkdirAll(dir, 0755)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			filePath := filepath.Join(dir, entry.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}
			var custom SpeciesDefinition
			if err := json.Unmarshal(data, &custom); err == nil && custom.Name != "" {
				SpeciesMap[strings.ToLower(custom.Name)] = custom
			}
		}
	}
}

func GetSpriteDimensions(sprite string) (int, int) {
	lines := strings.Split(sprite, "\n")
	height := len(lines)
	maxWidth := 0
	for _, l := range lines {
		w := runewidth.StringWidth(l)
		if w > maxWidth {
			maxWidth = w
		}
	}
	return maxWidth, height
}

func GetFrame(species string, compact bool, state StateType, stage PetStage, dir Direction, tick int) string {
	def, ok := SpeciesMap[species]
	if !ok {
		def = SpeciesMap["dragon"]
	}

	sprites := def.Normal
	if compact {
		sprites = def.Compact
	} else {
		// Evolutionary stage lookup
		if stage == StageAdult && def.Ancient != nil {
			sprites = *def.Ancient
		} else if stage == StageYoung && def.Juvenile != nil {
			sprites = *def.Juvenile
		}
	}

	var sf SpriteFrames
	switch state {
	case StateWalking:
		sf = sprites.Walking
	case StateSleeping:
		sf = sprites.Sleeping
	case StateHappy:
		sf = sprites.Happy
	case StateHungry:
		sf = sprites.Hungry
	case StateCurious:
		sf = sprites.Curious
	case StateAngry:
		sf = sprites.Angry
	case StateVictory:
		sf = sprites.Victory
	case StateIdle:
		fallthrough
	default:
		sf = sprites.Idle
	}

	frames := sf.Left
	if dir == DirRight {
		frames = sf.Right
	}
	if len(frames) == 0 {
		frames = sf.Left
	}
	if len(frames) == 0 {
		return "(?_?)"
	}

	idx := (tick / 3) % len(frames)
	return frames[idx]
}

// RenderGradientSprite paints a multi-line sprite with a top-to-bottom or highlighted truecolor gradient
func RenderGradientSprite(sprite string, theme ColorTheme, state StateType) string {
	lines := strings.Split(sprite, "\n")
	if len(lines) == 0 {
		return sprite
	}

	primary := theme.PetPrimary
	secondary := theme.PetSecondary
	accent := theme.PetAccent

	// Special state overrides
	switch state {
	case StateHappy, StateVictory:
		primary = accent
	case StateAngry, StateHungry:
		primary = "#FF5555"
		secondary = "#FFB86C"
	case StateSleeping:
		primary = "#8BE9FD"
		secondary = "#6272A4"
	}

	styledLines := make([]string, len(lines))
	total := len(lines)

	for i, l := range lines {
		ratio := float64(i) / float64(max(total-1, 1))
		var color string
		if ratio < 0.35 {
			color = primary
		} else if ratio < 0.70 {
			color = secondary
		} else {
			color = accent
		}

		style := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true)
		styledLines[i] = style.Render(l)
	}

	return strings.Join(styledLines, "\n")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

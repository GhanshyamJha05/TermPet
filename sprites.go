package main

import (
	"strings"

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
)

type SpriteFrames struct {
	Left  []string
	Right []string
}

type SpeciesSprites struct {
	Idle     SpriteFrames
	Walking  SpriteFrames
	Sleeping SpriteFrames
	Happy    SpriteFrames
	Hungry   SpriteFrames
}

type SpeciesDefinition struct {
	Name    string
	Normal  SpeciesSprites
	Compact SpeciesSprites
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
		},
	},
	"cat": {
		Name: "cat",
		Normal: SpeciesSprites{
			Idle: SpriteFrames{
				Left:  []string{" /\\_/\\\n( o.o )\n > ^ < "},
				Right: []string{" /\\_/\\\n( o.o )\n > ^ < "},
			},
			Walking: SpriteFrames{
				Left: []string{
					" /\\_/\\\n(=^..^)\n  /  \\ ",
					" /\\_/\\\n(=^..^)\n  \\  / ",
				},
				Right: []string{
					" /\\_/\\\n(^..^=)\n  /  \\ ",
					" /\\_/\\\n(^..^=)\n  \\  / ",
				},
			},
			Sleeping: SpriteFrames{
				Left: []string{
					" /\\_/\\   z\n( -.- ) z\n (___)o",
					" /\\_/\\   Z\n( -.- )  z\n (___)o",
				},
				Right: []string{
					"z   /\\_/\\\n z ( -.- )\n  o(___) ",
					"Z   /\\_/\\\n z  ( -.- )\n  o(___) ",
				},
			},
			Happy: SpriteFrames{
				Left: []string{
					" /\\_/\\  *<3*\n( ^.^ )\n > ^ < ",
					" /\\_/\\   * \n( >.< ) *<3*\n > ^ < ",
				},
				Right: []string{
					"*<3* /\\_/\\\n   ( ^.^ )\n    > ^ < ",
					" *  /\\_/\\\n*<3*( >.< )\n    > ^ < ",
				},
			},
			Hungry: SpriteFrames{
				Left:  []string{" /\\_/\\  feed me!\n( ;.; )\n > ^ < "},
				Right: []string{"feed me! /\\_/\\\n        ( ;.; )\n         > ^ < "},
			},
		},
		Compact: SpeciesSprites{
			Idle: SpriteFrames{
				Left:  []string{"(=^..^)"},
				Right: []string{"(^..^=)"},
			},
			Walking: SpriteFrames{
				Left:  []string{"(=^..^)<", "(=^..^)~"},
				Right: []string{">(^..^=)", "~(^..^=)"},
			},
			Sleeping: SpriteFrames{
				Left:  []string{"(-.-)zzZ"},
				Right: []string{"Zzz(-.-)"},
			},
			Happy: SpriteFrames{
				Left:  []string{"(^o^)<3"},
				Right: []string{"<3(^o^)"},
			},
			Hungry: SpriteFrames{
				Left:  []string{"(;_;)[feed me!]"},
				Right: []string{"[feed me!](;_;)"},
			},
		},
	},
	"dog": {
		Name: "dog",
		Normal: SpeciesSprites{
			Idle: SpriteFrames{
				Left:  []string{" /^ ^\\\n/ 0 0 \\\nV\\ Y /V\n  / \\  "},
				Right: []string{" /^ ^\\\n/ 0 0 \\\nV\\ Y /V\n  / \\  "},
			},
			Walking: SpriteFrames{
				Left: []string{
					" /^ ^\\\n/ ' ' \\\nV\\ Y /V~\n  / \\  ",
					" /^ ^\\\n/ ' ' \\\nV\\ Y /V\n  \\ / ~",
				},
				Right: []string{
					" /^ ^\\\n/ ' ' \\\n~V\\ Y /V\n  / \\  ",
					" /^ ^\\\n/ ' ' \\\n V\\ Y /V\n~ \\ /  ",
				},
			},
			Sleeping: SpriteFrames{
				Left: []string{
					" /^ ^\\  z\n/ - - \\  Z\nV\\ _ /V\n (___) ",
					" /^ ^\\   Z\n/ - - \\ z\nV\\ _ /V\n (___) ",
				},
				Right: []string{
					"z  /^ ^\\\n Z/ - - \\\n  V\\ _ /V\n   (___) ",
					" Z /^ ^\\\nz / - - \\\n  V\\ _ /V\n   (___) ",
				},
			},
			Happy: SpriteFrames{
				Left: []string{
					" /^ ^\\  *!\n/ ^ ^ \\\nV\\ U /V~\n  / \\  ",
					" /^ ^\\  !*\n/ ^ ^ \\\nV\\ U /V\n  \\ / ~",
				},
				Right: []string{
					"!* /^ ^\\\n  / ^ ^ \\\n ~V\\ U /V\n   / \\  ",
					"*! /^ ^\\\n  / ^ ^ \\\n  V\\ U /V\n ~ \\ /  ",
				},
			},
			Hungry: SpriteFrames{
				Left:  []string{" /^ ^\\  feed me!\n/ o o \\\nV\\ O /V\n  / \\  "},
				Right: []string{"feed me! /^ ^\\\n        / o o \\\n        V\\ O /V\n          / \\  "},
			},
		},
		Compact: SpeciesSprites{
			Idle: SpriteFrames{
				Left:  []string{"(U o_o)"},
				Right: []string{"(o_o U)"},
			},
			Walking: SpriteFrames{
				Left:  []string{"(U o_o)~", "(U o_o)"},
				Right: []string{"~(o_o U)", "(o_o U)"},
			},
			Sleeping: SpriteFrames{
				Left:  []string{"(U -_-)zZ"},
				Right: []string{"Zz(-_- U)"},
			},
			Happy: SpriteFrames{
				Left:  []string{"(U ^_^)/<3"},
				Right: []string{"<3\\(^_^ U)"},
			},
			Hungry: SpriteFrames{
				Left:  []string{"(U ;_;)[feed me!]"},
				Right: []string{"[feed me!](;_; U)"},
			},
		},
	},
	"ghost": {
		Name: "ghost",
		Normal: SpeciesSprites{
			Idle: SpriteFrames{
				Left:  []string{" .-'\"'-.\n/ #   # \\\n|       |\n \\/\\_/\\/ "},
				Right: []string{" .-'\"'-.\n/ #   # \\\n|       |\n \\/\\_/\\/ "},
			},
			Walking: SpriteFrames{
				Left: []string{
					" .-'\"'-.\n/ *   * \\\n|  <    |\n '~\\_/~' ",
					" .-'\"'-.\n/ *   * \\\n|  <    |\n '~/_\\~' ",
				},
				Right: []string{
					" .-'\"'-.\n/ *   * \\\n|    >  |\n '~\\_/~' ",
					" .-'\"'-.\n/ *   * \\\n|    >  |\n '~/_\\~' ",
				},
			},
			Sleeping: SpriteFrames{
				Left: []string{
					" .-'\"'-.  z\n/ -   - \\  Z\n|  ~    |\n '~'~'~' ",
					" .-'\"'-.   Z\n/ -   - \\ z\n|  ~    |\n '~'~'~' ",
				},
				Right: []string{
					"z  .-'\"'-.\n Z/ -   - \\\n  |    ~  |\n   '~'~'~' ",
					" Z .-'\"'-.\nz / -   - \\\n  |    ~  |\n   '~'~'~' ",
				},
			},
			Happy: SpriteFrames{
				Left: []string{
					" .-'\"'-. *!\n/ ^   ^ \\\n|   o   |\n '~\\_/~' ",
					" .-'\"'-.  !*\n/ ^   ^ \\\n|   O   |\n '~/_\\~' ",
				},
				Right: []string{
					"!* .-'\"'-.\n  / ^   ^ \\\n  |   o   |\n   '~\\_/~' ",
					" *!.-'\"'-.\n  / ^   ^ \\\n  |   O   |\n   '~/_\\~' ",
				},
			},
			Hungry: SpriteFrames{
				Left:  []string{" .-'\"'-. feed me!\n/ o   o \\\n|   O   |\n '~\\_/~' ",
				Right: []string{"feed me! .-'\"'-.\n        / o   o \\\n        |   O   |\n         '~\\_/~' "},
			},
		},
		Compact: SpeciesSprites{
			Idle: SpriteFrames{
				Left:  []string{"<('o ')>"},
				Right: []string{"<(' o')>"},
			},
			Walking: SpriteFrames{
				Left:  []string{"<('* ')~", "<('* ') "},
				Right: []string{"~(' *')>", " (' *')>"},
			},
			Sleeping: SpriteFrames{
				Left:  []string{"<(- -)zZ>"},
				Right: []string{"<Zz(- -)>"},
			},
			Happy: SpriteFrames{
				Left:  []string{"<(^o^)*>"},
				Right: []string{"<*(^o^)>"},
			},
			Hungry: SpriteFrames{
				Left:  []string{"<(;o;)[feed!]>"},
				Right: []string{"<[feed!](;o;)>"},
			},
		},
	},
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

func GetFrame(species string, compact bool, state StateType, dir Direction, tick int) string {
	def, ok := SpeciesMap[species]
	if !ok {
		def = SpeciesMap["dragon"]
	}

	sprites := def.Normal
	if compact {
		sprites = def.Compact
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

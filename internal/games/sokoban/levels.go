package sokoban

import "fmt"

var Levels = map[string]string{
	"level01": "#####\n# @ #\n# $ #\n# . #\n#####",
	"level02": "######\n# .  #\n# $$ #\n# ..@#\n######",
}

func Load(name string) (Board, error) {
	text, ok := Levels[name]
	if !ok {
		return Board{}, fmt.Errorf("unknown level %q", name)
	}
	return Parse(text)
}

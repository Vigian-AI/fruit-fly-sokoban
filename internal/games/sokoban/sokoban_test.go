package sokoban

import (
	"sokoban-fruit-fly/internal/game"
	"testing"
)

func TestParseAndPush(t *testing.T) {
	b, err := Parse("#####\n# @ #\n# $ #\n# . #\n#####")
	if err != nil {
		t.Fatal(err)
	}
	s := b.Start()
	if !b.Legal(s, game.Down) {
		t.Fatal("expected push to be legal")
	}
	s, moved, err := b.Step(s, game.Down)
	if err != nil || !moved {
		t.Fatalf("step: %v moved=%v", err, moved)
	}
	if !b.Won(s) {
		t.Fatal("box should be on target")
	}
}

func TestCornerDeadlock(t *testing.T) {
	b, err := Parse("####\n#@ #\n#$ #\n# .#\n####")
	if err != nil {
		t.Fatal(err)
	}
	s, moved, err := b.Step(b.Start(), game.Down)
	if err != nil || !moved {
		t.Fatal(err)
	}
	if !b.Deadlocked(s) {
		t.Fatal("expected corner deadlock")
	}
}

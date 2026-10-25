package game

type Action int

const (
	Up Action = iota
	Down
	Left
	Right
)

func (a Action) String() string {
	return []string{"UP", "DOWN", "LEFT", "RIGHT"}[a]
}

func Actions() [4]Action { return [4]Action{Up, Down, Left, Right} }

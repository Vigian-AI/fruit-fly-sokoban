# Methodology

Sokoban is modeled as a deterministic episodic MDP. The state contains the player and
all box coordinates; actions are up, down, left, and right. A transition moves the
player, optionally pushing one box. A solved state has every box on a target. A
deadlocked corner or the step limit terminates an unsuccessful episode.

The retina uses four fixed channels per board cell: walls, targets, boxes, and player.
The frozen network projects those inputs through seeded synapses and integrates them
with leaky integrate-and-fire neurons for several ticks. Spike frequency is the
fixed-dimensional feature vector.

The policy is `z = W x + b`, followed by a numerically stable masked softmax. The
baseline is a separate linear value head `V(x)`. For each on-policy episode:

`G_t = r_t + gamma r_(t+1) + ...`

`A_t = G_t - V(x_t)`

The policy parameters are updated in the direction that increases
`A_t log pi(a_t | x_t)`, while the baseline minimizes squared return error.
Adam-compatible optimizer state can be added without changing the frozen network;
the minimal implementation uses a deterministic gradient step.

package main

import "fmt"

type HumanAgent struct {
	id   int
	Sign string
}

func NewHumanAgent(id int, sign string) (agent *HumanAgent) {
	agent = new(HumanAgent)
	agent.id = id
	agent.Sign = sign
	return
}

func (agent *HumanAgent) FetchMessage() string {
	return "-"
}

func (agent *HumanAgent) FetchMove(state State, pa []Action) (action Action, err error) {
	fmt.Print("\n                         \r")
	fmt.Printf("%s > Your move? ", agent.Sign)

	var pos int
	_, err = fmt.Scanln(&pos)

	fmt.Print("\r\033[F\033[F")

	if err != nil {
		return MNKAction{-1, -1}, err
	}

	// TODO: Fix this. The agent must have real access to these variables! (board.m)
	return MNKAction{(pos - 1) / board.m, (pos - 1) % board.m}, nil
}

func (agent *HumanAgent) GameOver(state State) {}

func (agent *HumanAgent) GetSign() string {
	return agent.Sign
}

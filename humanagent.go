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

func (agent *HumanAgent) FetchMove(state [][]int) (pos int, err error) {
	fmt.Print("\n                         \r")
	fmt.Printf("%s > Your move? ", agent.Sign)

	_, err = fmt.Scanln(&pos)

	fmt.Print("\r\033[F\033[F")

	return
}

func (agent *HumanAgent) GameOver(state [][]int) {}

func (agent *HumanAgent) GetSign() string {
	return agent.Sign
}

package main

import "fmt"

type HumanAgent struct {
	Sign string
}

func NewHumanAgent(sign string) (agent *HumanAgent) {
	agent = new(HumanAgent)
	agent.Sign = sign
	return
}

func (agent *HumanAgent) FetchMove() (pos int, err error) {
	fmt.Print("\n                         \r")
	fmt.Printf("%s > Your move? ", agent.Sign)

	_, err = fmt.Scanln(&pos)

	fmt.Print("\r\033[F\033[F")

	return
}

func (agent *HumanAgent) GetSign() string {
	return agent.Sign
}

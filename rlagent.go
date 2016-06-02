package main

import (
	"encoding/gob"
	"fmt"
	"math/rand"
	"os"
)

type RLAgent struct {
	// Agent PlayerID
	id int

	// View settings
	Sign string

	// Game definition
	m, n, k int

	// RL parameters
	Learning          bool
	LearningRate      float64 //alpha
	DiscountFactor    float64
	ExplorationFactor float64 //epsilon

	// States stash
	values          map[string]float64
	prevStateAction [][]int
	prevReward      float64
	message         string
}

type RLAgentKnowledge struct {
	Values           map[string]float64
	Iterations       uint
	randomDispersion [9]int
}

var rlKnowledge RLAgentKnowledge

func NewRLAgent(id int, sign string, m, n, k int, learn bool) (agent *RLAgent) {
	agent = new(RLAgent)
	agent.id = id
	agent.Sign = sign

	agent.m = m
	agent.n = n
	agent.k = k

	// Default values
	agent.Learning = learn
	agent.LearningRate = 0.2
	agent.DiscountFactor = 0.8
	agent.ExplorationFactor = 0.25

	// Initiate stash
	if rlKnowledge.Iterations == 0 {
		rlKnowledge.Values = make(map[string]float64)
	}
	agent.values = rlKnowledge.Values

	return
}

func (agent *RLAgent) FetchMessage() (message string) {
	message = agent.message
	agent.message = "-"
	return
}

func (agent *RLAgent) FetchMove(state [][]int) (move int, err error) {
	var moveValue float64

	var e = rand.Float64()
	if e < agent.ExplorationFactor {
		agent.message = fmt.Sprintf("Exploratory action (%f)", e)

		// Choose a random move
		var emptyCells []int
		for i := range state {
			for j := range state[i] {
				if state[i][j] == 0 {
					emptyCells = append(emptyCells, i*agent.m+j+1)
				}
			}
		}

		move = emptyCells[rand.Intn(len(emptyCells))]
		rlKnowledge.randomDispersion[move-1]++
		var i = (move - 1) / agent.m
		var j = (move - 1) % agent.m
		state[i][j] = agent.id
		moveValue = agent.lookup(state)
		state[i][j] = 0

	} else {
		agent.message = fmt.Sprintf("Greedy action (%f)", e)

		// Choose a greedy move
		var maxReward float64 = 0
		var maxMove = 0
		var first = true

		for i := range state {
			for j := range state[i] {
				if state[i][j] == 0 {
					// REVIEW: Is tempState = copyState(state) too costly? Or perhaps do it in main
					state[i][j] = agent.id
					value := agent.lookup(state)
					state[i][j] = 0

					if value > maxReward || first {
						maxReward = value
						maxMove = i*agent.m + j + 1
						first = false
					}
				}
			}
		}
		move = maxMove
		moveValue = maxReward
	}

	if agent.Learning {
		agent.learn(moveValue)
	}

	agent.prevStateAction = copyState(state)
	agent.prevStateAction[(move-1)/agent.m][(move-1)%agent.m] = agent.id
	agent.prevReward = agent.value(agent.prevStateAction)

	return
}

func (agent *RLAgent) GameOver(state [][]int) {
	if agent.Learning {
		// TODO: Check for prevStateAction == state as well so we don't train twice
		agent.learn(agent.value(state))
	}

	rlKnowledge.Iterations++
}

func (agent *RLAgent) GetSign() string {
	return agent.Sign
}

// learn calculates new value for given state
func (agent *RLAgent) learn(qMax float64) {
	var mState = marshallState(agent.prevStateAction, agent.id)
	var oldVal = agent.values[mState]

	// REVIEW: Learning Rate may decrease gradually (for stochastic environments)
	// REVIEW: Discount Factor may increase gradually (when estimating reward)

	agent.values[mState] = oldVal + (agent.LearningRate *
		(agent.prevReward + (agent.DiscountFactor * qMax) - oldVal))
}

// lookup returns the Q-value for the given state
func (agent *RLAgent) lookup(state [][]int) float64 {
	var mState = marshallState(state, agent.id) // Marshalled state
	val, ok := agent.values[mState]
	if !ok {
		val = agent.value(state)
		agent.values[mState] = val
	}
	return val
}

// value returns the reward for the given state
func (agent *RLAgent) value(state [][]int) float64 {
	switch evaluate(state) {
	case agent.id: // Agent won
		return 1
	case 0: // Game goes on
		return 0
	case -1: // Draw
		return -0.5
	default: // Agent lost
		return -1
	}
}

// storeKnowledge writes the knowledge map to given path
func (k *RLAgentKnowledge) saveToFile(path string) bool {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("[error] Could not open writable knowledge file on disk!")
		fmt.Println(err)
		return false
	}
	defer file.Close()

	enc := gob.NewEncoder(file)
	err = enc.Encode(rlKnowledge)
	if err != nil {
		fmt.Println("[error] Encoding of knowledge failed!")
		fmt.Println(err)
		return false
	}

	return true
}

// retrieveKnowledge reads the knowledge from given path to knowledge map
func (k *RLAgentKnowledge) loadFromFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("[error] Could not open readable knowledge file on disk!")
		fmt.Println(err)
		return false
	}
	defer file.Close()

	dec := gob.NewDecoder(file)
	err = dec.Decode(&rlKnowledge)
	if err != nil {
		fmt.Println("[error] Decoding of knowledge failed!")
		fmt.Println(err)
		return false
	}

	return true
}

func marshallState(state [][]int, agentID int) (m string) {
	for i := range state {
		for j := range state[i] {
			// Regulate based on current agent's ID so all agents are one!
			if state[i][j] > 0 {
				if state[i][j] == agentID {
					m += "X"
				} else {
					m += "O"
				}
			} else {
				m += "-"
			}
		}
	}
	return
}

func copyState(state [][]int) (c [][]int) {
	c = make([][]int, len(state))
	for i := range state {
		c[i] = make([]int, len(state[i]))
		copy(c[i], state[i])
	}
	return
}

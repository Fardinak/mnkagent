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
	values map[string]float64
	prev   struct {
		state  MNKState
		action MNKAction
		reward float64
	}
	message string
}

type RLAgentKnowledge struct {
	Values           map[string]float64
	Iterations       uint
	randomDispersion []int
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
		rlKnowledge.randomDispersion = make([]int, m*n)
	} else {
		var tmp []int = make([]int, len(rlKnowledge.randomDispersion))
		copy(tmp, rlKnowledge.randomDispersion)
		rlKnowledge.randomDispersion = make([]int, m*n)
		copy(rlKnowledge.randomDispersion, tmp)
	}
	agent.values = rlKnowledge.Values

	return
}

func (agent *RLAgent) FetchMessage() (message string) {
	message = agent.message
	agent.message = ""
	return
}

func (agent *RLAgent) FetchMove(state State, possibleActions []Action) (Action, error) {
	// REVIEW: Rename to Move, and accept a function to do it, which returns the reward
	var s MNKState = state.(MNKState)
	var action MNKAction
	var qMax float64

	var e = rand.Float64()
	if e < agent.ExplorationFactor {
		agent.message = fmt.Sprintf("Exploratory action (%f)", e)

		// Choose a random move
		rndi := rand.Intn(len(possibleActions))
		action = possibleActions[rndi].GetParams().(MNKAction)
		rlKnowledge.randomDispersion[action.Y*agent.m+action.X]++
		qMax = agent.lookup(s, action)

	} else {
		agent.message = fmt.Sprintf("Greedy action (%f)", e)

		// Choose a greedy move
		var first = true
		for i := range s {
			for j := range s[i] {
				if s[i][j] == 0 {
					a := MNKAction{i, j}
					v := agent.lookup(s, a)

					if v > qMax || first {
						qMax = v
						action = a
						first = false
					}
				}
			}
		}
	}

	if agent.Learning {
		agent.learn(qMax)
	}

	agent.prev.state = s //.Clone()
	agent.prev.action = action
	agent.prev.reward = agent.value(agent.prev.state, agent.prev.action)

	return action, nil
}

func (agent *RLAgent) GameOver(state State) {
	var s MNKState = state.(MNKState)

	if agent.Learning {
		// Bypass the marshaller's action addition with (-1, -1)
		agent.learn(agent.lookup(s, MNKAction{-1, -1}))
	}

	// Restart for the next episode
	agent.prev.state = MNKState{}
	agent.prev.action = MNKAction{}
	agent.prev.reward = 0
	agent.message = ""

	rlKnowledge.Iterations++
}

func (agent *RLAgent) GetSign() string {
	return agent.Sign
}

// learn calculates new value for given state
func (agent *RLAgent) learn(qMax float64) {
	// Ignore an empty state-action (happens on first move)
	if len(agent.prev.state) == 0 {
		return
	}

	var mState = marshallState(agent.id, agent.prev.state, agent.prev.action)
	var oldVal = agent.values[mState]

	// REVIEW: Learning Rate may decrease gradually (for stochastic environments)
	// REVIEW: Discount Factor may increase gradually (when estimating reward)

	agent.values[mState] = oldVal + (agent.LearningRate *
		(agent.prev.reward + (agent.DiscountFactor * qMax) - oldVal))
}

// lookup returns the Q-value for the given state
func (agent *RLAgent) lookup(state MNKState, action MNKAction) float64 {
	var mState = marshallState(agent.id, state, action) // Marshalled state
	val, ok := agent.values[mState]
	if !ok {
		val = agent.value(state, action)
		agent.values[mState] = val
	}
	return val
}

// value returns the reward for the given state
func (agent *RLAgent) value(state MNKState, action MNKAction) float64 {
	// TODO: Fix this. The agent must have real access to the evaluation function
	if action != (MNKAction{-1, -1}) {
		switch board.EvaluateAction(agent.id, action) {
		case 1: // Agent won
			return 1
		case 0: // Game goes on
			return 0
		case -1: // Draw
			return -0.5
		}
	}

	switch board.Evaluate() {
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
	err = enc.Encode(k)
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
	err = dec.Decode(k)
	if err != nil {
		fmt.Println("[error] Decoding of knowledge failed!")
		fmt.Println(err)
		return false
	}

	return true
}

func marshallState(agentID int, state MNKState, action MNKAction) (m string) {
	for i := range state {
		for j := range state[i] {
			// Include action in state
			if i == action.Y && j == action.X {
				m += "X"
				continue
			}

			switch state[i][j] {
			case 0:
				m += "-"
			case agentID:
				m += "X"
			default:
				m += "O"
			}
		}
	}
	return
}

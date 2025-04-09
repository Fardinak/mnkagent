// Package ui provides user interface utilities for the mnkagent project
package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"mnkagent/common"
	"mnkagent/game"
)

// DisplayConfig contains display configuration
type DisplayConfig struct {
	NoDisplay bool
	M, N      int
	FirstRun  bool
}

// Display handles the game board display
type Display struct {
	config DisplayConfig
	signs  map[int]string
}

// NewDisplay creates a new display manager
func NewDisplay(config DisplayConfig, signs map[int]string) *Display {
	return &Display{
		config: config,
		signs:  signs,
	}
}

// ShowBoard displays the current game board
func (d *Display) ShowBoard(state common.State) {
	// Skip if display is disabled
	if d.config.NoDisplay {
		return
	}

	// Cast to MNK state
	board := state.(game.MNKState)
	
	// Reset cursor position if not first run
	if !d.config.FirstRun {
		// Move cursor to beginning of the line and up to the top of the board
		reset := "\r"
		for i := 0; i < d.config.N*2+1; i++ {
			reset += "\033[F"
		}
		fmt.Print(reset)
	} else {
		d.config.FirstRun = false
	}

	// Draw the board
	for i := 0; i < d.config.N; i++ {
		line := ""
		if i == 0 {
			// Top border
			line = "\u2554"
			for j := 0; j < d.config.M; j++ {
				line += "\u2550\u2550\u2550\u2550\u2550"
				if j < d.config.M-1 {
					line += "\u2564"
				} else {
					line += "\u2557"
				}
			}
		} else {
			// Middle separator
			line = "\u2551"
			for j := 0; j < d.config.M; j++ {
				line += "\u2500\u2500\u2500\u2500\u2500"
				if j < d.config.M-1 {
					line += "\u253c"
				} else {
					line += "\u2551"
				}
			}
		}
		fmt.Println(line)

		// Row content
		line = "\u2551"
		for j := 0; j < d.config.M; j++ {
			if j != 0 {
				line += "\u2502"
			}

			index := i*d.config.M + j + 1
			padding := [2]string{"", ""}
			var mark string

			if board[i][j] == 0 {
				// Empty cell shows position number
				mark = fmt.Sprintf("\033[37m%d\033[0m", index)

				if index < 10 {
					padding = [2]string{"  ", "  "}
				} else if index < 100 {
					padding = [2]string{" ", "  "}
				} else if index < 1000 {
					padding = [2]string{" ", " "}
				}
			} else {
				// Occupied cell shows player's sign
				mark = d.signs[board[i][j]]
				padding = [2]string{"  ", "  "}
			}

			line += padding[0] + mark + padding[1]
		}
		line += "\u2551"
		fmt.Println(line)

		// Bottom border
		if i+1 == len(board) {
			line = "\u255a"
			for j := 0; j < d.config.M; j++ {
				line += "\u2550\u2550\u2550\u2550\u2550"
				if j < d.config.M-1 {
					line += "\u2567"
				} else {
					line += "\u255d"
				}
			}
			fmt.Println(line)
		}
	}
}

// ShowMessages displays messages from agents
func (d *Display) ShowMessages(agents []common.Agent) {
	if d.config.NoDisplay {
		return
	}

	messages := make([]string, 0, len(agents))
	for _, agent := range agents {
		msg := agent.FetchMessage()
		if msg != "" {
			messages = append(messages, fmt.Sprintf("Agent %s: %s", agent.GetSign(), msg))
		}
	}

	if len(messages) > 0 {
		// Clear previous line
		fmt.Printf("\033[2K\r%s", strings.Join(messages, " / "))
	}
}

// ShowStats displays game statistics
func (d *Display) ShowStats(log []int, agents map[int]common.Agent, showRandomDispersion bool, randomDispersion []int) {
	// Determine the winner
	var winnerSign string
	winner := d.findMaxIndex(log)

	if winner == 0 {
		winnerSign = "DRAW"
	} else {
		winnerSign = agents[winner].GetSign()
	}

	// Show stats
	fmt.Printf("Stats: %s/%s/Draw = %d/%d/%d\nOverall winner: %s\n",
		agents[1].GetSign(), agents[2].GetSign(), log[1], log[2], log[0],
		winnerSign)

	// Show random move dispersion if requested
	if showRandomDispersion && randomDispersion != nil {
		fmt.Println("Random move dispersion:")
		for i := 0; i < len(randomDispersion); i++ {
			fmt.Printf("%d: %d\n", i+1, randomDispersion[i])
		}
	}
}

// ClearPrompt clears the current line
func (d *Display) ClearPrompt() {
	if !d.config.NoDisplay {
		fmt.Print("\033[2K\r")
	}
}

// ShowProgressBar displays a progress bar for training
func (d *Display) ShowProgressBar(progress int, width int, msg string, isComplete bool) {
	if d.config.NoDisplay {
		return
	}

	// Choose color based on completion status
	color := "\033[41;3m" // Red for in-progress
	if isComplete {
		color = "\033[46;3m" // Cyan for completed
	}
	colorReset := "\033[0m"

	// Format progress percentage
	fProgress := fmt.Sprintf("%d%%", progress)

	// Prepare message with ellipsis if too long
	m := msg
	if len(m) > width-8 {
		m = m[:width-8] + "..."
	}

	// Create the progress bar
	pb := make([]byte, width)
	
	// Fill with spaces
	for i := 0; i < width; i++ {
		pb[i] = ' '
	}

	// Calculate positions
	progIndex := width * progress / 100
	msgOffset := (width-8)/2 - len(m)/2
	if msgOffset < 1 {
		msgOffset = 1
	}

	// Place message in the center
	copy(pb[msgOffset:], m)

	// Place progress percentage at the right
	copy(pb[width-len(fProgress)-1:], fProgress)

	// Apply color up to the progress point
	progressBar := color + string(pb[:progIndex]) + colorReset + string(pb[progIndex:]) + "\r"

	fmt.Print(progressBar)
}

// ShowSeparator displays a separator line
func (d *Display) ShowSeparator() {
	if !d.config.NoDisplay {
		fmt.Print("___________________________________\n\n")
	}
}

// ResetFirstRun resets the first run flag
func (d *Display) ResetFirstRun() {
	d.config.FirstRun = true
}

// findMaxIndex returns the index of the maximum value in a slice
func (d *Display) findMaxIndex(arr []int) int {
	var maxVal int
	var maxIdx int
	
	for i, val := range arr {
		if val > maxVal {
			maxVal = val
			maxIdx = i
		}
	}
	
	return maxIdx
}

// GetTerminalSize returns the width and height of the terminal
func GetTerminalSize() (width, height int) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	output, err := cmd.Output()
	if err != nil {
		return 80, 24 // Default fallback size
	}
	
	fmt.Sscanf(string(output), "%d %d", &height, &width)
	return
}
package common

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
	"math/rand"
)

// NeuralNetwork represents a simple feed-forward neural network
type NeuralNetwork struct {
	InputSize   int
	HiddenSize  int
	OutputSize  int
	LearningRate float64
	
	// Weights
	WeightsIH [][]float64 // Input to Hidden
	WeightsHO [][]float64 // Hidden to Output
	
	// Biases
	BiasH []float64 // Hidden layer bias
	BiasO []float64 // Output layer bias
}

// NewNeuralNetwork creates a new neural network with random weights
func NewNeuralNetwork(inputSize, hiddenSize, outputSize int, learningRate float64) *NeuralNetwork {
	nn := &NeuralNetwork{
		InputSize:    inputSize,
		HiddenSize:   hiddenSize,
		OutputSize:   outputSize,
		LearningRate: learningRate,
	}
	
	// Initialize weights with random values
	nn.WeightsIH = make([][]float64, hiddenSize)
	for i := range nn.WeightsIH {
		nn.WeightsIH[i] = make([]float64, inputSize)
		for j := range nn.WeightsIH[i] {
			nn.WeightsIH[i][j] = rand.Float64()*2 - 1 // Values between -1 and 1
		}
	}
	
	nn.WeightsHO = make([][]float64, outputSize)
	for i := range nn.WeightsHO {
		nn.WeightsHO[i] = make([]float64, hiddenSize)
		for j := range nn.WeightsHO[i] {
			nn.WeightsHO[i][j] = rand.Float64()*2 - 1 // Values between -1 and 1
		}
	}
	
	// Initialize biases
	nn.BiasH = make([]float64, hiddenSize)
	for i := range nn.BiasH {
		nn.BiasH[i] = rand.Float64()*2 - 1
	}
	
	nn.BiasO = make([]float64, outputSize)
	for i := range nn.BiasO {
		nn.BiasO[i] = rand.Float64()*2 - 1
	}
	
	return nn
}

// Sigmoid activation function
func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

// SigmoidDerivative calculates derivative of sigmoid for backpropagation
func sigmoidDerivative(y float64) float64 {
	return y * (1 - y)
}

// Predict performs a forward pass and returns the network's output
func (nn *NeuralNetwork) Predict(inputs []float64) ([]float64, error) {
	if len(inputs) != nn.InputSize {
		return nil, fmt.Errorf("expected %d inputs, got %d", nn.InputSize, len(inputs))
	}
	
	// Calculate hidden layer outputs
	hidden := make([]float64, nn.HiddenSize)
	for i := 0; i < nn.HiddenSize; i++ {
		sum := nn.BiasH[i]
		for j := 0; j < nn.InputSize; j++ {
			sum += inputs[j] * nn.WeightsIH[i][j]
		}
		hidden[i] = sigmoid(sum)
	}
	
	// Calculate output layer
	outputs := make([]float64, nn.OutputSize)
	for i := 0; i < nn.OutputSize; i++ {
		sum := nn.BiasO[i]
		for j := 0; j < nn.HiddenSize; j++ {
			sum += hidden[j] * nn.WeightsHO[i][j]
		}
		outputs[i] = sigmoid(sum)
	}
	
	return outputs, nil
}

// Train trains the network using backpropagation
func (nn *NeuralNetwork) Train(inputs []float64, targets []float64) error {
	if len(inputs) != nn.InputSize {
		return fmt.Errorf("expected %d inputs, got %d", nn.InputSize, len(inputs))
	}
	if len(targets) != nn.OutputSize {
		return fmt.Errorf("expected %d targets, got %d", nn.OutputSize, len(targets))
	}
	
	// Forward pass
	// Calculate hidden layer outputs
	hidden := make([]float64, nn.HiddenSize)
	for i := 0; i < nn.HiddenSize; i++ {
		sum := nn.BiasH[i]
		for j := 0; j < nn.InputSize; j++ {
			sum += inputs[j] * nn.WeightsIH[i][j]
		}
		hidden[i] = sigmoid(sum)
	}
	
	// Calculate output layer
	outputs := make([]float64, nn.OutputSize)
	for i := 0; i < nn.OutputSize; i++ {
		sum := nn.BiasO[i]
		for j := 0; j < nn.HiddenSize; j++ {
			sum += hidden[j] * nn.WeightsHO[i][j]
		}
		outputs[i] = sigmoid(sum)
	}
	
	// Backpropagation
	// Calculate output layer errors
	outputErrors := make([]float64, nn.OutputSize)
	for i := 0; i < nn.OutputSize; i++ {
		outputErrors[i] = targets[i] - outputs[i]
	}
	
	// Calculate output layer gradients
	outputGradients := make([]float64, nn.OutputSize)
	for i := 0; i < nn.OutputSize; i++ {
		outputGradients[i] = outputErrors[i] * sigmoidDerivative(outputs[i]) * nn.LearningRate
	}
	
	// Calculate hidden layer errors
	hiddenErrors := make([]float64, nn.HiddenSize)
	for i := 0; i < nn.HiddenSize; i++ {
		sum := 0.0
		for j := 0; j < nn.OutputSize; j++ {
			sum += outputErrors[j] * nn.WeightsHO[j][i]
		}
		hiddenErrors[i] = sum
	}
	
	// Calculate hidden layer gradients
	hiddenGradients := make([]float64, nn.HiddenSize)
	for i := 0; i < nn.HiddenSize; i++ {
		hiddenGradients[i] = hiddenErrors[i] * sigmoidDerivative(hidden[i]) * nn.LearningRate
	}
	
	// Update weights and biases
	// Update hidden to output weights
	for i := 0; i < nn.OutputSize; i++ {
		for j := 0; j < nn.HiddenSize; j++ {
			nn.WeightsHO[i][j] += outputGradients[i] * hidden[j]
		}
		nn.BiasO[i] += outputGradients[i]
	}
	
	// Update input to hidden weights
	for i := 0; i < nn.HiddenSize; i++ {
		for j := 0; j < nn.InputSize; j++ {
			nn.WeightsIH[i][j] += hiddenGradients[i] * inputs[j]
		}
		nn.BiasH[i] += hiddenGradients[i]
	}
	
	return nil
}

// GobEncode implements the gob.GobEncoder interface
func (nn *NeuralNetwork) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	
	// Create a version that can be encoded
	data := struct {
		InputSize    int
		HiddenSize   int
		OutputSize   int
		LearningRate float64
		WeightsIH    [][]float64
		WeightsHO    [][]float64
		BiasH        []float64
		BiasO        []float64
	}{
		InputSize:    nn.InputSize,
		HiddenSize:   nn.HiddenSize,
		OutputSize:   nn.OutputSize,
		LearningRate: nn.LearningRate,
		WeightsIH:    nn.WeightsIH,
		WeightsHO:    nn.WeightsHO,
		BiasH:        nn.BiasH,
		BiasO:        nn.BiasO,
	}
	
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

// GobDecode implements the gob.GobDecoder interface
func (nn *NeuralNetwork) GobDecode(data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	
	// Create a temporary structure to decode into
	var temp struct {
		InputSize    int
		HiddenSize   int
		OutputSize   int
		LearningRate float64
		WeightsIH    [][]float64
		WeightsHO    [][]float64
		BiasH        []float64
		BiasO        []float64
	}
	
	err := dec.Decode(&temp)
	if err != nil {
		return err
	}
	
	// Update the neural network with the decoded data
	nn.InputSize = temp.InputSize
	nn.HiddenSize = temp.HiddenSize
	nn.OutputSize = temp.OutputSize
	nn.LearningRate = temp.LearningRate
	nn.WeightsIH = temp.WeightsIH
	nn.WeightsHO = temp.WeightsHO
	nn.BiasH = temp.BiasH
	nn.BiasO = temp.BiasO
	
	return nil
}
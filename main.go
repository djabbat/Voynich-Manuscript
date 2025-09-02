package main

import (
	"fmt"
	"math"
	"strings"
)

// calculateShannonEntropy computes the Shannon entropy in bits/symbol for a given string.
// Lower entropy may indicate more structured data (e.g., natural language).
func calculateShannonEntropy(data string) float64 {
	if len(data) == 0 {
		return 0
	}

	// Count frequency of each character
	charCounts := make(map[rune]int)
	for _, char := range data {
		charCounts[char]++
	}

	var entropy float64
	totalChars := float64(len(data))

	// Calculate entropy using formula: H = -Σ p(x_i) * log2(p(x_i))
	for _, count := range charCounts {
		probability := float64(count) / totalChars
		entropy -= probability * math.Log2(probability)
	}

	return entropy
}

// decodeLZ77 attempts to decompress a bitstream using LZ77-like algorithm
func decodeLZ77(bitStream string, offsetBits, lengthBits int) (string, error) {
	var output strings.Builder
	searchBuffer := "" // Sliding window/dictionary
	position := 0

	for position < len(bitStream) {
		// Check if we have enough bits for a command flag
		if position+1 > len(bitStream) {
			return output.String(), fmt.Errorf("unexpected end of stream at position %d", position)
		}

		// Read command flag (1 bit)
		flag := bitStream[position : position+1]
		position++

		if flag == "0" {
			// Literal character: read next 8 bits as ASCII
			if position+8 > len(bitStream) {
				return output.String(), fmt.Errorf("incomplete literal at position %d", position)
			}

			literalBits := bitStream[position : position+8]
			position += 8

			// Convert binary to character
			charCode := 0
			for i, bit := range literalBits {
				if bit == '1' {
					charCode += 1 << (7 - i)
				}
			}

			// Add printable characters only
			if charCode >= 32 && charCode <= 126 {
				character := byte(charCode)
				output.WriteByte(character)
				searchBuffer += string(character)

				// Maintain sliding window size
				if len(searchBuffer) > 1<<offsetBits {
					searchBuffer = searchBuffer[1:]
				}
			}

		} else if flag == "1" {
			// Back-reference: read (offsetBits + lengthBits) for (distance, length) tuple
			if position+offsetBits+lengthBits > len(bitStream) {
				return output.String(), fmt.Errorf("incomplete back-reference at position %d", position)
			}

			// Read offset bits
			offsetBitStr := bitStream[position : position+offsetBits]
			position += offsetBits
			offset := 0
			for i, bit := range offsetBitStr {
				if bit == '1' {
					offset += 1 << (offsetBits - 1 - i)
				}
			}

			// Read length bits
			lengthBitStr := bitStream[position : position+lengthBits]
			position += lengthBits
			length := 0
			for i, bit := range lengthBitStr {
				if bit == '1' {
					length += 1 << (lengthBits - 1 - i)
				}
			}

			// Validate and apply back-reference
			if offset > len(searchBuffer) || length == 0 {
				continue // Invalid reference, skip
			}

			startPos := len(searchBuffer) - offset
			for i := 0; i < length; i++ {
				if startPos+i >= len(searchBuffer) {
					break // Avoid out-of-bounds
				}
				character := searchBuffer[startPos+i]
				output.WriteByte(character)
				searchBuffer += string(character)
			}

			// Maintain sliding window size
			if len(searchBuffer) > 1<<offsetBits {
				searchBuffer = searchBuffer[len(searchBuffer)-(1<<offsetBits):]
			}
		}
	}

	return output.String(), nil
}

// generateBitStream creates a demonstration bitstream from text using simple encoding
func generateBitStream(text string) string {
	var bitStream strings.Builder
	for _, char := range text {
		// Simple 8-bit ASCII encoding
		for i := 7; i >= 0; i-- {
			if (char >> uint(i)) & 1 == 1 {
				bitStream.WriteString("1")
			} else {
				bitStream.WriteString("0")
			}
		}
	}
	return bitStream.String()
}

func main() {
	// Demonstration text (simulating possible Voynich content)
	testText := "the rain in spain falls mainly on the plain the rain in spain falls mainly"
	fmt.Printf("Original text: %s\n\n", testText)

	// Generate demonstration bitstream
	bitStream := generateBitStream(testText)
	fmt.Printf("Generated bitstream (%d bits):\n%s\n\n", len(bitStream), bitStream)

	// Test parameters for LZ77 decompression
	offsetBitsOptions := []int{9, 10, 11} // Bit lengths for offset field
	lengthBitsOptions := []int{3, 4, 5}   // Bit lengths for length field

	bestEntropy := math.MaxFloat64
	bestResult := ""
	bestParams := ""

	fmt.Println("Testing LZ77 parameters:")
	fmt.Println("OffsetBits | LengthBits | Entropy | Output Sample")
	fmt.Println("-----------|------------|---------|---------------")

	// Test all parameter combinations
	for _, offsetBits := range offsetBitsOptions {
		for _, lengthBits := range lengthBitsOptions {
			result, err := decodeLZ77(bitStream, offsetBits, lengthBits)
			if err != nil {
				fmt.Printf("%9d | %10d | %8s | Error: %v\n", 
					offsetBits, lengthBits, "N/A", err)
				continue
			}

			// Calculate entropy of decompressed result
			entropy := calculateShannonEntropy(result)
			
			// Display sample of output
			sample := result
			if len(sample) > 20 {
				sample = sample[:20] + "..."
			}

			fmt.Printf("%9d | %10d | %7.4f | %s\n", 
				offsetBits, lengthBits, entropy, sample)

			// Track best result (lowest entropy)
			if entropy < bestEntropy {
				bestEntropy = entropy
				bestResult = result
				bestParams = fmt.Sprintf("offsetBits=%d, lengthBits=%d", 
					offsetBits, lengthBits)
			}
		}
	}

	// Display best result
	fmt.Printf("\nBest parameters: %s\n", bestParams)
	fmt.Printf("Lowest entropy: %.4f bits/character\n", bestEntropy)
	fmt.Printf("Decompressed result (%d characters):\n%s\n", 
		len(bestResult), bestResult)

	// Entropy analysis
	originalEntropy := calculateShannonEntropy(testText)
	fmt.Printf("\nEntropy comparison:\n")
	fmt.Printf("Original text:  %.4f bits/character\n", originalEntropy)
	fmt.Printf("Decompressed:   %.4f bits/character\n", bestEntropy)
	fmt.Printf("Bitstream:      %.4f bits/character\n", 
		calculateShannonEntropy(bitStream))
}

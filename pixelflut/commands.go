package pixelflut

import (
	"math/rand"
)

// Commands represent a list of messages to be sent to a pixelflut server.
type Commands [][]byte

// Chunk splits commands into equally sized chunks, while flattening each chunk
// so that all commands are concatenated as a single `[]byte`.
func (c Commands) Chunk(numChunks int) [][]byte {
	chunks := make([][]byte, numChunks)
	chunkLength := len(c) / numChunks
	for i := 0; i < numChunks; i++ {
		cmdOffset := i * chunkLength
		for j := 0; j < chunkLength; j++ {
			chunks[i] = append(chunks[i], c[cmdOffset+j]...)
		}
	}
	return chunks
}

// Shuffle reorders commands randomly, in place.
func (c Commands) Shuffle() {
	for i := range c {
		j := rand.Intn(i + 1)
		c[i], c[j] = c[j], c[i]
	}
}

package idgen

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	DefaultPrefix    = "atom"
	MinHashLength    = 4
	MaxHashLength    = 6
	DefaultHashLength = 4
)

type Generator struct {
	prefix   string
	existing map[string]bool
}

func New(prefix string) *Generator {
	if prefix == "" {
		prefix = DefaultPrefix
	}
	return &Generator{
		prefix:   prefix,
		existing: make(map[string]bool),
	}
}

func (g *Generator) SetExisting(ids []string) {
	g.existing = make(map[string]bool)
	for _, id := range ids {
		g.existing[id] = true
	}
}

func (g *Generator) Generate(title string, createdAt time.Time) (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	data := fmt.Sprintf("%s|%s|%s", title, createdAt.Format(time.RFC3339Nano), hex.EncodeToString(randomBytes))
	hash := sha256.Sum256([]byte(data))
	hashHex := hex.EncodeToString(hash[:])

	for length := DefaultHashLength; length <= MaxHashLength; length++ {
		id := fmt.Sprintf("%s-%s", g.prefix, hashHex[:length])
		if !g.existing[id] {
			g.existing[id] = true
			return id, nil
		}
	}

	id := fmt.Sprintf("%s-%s", g.prefix, hashHex[:MaxHashLength+2])
	g.existing[id] = true
	return id, nil
}

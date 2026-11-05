package goak

import (
	"fmt"
	"github.com/tacf/goak/internal/idhash"
	"strings"
)

// ID identifies a widget. Zero means none.
type ID uint32

func hashStr(s string, seed ID) ID { return ID(idhash.Str(s, uint32(seed))) }

// DisplayLabel returns the visible part of a label, the text before "##".
func DisplayLabel(label string) string {
	if i := strings.Index(label, "##"); i >= 0 {
		return label[:i]
	}
	return label
}

// PushID pushes a string, integer or any printable value onto the ID stack.
// String and integer values hash without allocating.
func (c *Context) PushID(v any) {
	seed := c.idSeed()
	var id ID
	switch t := v.(type) {
	case string:
		id = hashLabel(t, seed)
	case int:
		id = hashUint('i', uint64(t), seed)
	case int32:
		id = hashUint('i', uint64(t), seed)
	case int64:
		id = hashUint('i', uint64(t), seed)
	case uint:
		id = hashUint('i', uint64(t), seed)
	case uint32:
		id = hashUint('i', uint64(t), seed)
	case uint64:
		id = hashUint('i', uint64(t), seed)
	default:
		id = hashStr(fmt.Sprintf("#v%T%v", v, v), seed)
	}
	c.idStack = append(c.idStack, id)
}

func hashUint(tag byte, v uint64, seed ID) ID { return ID(idhash.Uint(tag, v, uint32(seed))) }

// hashLabel applies the ## and ### rules.
func hashLabel(label string, seed ID) ID { return ID(idhash.Label(label, uint32(seed))) }

// PopID removes the top of the ID stack.
func (c *Context) PopID() {
	if len(c.idStack) == 0 {
		panic("goak: PopID without PushID")
	}
	if c.cur != nil && len(c.idStack) <= c.cur.idDepth {
		panic("goak: PopID without PushID inside this window")
	}
	c.idStack = c.idStack[:len(c.idStack)-1]
}

func (c *Context) idSeed() ID {
	if n := len(c.idStack); n > 0 {
		return c.idStack[n-1]
	}
	return 0
}

// GetID hashes label with the top of the ID stack. "a##b" hashes the whole
// string, so the part after ## only disambiguates. "a###b" hashes only
// "###b" so the visible text can change without changing the ID.
func (c *Context) GetID(label string) ID {
	return hashLabel(label, c.idSeed())
}

// ItemID is GetID plus per frame collision tracking used by widgets.
func (c *Context) ItemID(label string) ID {
	id := c.GetID(label)
	c.seen[id]++
	c.lastID = id
	if c.seen[id] > 1 {
		c.collisions++
	}
	return id
}

// Collisions is the number of duplicate widget IDs seen this frame.
func (c *Context) Collisions() int { return c.collisions }

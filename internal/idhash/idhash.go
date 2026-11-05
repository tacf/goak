// Package idhash is the FNV-1a hashing behind widget IDs.
package idhash

import "strings"

const (
	fnvOffset = 2166136261
	fnvPrime  = 16777619
)

// Str hashes s with FNV-1a after mixing seed.
func Str(s string, seed uint32) uint32 {
	h := uint32(fnvOffset)
	for i := 0; i < 4; i++ {
		h ^= uint32(seed >> (8 * i) & 0xff)
		h *= fnvPrime
	}
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= fnvPrime
	}
	if h == 0 {
		h = 1
	}
	return h
}

// Uint hashes a tagged integer with seed.
func Uint(tag byte, v uint64, seed uint32) uint32 {
	h := uint32(fnvOffset)
	for i := 0; i < 4; i++ {
		h ^= uint32(seed >> (8 * i) & 0xff)
		h *= fnvPrime
	}
	h ^= uint32(tag)
	h *= fnvPrime
	for i := 0; i < 8; i++ {
		h ^= uint32(v >> (8 * i) & 0xff)
		h *= fnvPrime
	}
	if h == 0 {
		h = 1
	}
	return h
}

// Label hashes a label applying the ## and ### rules.
func Label(label string, seed uint32) uint32 {
	if i := strings.LastIndex(label, "###"); i >= 0 {
		label = label[i:]
	}
	return Str(label, seed)
}

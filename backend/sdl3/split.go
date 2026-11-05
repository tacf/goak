package sdl3

// maxChunkVerts is the vertex count addressable by a uint16 index.
const maxChunkVerts = 1 << 16

// chunk is a triangle batch whose indices fit in uint16 after rebasing.
// Vertices VtxMin..VtxMax inclusive of the source list are referenced.
type chunk struct {
	VtxMin  int
	VtxMax  int
	Indices []uint16 // relative to VtxMin
}

// splitIndices splits triangle indices so each chunk spans at most limit
// vertices. Indices are absolute into the vertex array. Triangles are kept
// whole and order is preserved. Trailing indices that do not form a triangle
// are dropped.
func splitIndices(idx []uint32, limit int) []chunk {
	if limit <= 0 || limit > maxChunkVerts {
		limit = maxChunkVerts
	}
	var out []chunk
	var lo, hi int
	var cur []uint32
	flush := func() {
		if len(cur) == 0 {
			return
		}
		rel := make([]uint16, len(cur))
		for i, v := range cur {
			rel[i] = uint16(int(v) - lo)
		}
		out = append(out, chunk{VtxMin: lo, VtxMax: hi, Indices: rel})
		cur = nil
	}
	for i := 0; i+2 < len(idx); i += 3 {
		a, b, c := int(idx[i]), int(idx[i+1]), int(idx[i+2])
		tlo, thi := min(a, b, c), max(a, b, c)
		if thi-tlo+1 > limit {
			continue // triangle alone exceeds the limit
		}
		if len(cur) > 0 {
			nlo, nhi := min(lo, tlo), max(hi, thi)
			if nhi-nlo+1 > limit {
				flush()
			} else {
				lo, hi = nlo, nhi
			}
		}
		if len(cur) == 0 {
			lo, hi = tlo, thi
		}
		cur = append(cur, idx[i], idx[i+1], idx[i+2])
	}
	flush()
	return out
}

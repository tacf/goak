package sdl3

import "testing"

func TestSplitSmallIsOneChunk(t *testing.T) {
	got := splitIndices([]uint32{0, 1, 2, 2, 1, 3}, 0)
	if len(got) != 1 || got[0].VtxMin != 0 || got[0].VtxMax != 3 || len(got[0].Indices) != 6 {
		t.Fatalf("got %+v", got)
	}
}

func TestSplitRebases(t *testing.T) {
	got := splitIndices([]uint32{100000, 100001, 100002}, 0)
	if len(got) != 1 || got[0].VtxMin != 100000 {
		t.Fatalf("got %+v", got)
	}
	if got[0].Indices[0] != 0 || got[0].Indices[2] != 2 {
		t.Fatalf("not rebased: %v", got[0].Indices)
	}
}

func TestSplitOverLimit(t *testing.T) {
	// Triangles over disjoint vertex ranges, limit of 6 vertices.
	var idx []uint32
	for v := uint32(0); v < 30; v += 3 {
		idx = append(idx, v, v+1, v+2)
	}
	got := splitIndices(idx, 6)
	if len(got) != 5 {
		t.Fatalf("want 5 chunks, got %d", len(got))
	}
	total := 0
	for _, c := range got {
		if c.VtxMax-c.VtxMin+1 > 6 {
			t.Fatalf("chunk span too big %+v", c)
		}
		for _, i := range c.Indices {
			if int(i) > c.VtxMax-c.VtxMin {
				t.Fatalf("index out of chunk %+v", c)
			}
		}
		total += len(c.Indices)
	}
	if total != len(idx) {
		t.Fatalf("lost indices: %d of %d", total, len(idx))
	}
}

func TestSplitLargeRealLimit(t *testing.T) {
	// 70000 vertices as quads so the span passes 65536.
	n := 70000 / 4
	idx := make([]uint32, 0, n*6)
	for q := 0; q < n; q++ {
		v := uint32(q * 4)
		idx = append(idx, v, v+1, v+2, v, v+2, v+3)
	}
	got := splitIndices(idx, 0)
	if len(got) != 2 {
		t.Fatalf("want 2 chunks, got %d", len(got))
	}
	sum := 0
	for _, c := range got {
		if c.VtxMax-c.VtxMin+1 > maxChunkVerts {
			t.Fatal("span too big")
		}
		sum += len(c.Indices)
	}
	if sum != len(idx) {
		t.Fatal("lost indices")
	}
}

func TestSplitDropsPartialAndHuge(t *testing.T) {
	if got := splitIndices([]uint32{0, 1}, 0); len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
	if got := splitIndices([]uint32{0, 1, 10}, 4); len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
}

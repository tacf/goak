package goak

type stateEntry struct {
	val    any
	last   int
	maxAge int // 0 means the store default
}

// LongAge is the max age for state that must survive hidden windows.
const LongAge = 600

// StateStore keeps per ID values. Entries not touched for more than
// maxAge frames are dropped by Advance.
type StateStore struct {
	m      map[ID]*stateEntry
	frame  int
	MaxAge int
}

// NewStateStore returns an empty store keeping entries 2 frames.
func NewStateStore() *StateStore {
	return &StateStore{m: map[ID]*stateEntry{}, MaxAge: 2}
}

// Advance sets the current frame and prunes stale entries.
func (s *StateStore) Advance(frame int) {
	s.frame = frame
	for id, e := range s.m {
		age := s.MaxAge
		if e.maxAge > 0 {
			age = e.maxAge
		}
		if frame-e.last > age {
			delete(s.m, id)
		}
	}
}

// Len is the number of live entries.
func (s *StateStore) Len() int { return len(s.m) }

// StateGet returns the value for id or def when missing or of another type.
// A hit marks the entry as seen this frame.
func StateGet[T any](s *StateStore, id ID, def T) T {
	e := s.m[id]
	if e == nil {
		return def
	}
	v, ok := e.val.(T)
	if !ok {
		return def
	}
	e.last = s.frame
	return v
}

// StateSet stores v for id and marks it as seen this frame.
func StateSet[T any](s *StateStore, id ID, v T) {
	s.m[id] = &stateEntry{val: v, last: s.frame}
}

// StateSetKeep is StateSet with its own max age in frames.
func StateSetKeep[T any](s *StateStore, id ID, v T, maxAge int) {
	s.m[id] = &stateEntry{val: v, last: s.frame, maxAge: maxAge}
}

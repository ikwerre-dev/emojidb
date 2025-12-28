package core

import (
	"errors"
	"time"
)

type Row map[string]interface{}

type HotHeap struct {
	Rows      []Row
	Size      int
	MaxRows   int
	CreatedAt time.Time
}

type SealedClump struct {
	Rows     []Row
	Metadata ClumpMetadata
	SealedAt time.Time
}

type ClumpMetadata struct {
	RowCount      int
	SchemaVersion int
	CreatedAt     time.Time
}

func NewHotHeap(maxRows int) *HotHeap {
	return &HotHeap{
		Rows:      make([]Row, 0, maxRows),
		MaxRows:   maxRows,
		CreatedAt: time.Now(),
	}
}

func (t *Table) Insert(record Row) error {
	t.Mu.Lock()
	defer t.Mu.Unlock()

	if t.HotHeap == nil {
		t.HotHeap = NewHotHeap(1000)
	}

	for _, field := range t.Schema.Fields {
		val, ok := record[field.Name]
		if !ok {
			return errors.New("missing field: " + field.Name)
		}

		if field.Unique {
			if _, exists := t.UniqueIndices[field.Name][val]; exists {
				return errors.New("unique constraint violation: " + field.Name)
			}
		}
	}

	for _, field := range t.Schema.Fields {
		if field.Unique {
			t.UniqueIndices[field.Name][record[field.Name]] = struct{}{}
		}
	}

	t.HotHeap.Rows = append(t.HotHeap.Rows, record)

	if len(t.HotHeap.Rows) >= t.HotHeap.MaxRows {
		t.SealHotHeap()
	}

	return nil
}

func (t *Table) Flush() {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	if t.HotHeap != nil && len(t.HotHeap.Rows) > 0 {
		t.SealHotHeap()
	}
}

func (t *Table) SealHotHeap() {
	clump := &SealedClump{
		Rows:     t.HotHeap.Rows,
		SealedAt: time.Now(),
		Metadata: ClumpMetadata{
			RowCount:      len(t.HotHeap.Rows),
			SchemaVersion: t.Schema.Version,
			CreatedAt:     t.HotHeap.CreatedAt,
		},
	}

	t.SealedClumps = append(t.SealedClumps, clump)
	if t.Db != nil {
		t.Db.PersistClump(t.Name, clump)
	}
	t.HotHeap = NewHotHeap(t.HotHeap.MaxRows)
}

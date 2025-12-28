package safety

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/ikwerre-dev/emojidb/core"
	"github.com/ikwerre-dev/emojidb/crypto"
)

type FilterFunc func(core.Row) bool

func Update(db *core.Database, tableName string, filter FilterFunc, update core.Row) error {
	db.Mu.RLock()
	table, ok := db.Tables[tableName]
	db.Mu.RUnlock()

	if !ok {
		return errors.New("table not found")
	}

	table.Mu.Lock()
	defer table.Mu.Unlock()

	var toBackup []core.Row
	var indices []int
	for i, row := range table.HotHeap.Rows {
		if filter(row) {
			toBackup = append(toBackup, row)
			indices = append(indices, i)
		}
	}

	if len(toBackup) > 0 {
		if err := BatchBackupForSafety(db, tableName, toBackup); err != nil {
			return err
		}
		for _, idx := range indices {
			for k, v := range update {
				table.HotHeap.Rows[idx][k] = v
			}
		}
	}

	return nil
}

func Delete(db *core.Database, tableName string, filter FilterFunc) error {
	db.Mu.RLock()
	table, ok := db.Tables[tableName]
	db.Mu.RUnlock()

	if !ok {
		return errors.New("table not found")
	}

	table.Mu.Lock()
	defer table.Mu.Unlock()

	var toBackup []core.Row
	var newRows []core.Row
	for _, row := range table.HotHeap.Rows {
		if filter(row) {
			toBackup = append(toBackup, row)
		} else {
			newRows = append(newRows, row)
		}
	}

	if len(toBackup) > 0 {
		if err := BatchBackupForSafety(db, tableName, toBackup); err != nil {
			return err
		}
	}

	table.HotHeap.Rows = newRows
	return nil
}

func Restore(db *core.Database, timestamp time.Time, accepted bool) error {
	if !accepted {
		return errors.New("recovery aborted")
	}

	db.Mu.Lock()
	defer db.Mu.Unlock()

	_, err := db.SafetyFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	br := bufio.NewReader(db.SafetyFile)
	for {
		size, err := readIntEmoji(br)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		payload := make([]byte, size)
		for i := 0; i < int(size); i++ {
			b, err := crypto.DecodeOne(br)
			if err != nil {
				return err
			}
			payload[i] = b
		}

		decrypted, err := crypto.Decrypt(payload, db.Key)
		if err != nil {
			continue
		}

		var backup SafetyBackup
		if err := json.Unmarshal(decrypted, &backup); err != nil {
			continue
		}

		if backup.Timestamp.Truncate(time.Second).Equal(timestamp.Truncate(time.Second)) {
			if table, ok := db.Tables[backup.TableName]; ok {
				table.Mu.Lock()
				table.HotHeap.Rows = append(table.HotHeap.Rows, backup.Data)
				table.Mu.Unlock()
				return nil
			}
		}
	}

	return errors.New("recovery point not found")
}

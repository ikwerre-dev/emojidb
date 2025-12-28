package safety

import (
	"encoding/binary"
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

	for i, row := range table.HotHeap.Rows {
		if filter(row) {
			BackupForSafety(db, tableName, row)
			for k, v := range update {
				table.HotHeap.Rows[i][k] = v
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

	var newRows []core.Row
	for _, row := range table.HotHeap.Rows {
		if filter(row) {
			BackupForSafety(db, tableName, row)
		} else {
			newRows = append(newRows, row)
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

	for {
		var size uint32
		err := binary.Read(db.SafetyFile, binary.LittleEndian, &size)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		payload := make([]byte, size)
		if _, err := io.ReadFull(db.SafetyFile, payload); err != nil {
			return err
		}

		decoded, _ := crypto.DecodeFromEmojis(string(payload))
		decrypted, _ := crypto.Decrypt(decoded, db.Key)

		var backup SafetyBackup
		json.Unmarshal(decrypted, &backup)

		if backup.Timestamp.Equal(timestamp) {
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

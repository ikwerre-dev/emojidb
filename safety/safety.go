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

type SafetyBackup struct {
	Timestamp time.Time
	TableName string
	Data      core.Row
}

func BackupForSafety(db *core.Database, tableName string, row core.Row) error {
	backup := SafetyBackup{
		Timestamp: time.Now(),
		TableName: tableName,
		Data:      row,
	}

	data, err := json.Marshal(backup)
	if err != nil {
		return err
	}

	encrypted, err := crypto.Encrypt(data, db.Key)
	if err != nil {
		return err
	}
	emojiPayload := crypto.EncodeToEmojis(encrypted)
	payloadBytes := []byte(emojiPayload)

	db.Mu.Lock()
	defer db.Mu.Unlock()

	_, err = db.SafetyFile.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	if err := binary.Write(db.SafetyFile, binary.LittleEndian, uint32(len(payloadBytes))); err != nil {
		return err
	}
	if _, err := db.SafetyFile.Write(payloadBytes); err != nil {
		return err
	}

	return db.SafetyFile.Sync()
}

func ListRecoveryPoints(db *core.Database) ([]time.Time, error) {
	db.Mu.RLock()
	defer db.Mu.RUnlock()

	var points []time.Time
	_, err := db.SafetyFile.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	for {
		var size uint32
		err := binary.Read(db.SafetyFile, binary.LittleEndian, &size)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		payload := make([]byte, size)
		if _, err := io.ReadFull(db.SafetyFile, payload); err != nil {
			return nil, err
		}

		decoded, err := crypto.DecodeFromEmojis(string(payload))
		if err != nil {
			continue
		}
		decrypted, err := crypto.Decrypt(decoded, db.Key)
		if err != nil {
			continue
		}

		var backup SafetyBackup
		if err := json.Unmarshal(decrypted, &backup); err != nil {
			continue
		}

		if time.Since(backup.Timestamp) <= 31*time.Minute {
			points = append(points, backup.Timestamp)
		}
	}

	return points, nil
}

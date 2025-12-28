package storage

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"
)

type Config struct {
	Encrypt bool
}

type Header struct {
	Magic   [5]byte
	Version uint32
}

func WriteHeader(file *os.File) error {
	var h Header
	copy(h.Magic[:], "EMOJI")
	h.Version = 1
	return binary.Write(file, binary.LittleEndian, h)
}

func PersistClump(file *os.File, mu *sync.RWMutex, tableName string, clump interface{}, encrypt bool, key string, encryptFn func([]byte, string) ([]byte, error), encodeFn func([]byte) string) error {
	mu.Lock()
	defer mu.Unlock()

	_, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	data, err := json.Marshal(clump)
	if err != nil {
		return err
	}

	var finalData []byte
	isEncrypted := encrypt && key != ""

	if isEncrypted {
		encrypted, err := encryptFn(data, key)
		if err != nil {
			return err
		}
		emojiPayload := encodeFn(encrypted)
		finalData = []byte(emojiPayload)
	} else {
		finalData = data
	}

	tbNameBytes := []byte(tableName)
	if err := binary.Write(file, binary.LittleEndian, uint32(len(tbNameBytes))); err != nil {
		return err
	}
	if _, err := file.Write(tbNameBytes); err != nil {
		return err
	}

	var encFlag uint8
	if isEncrypted {
		encFlag = 1
	}
	if err := binary.Write(file, binary.LittleEndian, encFlag); err != nil {
		return err
	}

	if err := binary.Write(file, binary.LittleEndian, uint32(len(finalData))); err != nil {
		return err
	}
	if _, err := file.Write(finalData); err != nil {
		return err
	}

	return file.Sync()
}

func Load(file *os.File, mu *sync.RWMutex, key string, decryptFn func([]byte, string) ([]byte, error), decodeFn func(string) ([]byte, error), handleClump func(string, []byte) error) error {
	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	var h Header
	err = binary.Read(file, binary.LittleEndian, &h)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return WriteHeader(file)
		}
		return err
	}

	if string(h.Magic[:]) != "EMOJI" {
		return errors.New("invalid database file format")
	}

	for {
		var nameLen uint32
		err := binary.Read(file, binary.LittleEndian, &nameLen)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		nameBytes := make([]byte, nameLen)
		if _, err := io.ReadFull(file, nameBytes); err != nil {
			return err
		}
		tableName := string(nameBytes)

		var encFlag uint8
		if err := binary.Read(file, binary.LittleEndian, &encFlag); err != nil {
			return err
		}

		var dataLen uint32
		if err := binary.Read(file, binary.LittleEndian, &dataLen); err != nil {
			return err
		}

		data := make([]byte, dataLen)
		if _, err := io.ReadFull(file, data); err != nil {
			return err
		}

		var finalData []byte
		if encFlag == 1 {
			if key == "" {
				return errors.New("database is encrypted but no key provided")
			}
			encrypted, err := decodeFn(string(data))
			if err != nil {
				return err
			}
			decrypted, err := decryptFn(encrypted, key)
			if err != nil {
				return err
			}
			finalData = decrypted
		} else {
			finalData = data
		}

		if err := handleClump(tableName, finalData); err != nil {
			return err
		}
	}
	return nil
}

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
	"sort"
	"strings"
)

var EmojiAlphabet = []string{
	"😀", "😁", "😂", "🤣", "😃", "😄", "😅", "😆", "😉", "😊", "😋", "😎", "😍", "😘", "🥰", "😗",
	"😙", "😚", "☺️", "🙂", "🤗", "🤩", "🤔", "🤨", "😐", "😑", "😶", "🙄", "😏", "😣", "😥", "😮",
	"🤐", "😯", "😪", "😫", "😴", "😌", "😛", "😜", "😝", "🤤", "😒", "😓", "😔", "😕", "🙃", "🤑",
	"😲", "☹️", "🙁", "😖", "😞", "😟", "😤", "😢", "😭", "😦", "😧", "😨", "😩", "🤯", "😬", "😰",
	"😱", "🥵", "🥶", "😳", "🤪", "😵", "😡", "😠", "🤬", "😷", "🤒", "🤕", "🤢", "🤮", "🤧", "😇",
	"🤠", "🤡", "🥳", "🥴", "🥺", "🤥", "🤫", "🤭", "🧐", "🤓", "😈", "👿", "👹", "👺", "💀", "👻",
	"👽", "🤖", "💩", "😺", "😸", "😹", "😻", "😼", "😽", "🙀", "😿", "😾", "🙈", "🙉", "🙊", "💋",
	"💌", "💘", "💝", "💖", "💗", "💓", "💞", "💕", "💟", "❣️", "💔", "❤️", "🧡", "💛", "💚", "💙",
	"💜", "🤎", "🖤", "🤍", "💯", "💢", "💥", "💫", "💦", "💨", "🕳️", "💣", "💬", "👁️‍🗨️", "🗨️", "🗯️",
	"💭", "💤", "👋", "🤚", "🖐️", "✋", "🖖", "👌", "🤏", "✌️", "🤞", "🤟", "🤘", "🤙", "👈", "👉",
	"👆", "🖕", "👇", "☝️", "👍", "👎", "✊", "👊", "🤛", "🤜", "👏", "🙌", "👐", "🤲", "🤝", "🙏",
	"✍️", "💅", "🤳", "💪", "🦾", "🦵", "🦿", "👣", "👂", "🦻", "👃", "🧠", "🦷", "🦴", "👀", "👁️",
	"👅", "👄", "👶", "🧒", "👦", "👧", "🧑", "👱", "👨", "🧔", "👩", "🧓", "👴", "👵", "👨‍⚕️", "👩‍⚕️",
	"👨‍🎓", "👩‍🎓", "👨‍🏫", "👩‍🏫", "👨‍⚖️", "👩‍⚖️", "👨‍🌾", "👩‍🌾", "👨‍🍳", "👩‍🍳", "👨‍🔧", "👩‍🔧", "👨‍🏭", "👩‍🏭", "👨‍💼", "👩‍💼",
	"👨‍🔬", "👩‍🔬", "👨‍💻", "👩‍💻", "👨‍🎤", "👩‍🎤", "👨‍🎨", "👩‍🎨", "👨‍✈️", "👩‍✈️", "👨‍🚀", "👩‍🚀", "👨‍🚒", "👩‍🚒", "👮", "🕵️",
	"💂", "👷", "🤴", "👸", "👳", "👲", "🧕", "🤵", "👰", "🤰", "🤱", "👼", "🎅", "🤶", "🦸", "🦹",
}

func DeriveKey(key string) []byte {
	hash := sha256.Sum256([]byte(key))
	return hash[:]
}

func Encrypt(data []byte, key string) ([]byte, error) {
	block, err := aes.NewCipher(DeriveKey(key))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func Decrypt(ciphertext []byte, key string) ([]byte, error) {
	block, err := aes.NewCipher(DeriveKey(key))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func EncodeToEmojis(data []byte) string {
	var sb strings.Builder
	for _, b := range data {
		sb.WriteString(EmojiAlphabet[int(b)%len(EmojiAlphabet)])
	}
	return sb.String()
}

func DecodeFromEmojis(s string) ([]byte, error) {
	type eData struct {
		s    string
		base byte
	}
	var sorted []eData
	for i, e := range EmojiAlphabet {
		if i < 256 {
			sorted = append(sorted, eData{e, byte(i)})
		}
	}

	sort.Slice(sorted, func(i, j int) bool {
		return len(sorted[i].s) > len(sorted[j].s)
	})

	var result []byte
	remaining := s
	for len(remaining) > 0 {
		found := false
		for _, ed := range sorted {
			if strings.HasPrefix(remaining, ed.s) {
				result = append(result, ed.base)
				remaining = remaining[len(ed.s):]
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("invalid emoji in payload")
		}
	}
	return result, nil
}

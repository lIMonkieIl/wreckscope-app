package wf2

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Capture is newline-delimited raw UDP payloads with receipt timestamps.
// Use recordings only with consent: packets can include player usernames.
type Capture struct {
	At  int64  `json:"at"`
	UDP string `json:"udp"`
}

func WriteCapture(w io.Writer, payload []byte, at time.Time) error {
	return json.NewEncoder(w).Encode(Capture{At: at.UnixNano(), UDP: base64.StdEncoding.EncodeToString(payload)})
}
func Replay(path string, speed float64, receive func([]byte)) error {
	if speed <= 0 || speed > 100 {
		return fmt.Errorf("replay speed must be >0 and <=100")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 4096), 256*1024)
	var previous int64
	for sc.Scan() {
		var row Capture
		if err := json.Unmarshal(sc.Bytes(), &row); err != nil {
			return err
		}
		data, err := base64.StdEncoding.DecodeString(row.UDP)
		if err != nil {
			return err
		}
		if previous > 0 && row.At > previous {
			pause := time.Duration(float64(row.At-previous) / speed)
			if pause > 2*time.Second {
				pause = 2 * time.Second
			}
			time.Sleep(pause)
		}
		previous = row.At
		receive(data)
	}
	return sc.Err()
}

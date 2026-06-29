package results

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Writer struct {
	used    *os.File
	found   *os.File
	muUsed  sync.Mutex
	muFound sync.Mutex
}

func NewWriter(dir string) (*Writer, error) {
	open := func(name string) (*os.File, error) {
		return os.OpenFile(filepath.Join(dir, name),
			os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	}

	used, err := open("used.txt")
	if err != nil {
		return nil, err
	}
	found, err := open("found.txt")
	if err != nil {
		return nil, err
	}

	return &Writer{used: used, found: found}, nil
}

func (w *Writer) Write(mnemonic, address string, balanceSat, txCount int64) {
	if balanceSat == 0 && txCount == 0 {
		return
	}

	line := fmt.Sprintf("%s | %s | %d sat | %d tx\n", mnemonic, address, balanceSat, txCount)

	if balanceSat > 0 {
		w.muFound.Lock()
		w.found.WriteString(line)
		w.muFound.Unlock()
	} else {
		w.muUsed.Lock()
		w.used.WriteString(line)
		w.muUsed.Unlock()
	}
}

func (w *Writer) Close() {
	w.used.Close()
	w.found.Close()
}

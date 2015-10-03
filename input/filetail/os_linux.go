package filetail

import (
	"log"
	"syscall"
)

func ctimeFile(filename string) int64 {
	var st syscall.Stat_t
	if err := syscall.Stat(filename, &st); err != nil {
		log.Printf("[DEBUG] [%s] Can't stat file %s, error: ", module, filename, err)
		return -1
	}
	return st.Ctim.Sec
}

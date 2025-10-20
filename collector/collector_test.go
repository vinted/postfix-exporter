package collector

import (
	"os"
	"path"
	"testing"
)

func TestDirectoryWalk(t *testing.T) {
	dir, err := os.MkdirTemp("", "maildrop")
	if err != nil {
		t.Errorf("%v", err)
	}
	defer os.RemoveAll(dir)

	file, err := os.Create(path.Join(dir, "test_mail_file"))
	if err != nil {
		t.Errorf("%v", err)
	}
	defer file.Close()

	_, err = file.Write([]byte("000"))
	if err != nil {
		t.Errorf("%v", err)
	}

	maildropCount, maildropSize := DirectoryWalk("", dir)
	if maildropCount != 1 {
		t.Errorf("Wrong file count. Wanted 1. Got %v", maildropCount)
	}
	if maildropSize != 3 {
		t.Errorf("Wrong size. Wanted 3. Got %v", maildropSize)
	}
}

package collector

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestDirectoryWalk(t *testing.T) {
	dir, _ := ioutil.TempDir("", "maildrop")
	defer os.RemoveAll(dir)
	file, _ := ioutil.TempFile(dir, "test_mail_file")
	_, err := file.Write([]byte("000"))
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

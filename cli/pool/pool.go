package pool

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
)

func Writefile(filename string,nei string) error{
	fileName := filename
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println("file create failed. err: " + err.Error())
	} else {

		content := nei
		n, _ := f.Seek(0, os.SEEK_END)
		_, err = f.WriteAt([]byte(content), n)

		defer f.Close()
	}
	return err
}

func Readfile(filename string)string{
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("read fail", err)
		os.Exit(-1)
	}
	return  string(f)
}

func GetmasterAddress()string{
	add := os.Getenv("spmadd")
	return add
}

func Getmasterspid()string{
	add := os.Getenv("spid")
	return add
}


func calculateMD5(input string) string {
	hash := md5.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}

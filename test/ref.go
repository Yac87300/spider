package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

var cvm1,cvm2 sync.Map


func main(){
	file, err := os.Open("cvm1")
	if err != nil {
		fmt.Println("rule err:",err)
		os.Exit(-1)
	}
	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		if fileScanner.Text() == ""{
			continue
		}
		lip := strings.Split(fileScanner.Text(),"\t")
		//sdsjzx-jkm-api3	16	32	100	500	172.20.188.12
		_,cpu,mem,disk1,disk2,ip := lip[0],lip[1],lip[2],lip[3],lip[4],lip[5]
		keyname := ip
		varof :=  cpu  + ";" + mem + ";" + disk1 + ";" + disk2 + ";"
		cvm1.Store(keyname,varof)
	}
	cache2()

	check()


}


func cache2(){
	file, err := os.Open("cvm2")
	if err != nil {
		fmt.Println("rule err:",err)
		os.Exit(-1)
	}
	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		if fileScanner.Text() == ""{
			continue
		}
		lip := strings.Split(fileScanner.Text(),"\t")
		_,cpu,mem,disk1,disk2,ip := lip[0],lip[1],lip[2],lip[3],lip[4],lip[5]
		keyname := ip
		varof :=   cpu  + ";" + mem + ";" + disk1 + ";" + disk2 + ";"
		cvm2.Store(keyname,varof)
	}
}

func check(){
	cvm2.Range(func(key, value interface{})bool{
		varof,ok := cvm1.Load(key.(string))
		if ok{
			varof2,_ := cvm2.Load(key.(string))
			if varof != varof2 {
            fmt.Println(key.(string),varof2,varof)
			}else {

			}
		}else {
			//fmt.Println(key.(string),"not find in cvm2",varof)
		}
		return true
	})
}
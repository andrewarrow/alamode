package main

import "io/ioutil"
import "runtime"
import "os"
import "strings"

func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func ReadList(name string) []string {
	data, _ := ioutil.ReadFile(UserHomeDir() + "/.mql_" + name)
	return strings.Split(string(data), ",")
}
func SaveList(name string, list []string) {
	ioutil.WriteFile(UserHomeDir()+"/.mql_"+name, []byte(strings.Join(list, ",")), 0644)
}

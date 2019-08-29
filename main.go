package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kirsle/configdir"
)

type response struct {
	ip       string
	hostname string
	city     string
	region   string
	country  string
	loc      string
	org      string
}

var authToken string

func showHelp() {
	fmt.Println("")
	fmt.Println(`
	ipinfo
	------

	usage:
	$ ipinfo <ip address>
	`)
}

func loadKey() {
	confPath := configdir.LocalConfig("ipinfo")
	if err := configdir.MakePath(confPath); err != nil {
		panic(err)
	}

	confFile := filepath.Join(confPath, "config")
	if _, err := os.Stat(confFile); os.IsNotExist(err) {
		fh, err := os.Create(confFile)
		if err != nil {
			fmt.Printf("Could not create config file: %s\n", err.Error())
			return
		}

		fh.Write([]byte(""))
		fmt.Printf("Please put your ipinfo.io token to %s\n", confFile)
		fh.Close()
		return
	}
	token := make([]byte, 24)

	fh, err := os.OpenFile(confFile, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Printf("Could not open config file: %s\n", err.Error())
		return
	}

	defer fh.Close()
	_, err = fh.Read(token)
	if err != nil {
		fmt.Printf("Could not read config: %s\n", err.Error())
		return
	}

	if len(token) < 1 {
		fmt.Println("Invalid token")
		return
	}

	authToken = strings.ReplaceAll(
		strings.ReplaceAll(
			strings.TrimSpace(string(token)),
			"\n", "",
		),
		"\x00", "",
	)
}

func queryIP(ip string) {
	ipAddr := net.ParseIP(ip).String()
	if len(ipAddr) < 1 {
		fmt.Println("Invalid IP address")
		return
	}

	target, _ := url.Parse(fmt.Sprintf("https://ipinfo.io/%s", ipAddr))
	q := target.Query()
	q.Set("token", authToken)
	target.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", target.String(), nil)
	if err != nil {
		fmt.Println("error building request")
		fmt.Println(err.Error())
		return
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(fmt.Sprintf(
			"error sending http request to %s: %s", target.String(), err.Error(),
		))
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error reading response")
		return
	}

	fmt.Println(string(body))
}

func main() {
	loadKey()

	if len(os.Args) != 2 {
		showHelp()
		return
	}

	ip := os.Args[1]
	queryIP(ip)
}

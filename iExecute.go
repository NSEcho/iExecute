package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Rport    string `yaml:"rport"`
	Username string `yaml:"username"`
}

func getHome() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	return home, nil
}

func parseConfig() (*Config, error) {
	cfg := &Config{}

	home, err := getHome()
	if err != nil {
		return nil, err
	}

	filePath := path.Join(home, ".iExecute")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getRandomPort() string {
	n := 50000 + rand.Intn(10000)
	return strconv.Itoa(n)
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("You need to pass the command")
		os.Exit(1)
	}

	rand.Seed(time.Now().UTC().UnixNano())

	lport := getRandomPort()

	cfg, err := parseConfig()
	if err != nil {
		fmt.Printf("Error parsing config: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("/usr/local/bin/iproxy", lport, cfg.Rport)

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error spawning iproxy: %v\n", err)
	}

	time.Sleep(time.Second)

	var wg sync.WaitGroup
	ch := make(chan string)

	wg.Add(1)
	go func() {
		wg.Wait()
		close(ch)
	}()

	go run(cfg.Username, "localhost", lport, os.Args[1], &wg, ch)

	for out := range ch {
		fmt.Println(out)
	}

	cmd.Process.Kill()
}

func run(user, address, port, command string, wg *sync.WaitGroup, out chan string) {

	defer wg.Done()

	home, err := getHome()
	if err != nil {
		out <- fmt.Sprintf("Error getting home: %v\n", err)
		return
	}

	keyPath := path.Join(home, ".ssh/id_rsa")
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		out <- fmt.Sprintf("Error reading private key: %v\n", err)
		return
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		out <- fmt.Sprintf("Error parsing private key: %v\n", err)
		return
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	hostToConnect := fmt.Sprintf("%s:%s", address, port)

	client, err := ssh.Dial("tcp", hostToConnect, config)
	if err != nil {
		out <- fmt.Sprintf("Error connecting: %v\n", err)
		return
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		out <- fmt.Sprintf("Error creating session: %v\n", err)
		return
	}
	defer session.Close()

	var buf bytes.Buffer
	session.Stdout = &buf
	session.Run(command)

	out <- strings.TrimSpace(buf.String())
}

package main

import (
	"flag"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func UploadFiles(clients []*ssh.Client, source, destination string) {
	// TODO: Super gross, do something about this
	// TODO: Currently only support src depth of one
	// Get absolute paths of source and dest
	absSource, _ := filepath.Abs(source)
	absDestination, _ := filepath.Abs(destination)

	// Get sftp connections for putting files to remote
	var sftpClients []*sftp.Client
	for _, client := range clients {
		log.Println(fmt.Sprintf("[%s] Copying %s to %s", client.RemoteAddr(), absSource, absDestination))
		// open an SFTP session over an existing ssh connection.
		sftp, err := sftp.NewClient(client)
		if err != nil {
			log.Fatal("Failed to open sftp session: " + err.Error())
		}
		defer sftp.Close()
		sftpClients = append(sftpClients, sftp)

		// Create dest paths if they do not exist
		dirToCreate := ""
		for _, dir := range strings.Split(absDestination, "/") {
			dirToCreate += "/" + dir
			sftp.Mkdir(dirToCreate)
		}
	}

	// Walk source and put files on remote
	filepath.Walk(absSource, func(path string, info os.FileInfo, _ error) error {
		sourceStep, err := filepath.Rel(absSource, path)
		if err != nil {
			log.Fatal(err)
		}
		if sourceStep == "." {
			return nil
		}

		destFile := filepath.Join(absDestination, info.Name())

		if !info.IsDir() {
			for _, sftp := range sftpClients {
				f, err := sftp.Create(destFile)
				if err != nil {
					log.Fatal("Failed to create remote file: " + err.Error())
				}
				sourceFileContents, err := ioutil.ReadFile(path)
				if err != nil {
					log.Fatal("Failed to open source file: " + err.Error())
				}
				if _, err := f.Write(sourceFileContents); err != nil {
					log.Fatal("Failed to write remote file: " + err.Error())
				}
			}
		}

		return nil
	})
}

func RemoteRun(clients []*ssh.Client, command string) {
	for _, client := range clients {
		session, err := client.NewSession()
		if err != nil {
			log.Fatal("Failed to create session: " + err.Error())
		}
		defer session.Close()

		log.Println(fmt.Sprintf("[%s] %s", client.RemoteAddr(), command))
		if err := session.Run(command); err != nil {
			log.Fatal("Failed to run: " + err.Error())
		}

	}
}

func GetSshClient(host, username, password string) *ssh.Client {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}
	client, err := ssh.Dial("tcp", host+":22", config)
	if err != nil {
		log.Fatal("Failed to dial: " + err.Error())
	}
	return client
}

func GetSshClients(hosts []string) []*ssh.Client {
	var clients []*ssh.Client
	for _, host := range hosts {
		clients = append(clients, GetSshClient(host, username, password))
	}
	return clients
}

func ParseHostsFile(hostsFile string) []string {
	filename, _ := filepath.Abs(hostsFile)
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Failed to open hosts file: " + err.Error())
	}

	var hosts []string
	err = yaml.Unmarshal(yamlFile, &hosts)
	if err != nil {
		log.Fatal("Failed to parse hosts file: " + err.Error())
	}
	return hosts
}

type Config struct {
	Type        string
	Command     string
	Source      string
	Destination string
}

func ParseConfigFile(configFile string) []Config {
	filename, _ := filepath.Abs(configFile)
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Failed to open config file: " + err.Error())
	}

	var configs []Config
	err = yaml.Unmarshal(yamlFile, &configs)
	if err != nil {
		log.Fatal("Failed to parse config file: " + err.Error())
	}
	return configs
}

var hostsFile, username, password string

func init() {
	const hostsUsage = "yaml file of hosts"
	flag.StringVar(&hostsFile, "hosts", "", hostsUsage)
	flag.StringVar(&hostsFile, "h", "", hostsUsage+" (shorthand)")

	const (
		usernameUsage = "username for hosts"
		defaultUser   = "root"
	)
	flag.StringVar(&username, "username", defaultUser, usernameUsage)
	flag.StringVar(&username, "u", defaultUser, usernameUsage+" (shorthand)")

	const passwordUsage = "password for hosts"
	flag.StringVar(&password, "password", "", passwordUsage)
	flag.StringVar(&password, "p", "", passwordUsage+" (shorthand)")
}

func main() {
	flag.Parse()
	if username == "" {
		log.Fatal("Username required.")
	}

	hosts := ParseHostsFile(hostsFile)
	clients := GetSshClients(hosts)

	configFile := flag.Arg(0)
	configs := ParseConfigFile(configFile)

	for _, config := range configs {
		if config.Type == "copy-files" {
			UploadFiles(clients, config.Source, config.Destination)
		}
		if config.Type == "remote-command" {
			RemoteRun(clients, config.Command)
		}
	}
}

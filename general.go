package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"locallibs/support"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// Config struct
type Config struct {
	Etcd struct {
		// var name has to be uppercase here or it won't work
		Endpoints      []string `yaml:"endpoints"`
		BaseKeyToWrite string   `yaml:"baseKeyToWrite"`
		Timeout        int      `yaml:"timeout"`
		CertCa         string   `yaml:"cert-ca"`
		PeerCert       string   `yaml:"peer-cert"`
		PeerKey        string   `yaml:"peer-key"`
	}
}

// GetConfigContents Unmarshals the config contents from file into memory
func GetConfigContents(filename string) Config {
	var conf Config
	file := support.ReadConfigFileContents(filename)
	err := yaml.Unmarshal(file, &conf)
	if err != nil {
		fmt.Printf("There was an error decoding the yaml file. err = %s\n", err)
	}

	return conf
}

// GenerateID creates a random string 12 digits long and returns it to server as an id
func GenerateID() string {
	bytes := make([]byte, 12)
	rand.Read(bytes)
	for i := range bytes {
		n := int(bytes[i]) // convert to int
		/* now mod our number by 36 so we get a result bewteen 1 and 35, anything less than
		   10 is assigned a digit, and anything greater is assigned a lowercase letter.
		   48 decimal is 0x30 or ascii '0' and 97 decimal is 0x61 or ascii 'a'. */
		n = (n % 36)
		if n < 10 {
			n += 0x30
		} else {
			n += 0x57 // ten less than 0x61 to account for shift
		}
		bytes[i] = byte(n) // now convert back to byte
	}

	return string(bytes)
}

// TakeUserInput gets input from the user
func TakeUserInput() string {
	fmt.Println("Enter a message: ")
	in := bufio.NewReader(os.Stdin)
	msg, err := in.ReadString('\n')
	if err != nil {
		fmt.Printf("There was an error taking in input: %s\n", err)
	}

	return msg
}

// GetMicroTime returns a string of the microsecond level of right now (at runtime)
func GetMicroTime() string {
	now := time.Now()
	tstamp := now.Format(time.RFC3339Nano)

	return tstamp[:len(tstamp)-7] // microsecond resolution
}

// ReadEtcdContinuously continuously prints read variables to screen except the ones we wrote
func ReadEtcdContinuously(readch chan string, config Config, keyWritten string) {
	for true {
		values := ReadFromEtcd(config, config.Etcd.BaseKeyToWrite)

		// remove the value we wrote ourselves
		if _, ok := values[keyWritten]; ok {
			delete(values, keyWritten)
		}

		// now put them in a string for printing
		msg := ""
		for k, v := range values {
			msg += k + " :\t"
			msg += v + "\n"
		}

		// then delete printed key from etcd
		for keyToDelete := range values {
			DeleteFromEtcd(config, keyToDelete)
		}

		// and send into channel
		readch <- msg
	}
}

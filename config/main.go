package config

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
)

type ConfigData struct {
	Sk ed25519.PrivateKey // private key
}

func ReadConfig(path string) (config ConfigData, err error) {
	configInfo, findConfigErr := os.Stat(path)

	if os.IsNotExist(findConfigErr) {
		return config, errors.New(fmt.Sprintf("The provided configuration path %s does not exist.\n", path))
	} else if !configInfo.IsDir() {
		return config, errors.New(fmt.Sprintf("The provided configuration path %s should be a folder.\n", path))
	}

	entries, err := os.ReadDir(path)

	var hasprivatekey bool = false

	// TODO: This for loop is unnecessary
	for _, entry := range entries {
		switch entry.Name() {
		case "private.pem":
			hasprivatekey = true
		}
	}

	if !hasprivatekey {
		return config, errors.New(fmt.Sprintf("The private.pem file is missing from the configuration.\n"))
	}

	// read private key
	readprivate, readprivateerr := os.ReadFile(path + "/private.pem")
	if readprivateerr != nil {
		return config, errors.New(fmt.Sprintf("Failed to read the private.pem file.\n"))
	}

	privateblock, _ := pem.Decode(readprivate)
	secretkey, parseErr := x509.ParsePKCS8PrivateKey(privateblock.Bytes)
	if parseErr != nil {
		return config, errors.New(fmt.Sprintf("Failed to parse the private.pem key.\n"))
	}

	return ConfigData{
		Sk: secretkey.(ed25519.PrivateKey),
	}, nil
}

func WriteConfig(folder string) (config ConfigData) {
	os.MkdirAll(folder, os.ModePerm)

	_, secretkey, err := ed25519.GenerateKey(rand.Reader)
	// generate key
	if err != nil {
		log.Printf("Cannot generate private and public keys\n")
		os.Exit(1)
	}

	// dump private key to file
	var secretKeyBytes []byte
	secretKeyBytes, err = x509.MarshalPKCS8PrivateKey(secretkey)
	secretKeyBlock := &pem.Block{
		Type:  "ED-25519 PRIVATE KEY",
		Bytes: secretKeyBytes,
	}
	secretPem, err := os.Create(folder + "/private.pem")
	if err != nil {
		log.Printf("error when create private.pem: %s \n", err)
		os.Exit(1)
	}
	err = pem.Encode(secretPem, secretKeyBlock)
	if err != nil {
		log.Printf("error when encode private pem: %s \n", err)
		os.Exit(1)
	}

	return ConfigData{
		Sk: secretkey,
	}
}

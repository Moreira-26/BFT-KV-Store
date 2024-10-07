package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

type ConfigData struct {
	Pk *rsa.PublicKey  // public key
	Sk *rsa.PrivateKey // private key
}

func ReadConfig(path string) (config ConfigData, err error) {
	configInfo, findConfigErr := os.Stat(path)

	if os.IsNotExist(findConfigErr) {
		return config, errors.New(fmt.Sprintf("The provided configuration path %s does not exist.\n", path))
	} else if !configInfo.IsDir() {
		return config, errors.New(fmt.Sprintf("The provided configuration path %s should be a folder.\n", path))
	}

	entries, err := os.ReadDir(path)

	var haspublickey bool = false
	var hasprivatekey bool = false

	for _, entry := range entries {
		switch entry.Name() {
		case "private.pem":
			hasprivatekey = true
		case "public.pem":
			haspublickey = true
		}
	}

	if !hasprivatekey && !haspublickey {
		return config, errors.New(fmt.Sprintf("The configuration file should have the private.pem and public.pem files.\n"))
	} else if !hasprivatekey {
		return config, errors.New(fmt.Sprintf("The private.pem file is missing from the configuration.\n"))
	} else if !haspublickey {
		return config, errors.New(fmt.Sprintf("The public.pem file is missing from the configuration.\n"))
	}

	// read private key
	readprivate, readprivateerr := os.ReadFile(path + "/private.pem")
	if readprivateerr != nil {
		return config, errors.New(fmt.Sprintf("Failed to read the private.pem file.\n"))
	}

	privateblock, _ := pem.Decode(readprivate)
	secretkey, parseErr := x509.ParsePKCS1PrivateKey(privateblock.Bytes)
	if parseErr != nil {
		return config, errors.New(fmt.Sprintf("Failed to parse the private.pem key.\n"))
	}

	// read public key
	readpublic, readpublicerr := os.ReadFile(path + "/public.pem")
	if readpublicerr != nil {
		return config, errors.New(fmt.Sprintf("Failed to read the public.pem file.\n"))
	}

	publicblock, _ := pem.Decode(readpublic)
	publickey, parseErr := x509.ParsePKIXPublicKey(publicblock.Bytes)
	if parseErr != nil {
		return config, errors.New(fmt.Sprintf("Failed to parse the public.pem key.\n"))
	}

	return ConfigData{
		Pk: publickey.(*rsa.PublicKey),
		Sk: secretkey,
	}, nil
}

func WriteConfig(folder string) (config ConfigData) {
	os.MkdirAll(folder, os.ModePerm)

	// generate key
	secretkey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("Cannot generate RSA key\n")
		os.Exit(1)
	}
	publickey := &secretkey.PublicKey

	// dump private key to file
	var secretKeyBytes []byte = x509.MarshalPKCS1PrivateKey(secretkey)
	secretKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: secretKeyBytes,
	}
	secretPem, err := os.Create(folder + "/private.pem")
	if err != nil {
		fmt.Printf("error when create private.pem: %s \n", err)
		os.Exit(1)
	}
	err = pem.Encode(secretPem, secretKeyBlock)
	if err != nil {
		fmt.Printf("error when encode private pem: %s \n", err)
		os.Exit(1)
	}

	// dump public key to file
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publickey)
	if err != nil {
		fmt.Printf("error when dumping publickey: %s \n", err)
		os.Exit(1)
	}
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	publicPem, err := os.Create(folder + "/public.pem")
	if err != nil {
		fmt.Printf("error when create public.pem: %s \n", err)
		os.Exit(1)
	}
	err = pem.Encode(publicPem, publicKeyBlock)
	if err != nil {
		fmt.Printf("error when encode public pem: %s \n", err)
		os.Exit(1)
	}

	return ConfigData{
		Pk: publickey,
		Sk: secretkey,
	}
}

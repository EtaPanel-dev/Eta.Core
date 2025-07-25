package ssl

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/models/ssl"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"net/http"
	"time"
)

func NewConfigWithProxy(user registration.User, accountType, customCaURL string) *lego.Config {
	var (
		caDirURL string
	)
	caDirURL = GetCaDirURL(accountType, customCaURL)
	client := &http.Client{Timeout: 10 * time.Second}

	return &lego.Config{
		CADirURL:   caDirURL,
		UserAgent:  "1Panel",
		User:       user,
		HTTPClient: client,
		Certificate: lego.CertificateConfig{
			KeyType: certcrypto.RSA2048,
			Timeout: 60 * time.Second,
		},
	}
}

func NewRegisterClient(acmeAccount *ssl.WebsiteAcmeAccount) (*ssl.AcmeClient, error) {
	var (
		priKey crypto.PrivateKey
		err    error
	)

	const (
		KeyEC256   = certcrypto.EC256
		KeyEC384   = certcrypto.EC384
		KeyRSA2048 = certcrypto.RSA2048
		KeyRSA3072 = certcrypto.RSA3072
		KeyRSA4096 = certcrypto.RSA4096
	)

	type KeyType = certcrypto.KeyType

	if acmeAccount.PrivateKey != "" {
		switch KeyType(acmeAccount.KeyType) {
		case KeyEC256, KeyEC384:
			block, _ := pem.Decode([]byte(acmeAccount.PrivateKey))
			priKey, err = x509.ParseECPrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
		case KeyRSA2048, KeyRSA3072, KeyRSA4096:
			block, _ := pem.Decode([]byte(acmeAccount.PrivateKey))
			priKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
		}

	} else {
		priKey, err = certcrypto.GeneratePrivateKey(KeyType(acmeAccount.KeyType))
		if err != nil {
			return nil, err
		}
	}

	myUser := &ssl.AcmeUser{
		Email: acmeAccount.Email,
		Key:   priKey,
	}
	config := NewConfigWithProxy(myUser, acmeAccount.Type, acmeAccount.CaDirURL)
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}
	var reg *registration.Resource
	if acmeAccount.Type == "zerossl" || acmeAccount.Type == "google" || acmeAccount.Type == "freessl" {
		if acmeAccount.Type == "zerossl" {
			var res *ZeroSSLRes
			res, err = GetZeroSSLEabCredentials(acmeAccount.Email)
			if err != nil {
				return nil, err
			}
			if res.Success {
				acmeAccount.EabKid = res.EabKid
				acmeAccount.EabHmacKey = res.EabHmacKey
			} else {
				return nil, fmt.Errorf("get zero ssl eab credentials failed")
			}
		}

		eabOptions := registration.RegisterEABOptions{
			TermsOfServiceAgreed: true,
			Kid:                  acmeAccount.EabKid,
			HmacEncoded:          acmeAccount.EabHmacKey,
		}
		reg, err = client.Registration.RegisterWithExternalAccountBinding(eabOptions)
		if err != nil {
			return nil, err
		}
	} else {
		reg, err = client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return nil, err
		}
	}
	myUser.Registration = reg

	acmeClient := &ssl.AcmeClient{
		User:      myUser,
		Client:    client,
		Config:    config,
		ServerURL: acmeAccount.CaDirURL,
		KeyType:   acmeAccount.KeyType,
	}

	return acmeClient, nil
}

func NewAcmeClient(acmeAccount *ssl.WebsiteAcmeAccount) (*ssl.AcmeClient, error) {
	if acmeAccount.Email == "" {
		return nil, errors.New("email can not blank")
	}

	client, err := NewRegisterClient(acmeAccount)
	if err != nil {
		return nil, err
	}
	return client, nil
}

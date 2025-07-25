package ssl

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/handler"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/models/ssl"
	"github.com/gin-gonic/gin"
	"github.com/go-acme/lego/v4/certcrypto"
	"golang.org/x/crypto/acme"
	"os"
)

type CreateSSLRequest struct {
	AcmeClientID int    `json:"acme_client_id" binding:"required"`
	Domain       string `json:"domain" binding:"required"`
}

func IssueSSL(c *gin.Context) {
	var req CreateSSLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Respond(c, 400, "请求参数错误", nil)
		return
	}

	acmeClient, err := GetAcmeClient(req.AcmeClientID)
	if err != nil {
		handler.Respond(c, 404, "未找到 ACME 客户端", nil)
		return
	}

	acmeUser, err := GetUserById(int(acmeClient.User.ID))
	if err != nil {
		handler.Respond(c, 404, "未找到 ACME 用户", nil)
		return
	}

	if err := requestCertificate(req.Domain, acmeClient, acmeUser); err != nil {
		handler.Respond(c, 500, fmt.Sprintf("申请证书失败: %v", err), nil)
		return
	}

	handler.Respond(c, 200, "证书申请成功", nil)
}

func requestCertificate(domain string, acmeClient ssl.AcmeClient, acmeUser ssl.AcmeUser) error {
	client := &acme.Client{
		DirectoryURL: acmeClient.ServerURL,
	}

	account := &acme.Account{
		Contact: []string{"mailto:" + acmeUser.Email},
	}

	if _, err := client.Register(context.Background(), account, acme.AcceptTOS); err != nil {
		return err
	}

	// 创建订单
	order, err := client.AuthorizeOrder(context.Background(), acme.DomainIDs(domain))
	if err != nil {
		return fmt.Errorf("创建订单失败: %v", err)
	}

	// 完成所有授权挑战
	for _, authzURL := range order.AuthzURLs {
		authz, err := client.GetAuthorization(context.Background(), authzURL)
		if err != nil {
			return fmt.Errorf("获取授权失败: %v", err)
		}

		// 选择HTTP挑战
		var chal *acme.Challenge
		for _, c := range authz.Challenges {
			if c.Type == "http-01" {
				chal = c
				break
			}
		}

		if chal == nil {
			return fmt.Errorf("没有找到HTTP挑战")
		}

		// 接受挑战
		if _, err := client.Accept(context.Background(), chal); err != nil {
			return fmt.Errorf("接受挑战失败: %v", err)
		}

		// 等待授权完成
		if _, err := client.WaitAuthorization(context.Background(), authz.URI); err != nil {
			return fmt.Errorf("等待授权完成失败: %v", err)
		}
	}

	// 创建CSR
	privKey, err := certcrypto.GeneratePrivateKey(certcrypto.KeyType(acmeClient.KeyType))
	if err != nil {
		return fmt.Errorf("生成私钥失败: %v", err)
	}

	template := x509.CertificateRequest{
		Subject:  pkix.Name{CommonName: domain},
		DNSNames: []string{domain},
	}
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, &template, privKey)
	if err != nil {
		return fmt.Errorf("创建CSR失败: %v", err)
	}

	// 使用CreateOrderCert申请证书
	der, _, err := client.CreateOrderCert(context.Background(), order.FinalizeURL, csrDER, true)
	if err != nil {
		return fmt.Errorf("申请证书失败: %v", err)
	}

	// 保存证书和私钥
	if err := saveCertificateAndKey(domain, der, privKey); err != nil {
		return fmt.Errorf("保存证书失败: %v", err)
	}

	return nil
}

func saveCertificateAndKey(domain string, certs [][]byte, privKey crypto.PrivateKey) error {
	// 类型断言
	ecKey, ok := privKey.(*ecdsa.PrivateKey)
	if !ok {
		return fmt.Errorf("期望ECDSA私钥，但得到的是 %T", privKey)
	}

	// 保存私钥
	keyOut, err := os.OpenFile(domain+".key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	privBytes, err := x509.MarshalECPrivateKey(ecKey)
	if err != nil {
		return err
	}

	err = pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
	if err != nil {
		return err
	}

	return nil
}

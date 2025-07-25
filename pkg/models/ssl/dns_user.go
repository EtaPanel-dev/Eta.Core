package ssl

import (
	"encoding/json"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/providers/dns/baiducloud"
	"github.com/go-acme/lego/v4/providers/dns/clouddns"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/providers/dns/cloudns"
	"github.com/go-acme/lego/v4/providers/dns/dnspod"
	"github.com/go-acme/lego/v4/providers/dns/dynu"
	"github.com/go-acme/lego/v4/providers/dns/freemyip"
	"github.com/go-acme/lego/v4/providers/dns/godaddy"
	"github.com/go-acme/lego/v4/providers/dns/huaweicloud"
	"github.com/go-acme/lego/v4/providers/dns/namecheap"
	"github.com/go-acme/lego/v4/providers/dns/namedotcom"
	"github.com/go-acme/lego/v4/providers/dns/namesilo"
	"github.com/go-acme/lego/v4/providers/dns/rainyun"
	"github.com/go-acme/lego/v4/providers/dns/regru"
	"github.com/go-acme/lego/v4/providers/dns/spaceship"
	"github.com/go-acme/lego/v4/providers/dns/tencentcloud"
	"github.com/go-acme/lego/v4/providers/dns/vercel"
	"github.com/go-acme/lego/v4/providers/dns/volcengine"
	"github.com/go-acme/lego/v4/providers/dns/westcn"
	"gorm.io/gorm"
	"time"
)

const (
	AliyunDnsProvider      = 0
	QCloudDnsProvider      = 1
	HuaweiCloudDnsProvider = 2
	GoDaddyDnsProvider     = 3
	CloudflareDnsProvider  = 4
	VercelDnsProvider      = 5
	CloudDnsProvider       = 6
	NameSiloDnsProvider    = 7
	NameCheapDnsProvider   = 8
	NameConDns             = 9
	FreeMyIpDnsProvider    = 10
	RainYunDnsProvider     = 11
	WestDnsProvider        = 12
	ClouDnsProvider        = 13
	SpaceshipDnsProvider   = 14
	VolcEngineDnsProvider  = 15
	DnsPodDnsProvider      = 16
	RegRuDnsProvider       = 17
	DynuDnsProvider        = 18
	BaiduCloudDnsProvider  = 19
)

type DnsUser struct {
	gorm.Model
	ProviderId int    `json:"provider_id" gorm:"not null"`
	Key        string `json:"key" gorm:"not"`
	Value      string `json:"value" gorm:"not null"`
}

func (d *DnsUser) getProviderId() int {
	return d.ProviderId
}

func (d *DnsUser) getKey() string {
	return d.Key
}

func (d *DnsUser) getValue() string {
	return d.Value
}

type DnsType int

type DNSParam struct {
	ID           string `json:"id"`
	Token        string `json:"token"`
	AccessKey    string `json:"accessKey"`
	SecretKey    string `json:"secretKey"`
	Email        string `json:"email"`
	APIkey       string `json:"apiKey"`
	APIUser      string `json:"apiUser"`
	APISecret    string `json:"apiSecret"`
	SecretID     string `json:"secretID"`
	ClientID     string `json:"clientID"`
	Password     string `json:"password"`
	Region       string `json:"region"`
	Username     string `json:"username"`
	AuthID       string `json:"authID"`
	SubAuthID    string `json:"subAuthID"`
	AuthPassword string `json:"authPassword"`
}

var (
	propagationTimeout = 30 * time.Minute
	pollingInterval    = 10 * time.Second
	ttl                = 3600
	dnsTimeOut         = 30 * time.Minute
	manualDnsTimeout   = 10 * time.Minute
)

func getDNSProviderConfig(dnsType DnsType, params string) (challenge.Provider, error) {
	var (
		param DNSParam
		p     challenge.Provider
		err   error
	)
	if err := json.Unmarshal([]byte(params), &param); err != nil {
		return nil, err
	}
	switch dnsType {
	case DnsPodDnsProvider:
		config := dnspod.NewDefaultConfig()
		config.LoginToken = param.ID + "," + param.Token
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		config.TTL = ttl
		p, err = dnspod.NewDNSProviderConfig(config)
	case AliyunDnsProvider:
		config := alidns.NewDefaultConfig()
		config.SecretKey = param.SecretKey
		config.APIKey = param.AccessKey
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		config.TTL = ttl
		p, err = alidns.NewDNSProviderConfig(config)
	case CloudflareDnsProvider:
		config := cloudflare.NewDefaultConfig()
		config.AuthEmail = param.Email
		config.AuthToken = param.APIkey
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		config.TTL = ttl
		p, err = cloudflare.NewDNSProviderConfig(config)
	case CloudDnsProvider:
		config := clouddns.NewDefaultConfig()
		config.ClientID = param.ClientID
		config.Email = param.Email
		config.Password = param.Password
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		config.TTL = ttl
		p, err = clouddns.NewDNSProviderConfig(config)
	case NameCheapDnsProvider:
		config := namecheap.NewDefaultConfig()
		config.APIKey = param.APIkey
		config.APIUser = param.APIUser
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		config.TTL = ttl
		p, err = namecheap.NewDNSProviderConfig(config)
	case NameSiloDnsProvider:
		config := namesilo.NewDefaultConfig()
		config.APIKey = param.APIkey
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		config.TTL = ttl
		p, err = namesilo.NewDNSProviderConfig(config)
	case GoDaddyDnsProvider:
		config := godaddy.NewDefaultConfig()
		config.APIKey = param.APIkey
		config.APISecret = param.APISecret
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		config.TTL = ttl
		p, err = godaddy.NewDNSProviderConfig(config)
	case NameConDns:
		config := namedotcom.NewDefaultConfig()
		config.APIToken = param.Token
		config.Username = param.APIUser
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		config.TTL = ttl
		p, err = namedotcom.NewDNSProviderConfig(config)
	case QCloudDnsProvider:
		config := tencentcloud.NewDefaultConfig()
		config.SecretID = param.SecretID
		config.SecretKey = param.SecretKey
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		config.TTL = ttl
		p, err = tencentcloud.NewDNSProviderConfig(config)
	case RainYunDnsProvider:
		config := rainyun.NewDefaultConfig()
		config.APIKey = param.APIkey
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		config.TTL = ttl
		p, err = rainyun.NewDNSProviderConfig(config)
	case VolcEngineDnsProvider:
		config := volcengine.NewDefaultConfig()
		config.SecretKey = param.SecretKey
		config.AccessKey = param.AccessKey
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		config.TTL = ttl
		p, err = volcengine.NewDNSProviderConfig(config)
	case HuaweiCloudDnsProvider:
		config := huaweicloud.NewDefaultConfig()
		config.AccessKeyID = param.AccessKey
		config.SecretAccessKey = param.SecretKey
		config.Region = param.Region
		if config.Region == "" {
			config.Region = "cn-north-1"
		}
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		config.TTL = int32(ttl)
		p, err = huaweicloud.NewDNSProviderConfig(config)
	case FreeMyIpDnsProvider:
		config := freemyip.NewDefaultConfig()
		config.Token = param.Token
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = freemyip.NewDNSProviderConfig(config)
	case VercelDnsProvider:
		config := vercel.NewDefaultConfig()
		config.AuthToken = param.Token
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = vercel.NewDNSProviderConfig(config)
	case SpaceshipDnsProvider:
		config := spaceship.NewDefaultConfig()
		config.APIKey = param.APIkey
		config.APISecret = param.APISecret
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		config.TTL = ttl
		p, err = spaceship.NewDNSProviderConfig(config)
	case WestDnsProvider:
		config := westcn.NewDefaultConfig()
		config.Username = param.Username
		config.Password = param.Password
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		config.TTL = ttl
		p, err = westcn.NewDNSProviderConfig(config)
	case ClouDnsProvider:
		config := cloudns.NewDefaultConfig()
		config.AuthID = param.AuthID
		config.SubAuthID = param.SubAuthID
		config.AuthPassword = param.AuthPassword
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		config.TTL = ttl
		p, err = cloudns.NewDNSProviderConfig(config)
	case RegRuDnsProvider:
		config := regru.NewDefaultConfig()
		config.Username = param.Username
		config.Password = param.Password
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		config.TTL = ttl
		p, err = regru.NewDNSProviderConfig(config)
	case DynuDnsProvider:
		config := dynu.NewDefaultConfig()
		config.APIKey = param.APIkey
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		config.TTL = ttl
		p, err = dynu.NewDNSProviderConfig(config)
	case BaiduCloudDnsProvider:
		config := baiducloud.NewDefaultConfig()
		config.AccessKeyID = param.AccessKey
		config.SecretAccessKey = param.SecretKey
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		config.TTL = ttl
		p, err = baiducloud.NewDNSProviderConfig(config)
	}

	if err != nil {
		return nil, err
	}
	return p, nil
}

type CreateDnsAccountRequest struct {
	ProviderId int    `json:"provider_id" binding:"required"`
	Key        string `json:"key" binding:"required"`
	Value      string `json:"value" binding:"required"`
}

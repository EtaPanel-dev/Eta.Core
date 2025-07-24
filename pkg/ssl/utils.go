package ssl

func GetCaDirURL(accountType, customCaURL string) string {
	var caDirURL string
	switch accountType {
	case "letsencrypt":
		caDirURL = "https://acme-v02.api.letsencrypt.org/directory"
	case "zerossl":
		caDirURL = "https://acme.zerossl.com/v2/DV90"
	case "buypass":
		caDirURL = "https://api.buypass.com/acme/directory"
	case "google":
		caDirURL = "https://dv.acme-v02.api.pki.goog/directory"
	case "freessl":
		caDirURL = "https://acmepro.freessl.cn/v2/DV"
	case "custom":
		caDirURL = customCaURL
	}
	return caDirURL
}

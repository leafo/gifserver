package gifserver

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
)

func checkSignature(r *http.Request) error {
	if serverConfig.Secret == "" {
		return nil
	}

	params := r.URL.Query()
	sig := params.Get("sig")

	if sig == "" {
		return fmt.Errorf("Missing signature")
	}

	patt := regexp.MustCompile(`[?&]sig=[^?&]+`)

	toCheck := r.URL.Path
	strippedQuery := patt.ReplaceAllString(r.URL.RawQuery, "")

	if strippedQuery != "" {
		toCheck = toCheck + "?" + strippedQuery
	}

	mac := hmac.New(sha1.New, []byte(serverConfig.Secret))
	mac.Write([]byte(toCheck))
	expectedSig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if expectedSig != sig {
		return fmt.Errorf("Invalid signature")
	}

	return nil
}

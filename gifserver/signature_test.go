package gifserver

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/http"
	"net/url"
	"testing"
)

func loadURL(rawURL string) *http.Request {
	parsed, _ := url.Parse(rawURL)
	return &http.Request{
		URL: parsed,
	}
}

func TestCheckSignature(t *testing.T) {
	err := checkSignature(loadURL("http://localhost:9090/transcode?url=hello"), "secret")
	if err == nil {
		t.Error("expecting failure from missing signature")
	}

	err = checkSignature(loadURL("http://localhost:9090/transcode?url=hello&sig=nope"), "secret")
	if err == nil {
		t.Error("expecting failure from invalid signature")
	}

	mac := hmac.New(sha1.New, []byte("secret"))
	mac.Write([]byte("/transcode?url=hello"))
	expectedSig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	err = checkSignature(loadURL("http://localhost:9090/transcode?url=hello&sig=" + expectedSig), "secret")
	if err != nil {
		t.Error("expecting signature to not fail")
	}

}

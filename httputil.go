package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var (
	httpClient = &http.Client{}
	pool       *x509.CertPool
)

type localCookieJar struct {
	jar map[string][]*http.Cookie
}

func (p *localCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	p.jar[u.Host] = cookies
}

func (p *localCookieJar) Cookies(u *url.URL) []*http.Cookie {
	return p.jar[u.Host]
}

func init() {
	pool = x509.NewCertPool()
	pool.AppendCertsFromPEM(pemCerts)
	httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: pool},
		},
	}

	jar := &localCookieJar{}
	jar.jar = make(map[string][]*http.Cookie)
	httpClient.Jar = jar
}

func postJsonData(api_end_point string, params interface{}) string {
	url1 := api_end_point
	postData, _ := json.Marshal(params)
	reqData := strings.NewReader(string(postData[:]))
	req, err := http.NewRequest("POST", url1, reqData)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)

	if err != nil {
		fmt.Printf("\n\nError : %s", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%s", body)
}

func debug(data []byte, err error) {
	if err == nil {
		fmt.Printf("%s\n\n", data)
	} else {
		log.Fatalf("%s\n\n", err)
	}
}

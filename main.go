package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/99designs/keyring"
)

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

var ring, initRingErr = keyring.Open(keyring.Config{
	//Keyring with encrypted application data
	ServiceName: "IAP_Proxy",
})

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

type proxy struct {
	authToken string
	host      url.URL
}

func (p *proxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	log.Println(req.RemoteAddr, " ", req.Method, " ", req.URL)

	client := &http.Client{}

	//http: Request.RequestURI can't be set in client requests.
	//http://golang.org/src/pkg/net/http/client.go
	req.RequestURI = ""
	req.URL.Host = p.host.Host
	req.URL.Scheme = p.host.Scheme

	//Attching token to http request
	req.Header.Add("Authorization", "Bearer "+p.authToken)

	delHopHeaders(req.Header)

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		appendHostToXForwardHeader(req.Header, clientIP)
	}

	//Request to IAP Proxy
	resp, err := client.Do(req)
	if err != nil {
		http.Error(wr, "Server Error", http.StatusInternalServerError)
		log.Fatal("ServeHTTP:", err)
	}
	defer resp.Body.Close()

	log.Println(req.RemoteAddr, " ", resp.Status)

	delHopHeaders(resp.Header)

	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)
	io.Copy(wr, resp.Body)
}

func main() {

	if initRingErr != nil {
		fmt.Printf("Error initialize key ring: %s", initRingErr.Error())
		return
	}

	clientID := os.Getenv("IAP_CLIENT_ID")
	iapHOSTENV := os.Getenv("IAP_HOST")

	//Service behind IAP Proxy (eg. iqlusion validator)
	iapHostURL, err := url.Parse(iapHOSTENV)

	if err != nil {
		fmt.Println(err)
		return
	}

	var addr = flag.String("addr", "127.0.0.1:8080", "The addr of the application.")
	//Path to credentials file
	var credentials = flag.String("cred", "", "Path to Service Account Web Token")
	flag.Parse()

	// If the user passed an argument, read the file
	if *credentials != "" {
		sa, err := ioutil.ReadFile(*credentials)
		if err != nil {
			fmt.Println(err)
			return
		}

		//Adding Proxy_Credentials in the keyring where data is sa
		err = ring.Set(keyring.Item{
			Key:  "Proxy_Credentials",
			Data: sa,
		})
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Credentials saved to keyring. Please delete %s \n", *credentials)
		return
	}

	//Brought in from iap.go file
	iap, err := newIAP(clientID)

	if err != nil {
		fmt.Println(err)
		return
	}

	token, err := iap.GetToken()
	fmt.Println(token)
	if err != nil {
		fmt.Println(err)
		return
	}

	handler := &proxy{token, *iapHostURL}

	log.Println("Starting proxy server on", *addr)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatal("ListenAndServe:", err)
	}

}

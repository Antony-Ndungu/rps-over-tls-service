package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/rpc"
	"os"

	"github.com/Antony-Ndungu/rpc-service/contract"
)

func main() {
	rootCACertPool := x509.NewCertPool()
	rootCA, err := ioutil.ReadFile("../minica.pem")
	if err != nil {
		fmt.Printf("failed to read CA cert %v\n", err)
		os.Exit(1)
	}
	rootCACertPool.AppendCertsFromPEM(rootCA)
	conn, err := tls.Dial("tcp", fmt.Sprintf("%v:%v", "server-cert", "1234"), &tls.Config{
		RootCAs: rootCACertPool,
		GetClientCertificate: func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			c, err := tls.LoadX509KeyPair("../client-cert/cert.pem", "../client-cert/key.pem")
			if err != nil {
				return nil, err
			}
			return &c, nil
		},
	})
	if err != nil {
		fmt.Printf("dialing failed: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	client := rpc.NewClient(conn)
	var reply contract.CatsResponse
	args := contract.CatsRequest{Cursor: 0, Limit: 10}
	err = client.Call("CatsHandler.GetCats", &args, &reply)
	if err != nil {
		fmt.Printf("method call failed: %v\n", err)
		os.Exit(1)
	}
	for _, v := range reply.Cats {
		fmt.Println(v.Name)
	}
}

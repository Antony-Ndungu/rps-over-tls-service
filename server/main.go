package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"flag"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/Antony-Ndungu/rpc-service/contract"
)

// CatsHandler provides access to a cat's methods remotely
type CatsHandler struct {
	db *sql.DB
}

func newCatsHandler(db *sql.DB) *CatsHandler {
	return &CatsHandler{db}
}

// GetCats fetches a list of cats
func (h *CatsHandler) GetCats(args *contract.CatsRequest, reply *contract.CatsResponse) error {
	rows, err := h.db.QueryContext(context.Background(), "SELECT id, name, weight, created_on, last_updated_on FROM catsapi.cats WHERE id > $1 ORDER BY id DESC LIMIT $2 ", args.Cursor, args.Limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cat contract.Cat
		var lastUpdatedOn sql.NullString
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Weight, &cat.CreatedOn, &lastUpdatedOn); err != nil {
			return err
		}
		if lastUpdatedOn.Valid {
			cat.LastUpdatedOn = lastUpdatedOn.String
		} else {
			cat.LastUpdatedOn = ""
		}
		reply.Cats = append(reply.Cats, &cat)
	}
	return nil
}

func main() {
	var (
		rpcAddr        = flag.String("http", "0.0.0.0:1234", "HTTP service address.")
		dataSourceName = flag.String("data-source-name", os.Getenv("DATA_SOURCE_NAME"), "PostgreSQL data source name.")
	)

	flag.Parse()
	if len(*dataSourceName) == 0 {
		log.Fatal("Missing data-source-name flag")
	}
	db, err := sql.Open("postgres", *dataSourceName)
	if err != nil {
		log.Fatalf("unable to use the data source name: %v", err)
	}
	defer db.Close()

	ctx, cancelFunc := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelFunc()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("database connection cannot be made: %v", err)
	}
	log.Println("Pinged the database successfully")

	rpc.Register(newCatsHandler(db))
	clientCAPool := x509.NewCertPool()
	clientCA, err := ioutil.ReadFile("../minica.pem")
	if err != nil {
		log.Fatalf("failed to load client ca certs: %v\n", err)
	}
	clientCAPool.AppendCertsFromPEM(clientCA)
	log.Printf("RPC server listening at %v", *rpcAddr)
	l, err := tls.Listen("tcp", *rpcAddr, &tls.Config{
		ClientCAs:  clientCAPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			c, err := tls.LoadX509KeyPair("../server-cert/cert.pem", "../server-cert/key.pem")
			if err != nil {
				log.Printf("failed to load certs: %v\n", err)
				return nil, err
			}
			return &c, nil
		},
	})
	if err != nil {
		log.Fatalf("failed to listen on the given address: %v", *rpcAddr)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("failed to accept a connection: %v", err)
			break
		}
		go rpc.ServeConn(conn)
	}
}

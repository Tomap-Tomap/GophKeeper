// Package parameters defines structure's parameters for client/servers work.
package parameters

import (
	"flag"
	"os"
	"strconv"
)

// ServerParameters contains parameters for server.
type ServerParameters struct {
	DSN               string
	TokenSecret       string
	PathToFileStorage string
	GRPCAddr          string

	TokenDuration uint
	ChunkSize     uint
}

// ParseFlagsServer return server's parameters from console or env.
func ParseFlagsServer() (p ServerParameters) {
	f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	f.StringVar(&p.DSN, "dsn", "host=localhost port=5433 user=gophkeeper password=gophkeeper dbname=gophkeeper sslmode=disable", "dsn to DB")
	f.StringVar(&p.TokenSecret, "secret", "secret", "secret for token sign")
	f.StringVar(&p.PathToFileStorage, "f", "/tmp", "path to file storage")
	f.StringVar(&p.GRPCAddr, "a", "localhost:3388", "address and port to run grpc server")
	f.UintVar(&p.TokenDuration, "td", 60, "how much token to be valid in minutes")
	f.UintVar(&p.ChunkSize, "cs", 1024, "how much bytes grpc server push on client")

	if DSN := os.Getenv("DSN"); DSN != "" {
		p.DSN = DSN
	}

	if tokenSecret := os.Getenv("TOKEN_SECRET"); tokenSecret != "" {
		p.TokenSecret = tokenSecret
	}

	if pathToFileStorage := os.Getenv("FILE_STORAGE_PATH"); pathToFileStorage != "" {
		p.PathToFileStorage = pathToFileStorage
	}

	if grpcAddr := os.Getenv("GRPC_ADDR"); grpcAddr != "" {
		p.GRPCAddr = grpcAddr
	}

	if tokenDuration := os.Getenv("TOKEN_DURATION"); tokenDuration != "" {
		intTD, err := strconv.ParseUint(tokenDuration, 10, 32)

		if err == nil {
			p.TokenDuration = uint(intTD)
		}
	}

	if chunkSize := os.Getenv("CHUNK_SIZE"); chunkSize != "" {
		intCS, err := strconv.ParseUint(chunkSize, 10, 32)

		if err == nil {
			p.ChunkSize = uint(intCS)
		}
	}

	return
}

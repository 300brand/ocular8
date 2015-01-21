package main

import (
	"flag"
)

var (
	dsn = flag.String("mongo", "mongodb://localhost:27017/test", "Connection string to MongoDB")
	nsq = flag.String("nsqd", "localhost:4151", "HTTP location for NSQd")
)

func main() {
	flag.Parse()
}

// Copyright 2012-2019 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/4406arthur/bello/pkg/entity"
	"github.com/nats-io/nats.go"
	"github.com/pquerna/ffjson/ffjson"
)

// NOTE: Can test with demo servers.
// go run stt_consumer.go -sub voice-0

func usage() {
	log.Printf("Usage: stt_consumer [-nats server] [-sub subject] [-t]\n")
	flag.PrintDefaults()
}

func showUsageAndExit(exitcode int) {
	usage()
	os.Exit(exitcode)
}

func printMsg(m *nats.Msg, i int) {
	log.Printf("[#%d] Received on [%s]: '%s'\n", i, m.Subject, string(m.Data))
}

func main() {
	var urls = flag.String("nats", "localhost:4222", "The nats server URLs (separated by comma)")
	var sub = flag.String("sub", "", "subscribe taget")
	var userCreds = flag.String("creds", "", "User Credentials File")
	var showTime = flag.Bool("t", false, "Display timestamps")
	var queueGroup = flag.String("q", "UASG1184", "Queue Group Name")
	var showHelp = flag.Bool("h", false, "Show help message")

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	if *showHelp {
		showUsageAndExit(0)
	}

	// Connect Options.
	opts := []nats.Option{nats.Name("NATS Sample Responder")}
	opts = setupConnOptions(opts)

	// Use UserCredentials
	if *userCreds != "" {
		opts = append(opts, nats.UserCredentials(*userCreds))
	}

	// Connect to NATS
	nc, err := nats.Connect(*urls, opts...)
	if err != nil {
		log.Fatal(err)
	}

	//demo respond
	mockResult := &entity.Response{
		ErrCode:     0,
		State:       "result",
		RecogResult: "good job",
	}

	i := 1
	nc.QueueSubscribe(*sub, *queueGroup, func(msg *nats.Msg) {
		printMsg(msg, i)
		if i%5 == 0 {
			result, _ := ffjson.Marshal(mockResult)
			msg.Respond(result)
		}
		i++
	})
	nc.Flush()

	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on [%s]", *sub)
	if *showTime {
		log.SetFlags(log.LstdFlags)
	}
	// Setup the interrupt handler to drain so we don't miss
	// requests when scaling down.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println()
	log.Printf("Draining...")
	nc.Drain()
	log.Fatalf("Exiting")
}

func setupConnOptions(opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Minute
	reconnectDelay := time.Second

	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectHandler(func(nc *nats.Conn) {
		log.Printf("Disconnected: will attempt reconnects for %.0fm", totalWait.Minutes())
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Printf("Reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Fatalf("Exiting: %v", nc.LastError())
	}))
	return opts
}

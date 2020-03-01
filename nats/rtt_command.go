// Copyright 2020 The NATS Authors
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
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"gopkg.in/alecthomas/kingpin.v2"
)

func configureRTTCommand(app *kingpin.Application) {
	app.Command("rtt", "Compute round-trip time to NATS server").Action(rtt)
}

func rtt(pc *kingpin.ParseContext) error {
	if !strings.Contains(servers, "://") {
		servers = fmt.Sprintf("nats://%s", servers)
	}

	u, err := url.Parse(servers)
	if err != nil {
		return err
	}

	if net.ParseIP(u.Hostname()) == nil {
		addrs, _ := net.LookupHost(u.Hostname())
		if len(addrs) <= 1 {
			goto doSingle
		}
		opts := natsOpts()
		nc, err := newNatsConn(servers, opts...)
		if err != nil {
			return err
		}
		log.Printf("[%s]\n", nc.ConnectedUrl())
		if nc.TLSRequired() {
			opts = append(opts, useTLSName(u.Hostname()))
		}
		nc.Close()
		for _, a := range addrs {
			server := fmt.Sprintf("%s://%s", u.Scheme, net.JoinHostPort(a, u.Port()))
			_, rtt, err := calcRTT(server, opts)
			if err != nil {
				return err
			}
			log.Printf("    [%-15s] = avg %v\n", a, rtt)
		}
		return nil
	}

doSingle:
	url, rtt, err := calcRTT(servers, natsOpts())
	if err != nil {
		return err
	}
	log.Printf("[%s] = avg %v\n", url, rtt)

	return nil
}

func useTLSName(name string) nats.Option {
	return func(o *nats.Options) error {
		o.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12, ServerName: name}
		return nil
	}
}

func calcRTT(server string, opts []nats.Option) (string, time.Duration, error) {
	const rttIters = 5

	nc, err := newNatsConn(server, opts...)
	if err != nil {
		return "", 0, err
	}
	defer nc.Close()

	time.Sleep(25 * time.Millisecond)

	var totalTime time.Duration
	for i := 1; i <= rttIters; i++ {
		start := time.Now()
		nc.Flush()
		rtt := time.Since(start)
		totalTime += rtt
	}
	return nc.ConnectedUrl(), (totalTime / rttIters), nil
}

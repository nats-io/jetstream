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
	"log"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
)

func configureRTTCommand(app *kingpin.Application) {
	app.Command("rtt", "Compute round-trip time to NATS server").Action(rtt)
}

const rttIters = 5

func rtt(pc *kingpin.ParseContext) error {
	// TODO(dlc) - we could resolve if we have multiple A records and test each one.
	nc, err := newNatsConn(servers, natsOpts()...)
	if err != nil {
		return err
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

	log.Printf("[%s] = avg %v\n", nc.ConnectedUrl(), totalTime/rttIters)
	return nil
}

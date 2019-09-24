/*
 * Copyright 2018-present Open Networking Foundation

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// This file implements an exit handler that tries to shut down all the
// running servers before finally exiting. There are 2 triggers to this
// clean exit thread: signals and an exit channel.

package afrouter

import (
	"github.com/opencord/voltha-go/common/log"
	"os"
	"os/signal"
	"syscall"
)

var errChan = make(chan error)
var doneChan = make(chan error)

func InitExitHandler() error {

	// Start the signal handler
	go signalHandler()
	// Start the error handler
	go errHandler()

	return nil
}

func signalHandler() {
	// Make signal channel and register notifiers for Interupt and Terminate
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	signal.Notify(sigchan, syscall.SIGTERM)
	signal.Notify(sigchan, syscall.SIGKILL)

	// Block until we receive a signal on the channel
	<-sigchan

	log.Info("shutting down on signal as requested")

	cleanExit(nil)
}

func errHandler() {

	err := <-errChan

	cleanExit(err)
}

func cleanExit(err error) {

	//Closing the streaming connections
	for _, cl := range clusters {
		for _, bknd := range cl.backends {
			for streamReq, _ := range bknd.activeRequests {
				if streamReq.isStreamingResponse {
					connection := streamReq.backend.connections
					for _, conn := range connection {
						log.Info("Closing the streaming request ",streamReq.methodInfo)
						conn.close()
					}
				}
			}
		}
	}

	// Log the shutdown
	if arProxy != nil {
		for _, srvr := range arProxy.servers {
			if srvr.running {
				log.With(log.Fields{"server": srvr.name}).Debug("Closing server")
				srvr.proxyServer.GracefulStop()
				srvr.proxyListener.Close()
			}
		}
	}
	doneChan <- err
	//os.Exit(0)
}

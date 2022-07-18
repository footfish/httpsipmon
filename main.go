package main

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/jart/gosip/sip"
	"github.com/jart/gosip/util"
)

const sipUser, userAgent = "checksip", "httpsipmon/1.0"
const httpPort = "8080"
const appName = "httpsipmon"

// Runs a http server which will send a SIP OPTIONS message each time page is fetched.
// The http server returns the status code from the remote SIP server
// Can be used to check a remote sip servers status with http.
func main() {

	if len(os.Args) == 1 {
		fmt.Println(appName + " is an http server which sends OPTIONS to a SIP server and returns the response as an http status code")
		fmt.Println("usage: " + appName + " HOST:PORT")
		os.Exit(0)
	}
	hostAddress := os.Args[1]
	http.HandleFunc("/", sipmon(hostAddress))
	http.ListenAndServe(":"+httpPort, nil)

}

func sipmon(hostAddress string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code, err := sendOptions(hostAddress)
		w.WriteHeader(code)
		if err != nil {
			fmt.Fprintf(w, "response %d - %s", code, err)
		} else {
			fmt.Fprintf(w, "OK")
		}
	}
}

// send an OPTIONS message to hostAddress and return the status code.
func sendOptions(hostAddress string) (statusCode int, err error) {
	sock, err := net.Dial("udp", hostAddress)
	if err != nil {
		return 418, fmt.Errorf(appName+" failed to create socket %s", hostAddress)
	}

	defer sock.Close()
	raddr := sock.RemoteAddr().(*net.UDPAddr)
	laddr := sock.LocalAddr().(*net.UDPAddr)

	options := sip.Msg{
		CSeq:       util.GenerateCSeq(),
		CallID:     util.GenerateCallID(),
		Method:     "OPTIONS",
		CSeqMethod: "OPTIONS",
		Accept:     "application/sdp",
		UserAgent:  userAgent,
		Request: &sip.URI{
			Scheme: "sip",
			User:   sipUser,
			Host:   raddr.IP.String(),
			Port:   uint16(raddr.Port),
		},
		Via: &sip.Via{
			Version: "2.0",
			Host:    laddr.IP.String(),
			Port:    uint16(laddr.Port),
			Param:   &sip.Param{Name: "branch", Value: util.GenerateBranch()},
		},
		Contact: &sip.Addr{
			Uri: &sip.URI{
				Host: laddr.IP.String(),
				Port: uint16(laddr.Port),
			},
		},
		From: &sip.Addr{
			Uri: &sip.URI{
				User: sipUser,
				Host: laddr.IP.String(),
				//Port: 5060,
			},
			Param: &sip.Param{Name: "tag", Value: util.GenerateTag()},
		},
		To: &sip.Addr{
			Uri: &sip.URI{
				Host: raddr.IP.String(),
				Port: uint16(raddr.Port),
			},
		},
	}

	var b bytes.Buffer
	options.Append(&b)
	if amt, err := sock.Write(b.Bytes()); err != nil || amt != b.Len() {
		return 418, fmt.Errorf(appName+" can't write to socket %s", hostAddress)
	}

	memory := make([]byte, 2048)
	sock.SetDeadline(time.Now().Add(time.Second))
	amt, err := sock.Read(memory)
	if err != nil {
		return 504, fmt.Errorf(appName + " timeout waiting for response")
	}

	msg, err := sip.ParseMsg(memory[0:amt])
	if err != nil {
		return 500, fmt.Errorf(appName + " can't parse SIP response")
	}

	if msg.Status != 200 {
		return msg.Status, fmt.Errorf("SIP server says: " + msg.Phrase)
	}
	return msg.Status, nil
}

package upstream

import (
	"time"

	"github.com/AdguardTeam/golibs/log"
	"github.com/miekg/dns"
)

//
// plain DNS
//
type plainDNS struct {
	address   string
	timeout   time.Duration
	preferTCP bool
}

// type check
var _ Upstream = &plainDNS{}

// Address returns the original address that we've put in initially, not resolved one
func (p *plainDNS) Address() string {
	if p.preferTCP {
		return "tcp://" + p.address
	}
	return p.address
}

func (p *plainDNS) Exchange(m *dns.Msg) (*dns.Msg, error) {
	q := m.Question[0].String()
	if p.preferTCP {
		log.Tracef("\nEstablishing DoTCP connection for: %s\nTime: %v\n", q, time.Now().Format(time.StampMilli))
		tcpClient := dns.Client{Net: "tcp", Timeout: p.timeout}
		log.Tracef("\nEstablished DoTCP connection for: %s\nTime: %v\n", q, time.Now().Format(time.StampMilli))
		logBegin(p.Address(), m)
		log.Tracef("\nSending DoTCP query: %s\nTime: %v\n", q, time.Now().Format(time.StampMilli))
		reply, _, tcpErr := tcpClient.Exchange(m, p.address)
		log.Tracef("\nDoTCP answer received for: %s\nTime: %v\n", q, time.Now().Format(time.StampMilli))
		logFinish(p.Address(), tcpErr)
		return reply, tcpErr
	}

	client := dns.Client{Timeout: p.timeout, UDPSize: dns.MaxMsgSize}

	logBegin(p.Address(), m)
	log.Tracef("\nSending DoUDP query: %s\nTime: %v\n", q, time.Now().Format(time.StampMilli))
	reply, _, err := client.Exchange(m, p.address)
	log.Tracef("\nDoUDP answer received for: %s\nTime: %v\n", q, time.Now().Format(time.StampMilli))
	logFinish(p.Address(), err)

	if reply != nil && reply.Truncated {
		log.Tracef("Truncated message was received, retrying over TCP, question: %s", m.Question[0].String())
		log.Tracef("\nEstablishing DoTCP connection for: %s\nTime: %v\n", q, time.Now().Format(time.StampMilli))
		tcpClient := dns.Client{Net: "tcp", Timeout: p.timeout}
		log.Tracef("\nEstablished DoTCP connection for: %s\nTime: %v\n", q, time.Now().Format(time.StampMilli))
		logBegin(p.Address(), m)
		log.Tracef("\nSending DoTCP query: %s\nTime: %v\n", q, time.Now().Format(time.StampMilli))
		reply, _, err = tcpClient.Exchange(m, p.address)
		log.Tracef("\nDoTCP answer received for: %s\nTime: %v\n", q, time.Now().Format(time.StampMilli))
		logFinish(p.Address(), err)
	}

	return reply, err
}

func (p *plainDNS) Reset() {
}

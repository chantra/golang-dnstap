package dnstap

/*
    Copyright (c) 2013 by Internet Systems Consortium, Inc. ("ISC")

    Permission to use, copy, modify, and/or distribute this software for any
    purpose with or without fee is hereby granted, provided that the above
    copyright notice and this permission notice appear in all copies.

    THE SOFTWARE IS PROVIDED "AS IS" AND ISC DISCLAIMS ALL WARRANTIES
    WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
    MERCHANTABILITY AND FITNESS.  IN NO EVENT SHALL ISC BE LIABLE FOR
    ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
    WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
    ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT
    OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
*/

import "bytes"
import "fmt"
import "net"
import "strconv"
import "time"

import "github.com/miekg/dns"

import dnstapProto "github.com/dnstap/golang-dnstap/dnstap.pb"

const quietTimeFormat = "15:04:05"

func textConvertTime(s *bytes.Buffer, secs *uint64, nsecs *uint32) {
    if secs != nil {
        s.WriteString(time.Unix(int64(*secs), 0).Format(quietTimeFormat))
    } else {
        s.WriteString("??:??:??")
    }
    if nsecs != nil {
        s.WriteString(fmt.Sprintf(".%06d", *nsecs / 1000))
    } else {
        s.WriteString(".??????")
    }
}

func textConvertMessage(m *dnstapProto.Message, s *bytes.Buffer) {
    isQuery := false
    printQueryAddress := false

    switch *m.Type {
    case dnstapProto.Message_CLIENT_QUERY,
         dnstapProto.Message_RESOLVER_QUERY,
         dnstapProto.Message_AUTH_QUERY,
         dnstapProto.Message_FORWARDER_QUERY:
            isQuery = true
    case dnstapProto.Message_CLIENT_RESPONSE,
         dnstapProto.Message_RESOLVER_RESPONSE,
         dnstapProto.Message_AUTH_RESPONSE,
         dnstapProto.Message_FORWARDER_RESPONSE:
            isQuery = false
    }

    if isQuery {
        textConvertTime(s, m.QueryTimeSec, m.QueryTimeNsec)
    } else {
        textConvertTime(s, m.ResponseTimeSec, m.ResponseTimeNsec)
    }
    s.WriteString(" ")

    switch *m.Type {
    case dnstapProto.Message_CLIENT_QUERY,
         dnstapProto.Message_CLIENT_RESPONSE: {
            s.WriteString("C")
         }
    case dnstapProto.Message_RESOLVER_QUERY,
         dnstapProto.Message_RESOLVER_RESPONSE: {
             s.WriteString("R")
         }
    case dnstapProto.Message_AUTH_QUERY,
         dnstapProto.Message_AUTH_RESPONSE: {
             s.WriteString("A")
         }
    case dnstapProto.Message_FORWARDER_QUERY,
         dnstapProto.Message_FORWARDER_RESPONSE: {
             s.WriteString("F")
         }
    case dnstapProto.Message_STUB_QUERY,
         dnstapProto.Message_STUB_RESPONSE: {
             s.WriteString("S")
         }
    }

    if isQuery {
        s.WriteString("Q ")
    } else {
        s.WriteString("R ")
    }

    switch *m.Type {
    case dnstapProto.Message_CLIENT_QUERY,
         dnstapProto.Message_AUTH_QUERY:
            printQueryAddress = true
    }

    if printQueryAddress {
        if m.QueryAddress != nil {
            s.WriteString(net.IP(m.QueryAddress).String())
        }
    } else {
        if m.ResponseAddress != nil {
            s.WriteString(net.IP(m.ResponseAddress).String())
        }
    }
    s.WriteString(" ")

    if m.SocketProtocol != nil {
        s.WriteString(m.SocketProtocol.String())
    }
    s.WriteString(" ")

    if isQuery {
        s.WriteString(strconv.Itoa(len(m.QueryMessage)))
        s.WriteString("b ")
    } else {
        s.WriteString(strconv.Itoa(len(m.ResponseMessage)))
        s.WriteString("b ")
    }

    if m.QueryName != nil {
        name, _, err := dns.UnpackDomainName(m.QueryName, 0)
        if err != nil {
            s.WriteString("X ")
        }
        s.WriteString(strconv.Quote(name))
    } else {
        s.WriteString("X ")
    }
    s.WriteString(" ")

    if m.QueryClass != nil {
        s.WriteString(dns.Class(*m.QueryClass).String())
    } else {
        s.WriteString("X ")
    }
    s.WriteString(" ")

    if m.QueryType != nil {
        s.WriteString(dns.Type(*m.QueryType).String())
    } else {
        s.WriteString("X")
    }

    s.WriteString("\n")
}

func textConvertPayload(dt *dnstapProto.Dnstap) (out []byte) {
    var s bytes.Buffer

    if *dt.Type == dnstapProto.Dnstap_MESSAGE {
        textConvertMessage(dt.Message, &s)
    }

    return s.Bytes()
}

func QuietTextConvert(buf []byte) (out []byte, ok bool) {
    dt, ok := Unpack(buf)
    if ok {
        return textConvertPayload(dt), true
    } else {
        return nil, false
    }
}

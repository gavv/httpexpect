package httpexpect

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type wsMessageType int

func (typ wsMessageType) String() string {
	s := "unknown"

	switch typ {
	case websocket.TextMessage:
		s = "text"
	case websocket.BinaryMessage:
		s = "binary"
	case websocket.CloseMessage:
		s = "close"
	case websocket.PingMessage:
		s = "ping"
	case websocket.PongMessage:
		s = "pong"
	}

	return fmt.Sprintf("%s(%d)", s, typ)
}

func wsMessageTypes(types []int) []interface{} {
	ret := make([]interface{}, 0, len(types))
	for _, t := range types {
		ret = append(ret, wsMessageType(t))
	}
	return ret
}

type wsCloseCode int

// https://developer.mozilla.org/en-US/docs/Web/API/CloseEvent/code
func (code wsCloseCode) String() string {
	s := "Unknown"

	switch code {
	case 1000:
		s = "NormalClosure"
	case 1001:
		s = "GoingAway"
	case 1002:
		s = "ProtocolError"
	case 1003:
		s = "UnsupportedData"
	case 1004:
		s = "Reserved"
	case 1005:
		s = "NoStatusReceived"
	case 1006:
		s = "AbnormalClosure"
	case 1007:
		s = "InvalidFramePayloadData"
	case 1008:
		s = "PolicyViolation"
	case 1009:
		s = "MessageTooBig"
	case 1010:
		s = "MandatoryExtension"
	case 1011:
		s = "InternalServerError"
	case 1012:
		s = "ServiceRestart"
	case 1013:
		s = "TryAgainLater"
	case 1014:
		s = "BadGateway"
	case 1015:
		s = "TLSHandshake"
	}

	return fmt.Sprintf("%s(%d)", s, code)
}

func wsCloseCodes(codes []int) []interface{} {
	ret := make([]interface{}, 0, len(codes))
	for _, c := range codes {
		ret = append(ret, wsCloseCode(c))
	}
	return ret
}

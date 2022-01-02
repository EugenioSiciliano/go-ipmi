package ipmi

import (
	"fmt"
)

type (
	AuthType uint8
)

const (
	AuthTypeNone     AuthType = 0x00
	AuthTypeMD2      AuthType = 0x01
	AuthTypeMD5      AuthType = 0x02
	AuthTypePassword AuthType = 0x04
	AuthTypeOEM      AuthType = 0x05
	AuthTypeRMCPPlus AuthType = 0x06
)

const (
	SessionHeader20SizeMax int = 18
	SessionHeader20SizeMin int = 12

	SessionHeader15SizeMax int = 26
	SessionHeader15SizeMin int = 10
)

// SessionHeader15 for IPMI 1.5
// see 22.12, Table 13.
//
// Whether the session header fields are present in a packet is based on
// whether the channel is specified as supporting multiple sessions or not.
// In addition, which session fields are present is based on the authentication type.
// Single-session connections and session-less channels do not include session header fields.
//
// Session header fields are present on all packets where the channel and
// connection mode is specified as supporting multiple sessions, even if
// the particular implementation only supports one session.
//
// Note that the command tables do not show the session header fields except for the
// Get Channel Authentication Capabilities, Get Session Challenge, and Activate Session commands.
// However, they are still required for all commands on a multi-session connection.
type SessionHeader15 struct {
	// For IPMI 1.5, it's value is 00h, 01h, 02h, 04h, 05h
	AuthType AuthType

	// For IPMI v2.0 RMCP+ there are separate sequence numbers tracked for authenticated and unauthenticated packets.
	// 0000_0000h is used for packets that are sent outside of a session.
	Sequence uint32

	SessionID uint32

	// The Authentication Code field in the session header may or may not be present based on the Authentication Type. The authentication code field is absent whenever the Authentication Type is NONE. Whether the authentication code field is present or not when the Authentication Type = OEM is dependent on the OEM identified in the Get Channel Authentication Capabilities command.
	//
	// 16 bytes, not present when Authentication Type set to none
	AuthCode []byte // IPMI 1.5

	// Payload length in bytes. 1-based.
	// IPMI 1.5 should be uint8
	// You should construct SessionHeader after the payload is created, thus you can fill the length here.
	PayloadLength uint8
}

func (h *SessionHeader15) Pack() []byte {
	var msg = make([]byte, 10+len(h.AuthCode))
	packUint8(uint8(h.AuthType), msg, 0)
	packUint32L(h.Sequence, msg, 1)
	packUint32L(h.SessionID, msg, 5)

	if h.AuthType != AuthTypeNone {
		packBytes(h.AuthCode, msg, 9)
	}

	packUint8(h.PayloadLength, msg, len(msg))
	return msg
}

func (h *SessionHeader15) Unpack(msg []byte) error {
	if len(msg) < SessionHeader15SizeMin {
		return ErrUnpackedDataTooShort
	}

	b, _, _ := unpackUint8(msg, 0)
	h.AuthType = AuthType(b)

	h.Sequence, _, _ = unpackUint32L(msg, 1)
	h.SessionID, _, _ = unpackUint32L(msg, 5)

	var payloadLengthIndex = 9
	if h.AuthType != AuthTypeNone {
		if len(msg) < SessionHeader15SizeMax {
			return ErrUnpackedDataTooShort
		}
		h.AuthCode, _, _ = unpackBytes(msg, 9, 16)
		payloadLengthIndex = 25
	}

	h.PayloadLength, _, _ = unpackUint8(msg, payloadLengthIndex)
	return nil
}

type Session15 struct {
	SessionHeader15 *SessionHeader15

	Payload []byte

	// legacy PAD not needed for IPMI v2.0
	LegacyPAD byte
}

func (s *Session15) Pack() []byte {
	out := s.SessionHeader15.Pack()
	out = append(out, s.Payload...)
	return out
}

func (s *Session15) Unpack(msg []byte) error {
	sessionHeader := &SessionHeader15{}
	err := sessionHeader.Unpack(msg)
	if err != nil {
		return fmt.Errorf("unpack SessionHeader15 failed, err: %s", err)
	}
	s.SessionHeader15 = sessionHeader

	sessionHeaderSize := len(sessionHeader.Pack())
	sessionPayloadSize := int(sessionHeader.PayloadLength)
	if len(msg) < sessionHeaderSize+sessionPayloadSize {
		return ErrUnpackedDataTooShort
	}
	s.Payload, _, _ = unpackBytes(msg, sessionHeaderSize, sessionPayloadSize)
	return nil
}

// SessionHeader20 for IPMI 2.0
type SessionHeader20 struct {
	// For IPMI 2.0, it's value is always 06h
	AuthType AuthType

	PayloadEncrypted     bool
	PayloadAuthenticated bool
	PayloadType          PayloadType

	// The complete identification of an OEM Payload is given by the combination of a three-byte IANA ID for the OEM, a reserved byte, plus a twobyte OEM Payload ID that is assigned and defined by the given OEM
	OEMIANA      uint32
	OEMPayloadID uint16

	// Should be set to bmcSessionID (generated by bmc, cached by remote console)
	SessionID uint32

	// For IPMI v2.0 RMCP+ there are separate sequence numbers tracked for authenticated and unauthenticated packets.
	// 0000_0000h is used for packets that are sent outside of a session.
	Sequence uint32

	// Payload length in bytes. 1-based.
	// You should construct SessionHeader after the payload is created, thus you can fill the length here.
	// IPMI 2.0 should be uint16
	PayloadLength uint16
}

func (h *SessionHeader20) Pack() []byte {
	// the longest length of SessionHeader20 is 18
	msg := make([]byte, 18)

	packUint8(uint8(h.AuthType), msg, 0)

	var encryptedMask uint8 = 0x00
	if h.PayloadEncrypted {
		encryptedMask = 0x80
	}
	var authenticatedMask uint8 = 0x00
	if h.PayloadAuthenticated {
		authenticatedMask = 0x40
	}
	packUint8(encryptedMask|authenticatedMask|uint8(h.PayloadType), msg, 1)

	var sessionIDIndex int
	var msgEndIndex int
	if h.PayloadType == PayloadTypeOEM {
		packUint32L(h.OEMIANA, msg, 2)
		packUint16L(h.OEMPayloadID, msg, 6)
		sessionIDIndex = 8
		msgEndIndex = 18
	} else {
		sessionIDIndex = 2
		msgEndIndex = 12
	}

	packUint32L(h.SessionID, msg, sessionIDIndex)
	packUint32L(h.Sequence, msg, sessionIDIndex+4)
	packUint16L(h.PayloadLength, msg, sessionIDIndex+8)

	return msg[:msgEndIndex]
}

func (h *SessionHeader20) Unpack(msg []byte) error {
	if len(msg) < SessionHeader20SizeMin {
		return ErrUnpackedDataTooShort
	}

	authType, _, _ := unpackUint8(msg, 0)
	h.AuthType = AuthType(authType)

	payloadType, _, _ := unpackUint8(msg, 1)
	h.PayloadEncrypted = payloadType&0x80 == 0x80     // bit 7 is set
	h.PayloadAuthenticated = payloadType&0x40 == 0x40 // bit 6 is set
	h.PayloadType = PayloadType(payloadType & 0x3f)   // clear bit 7 and bit 6

	var sessionIDIndex int
	if h.PayloadType == PayloadTypeOEM {
		if len(msg) < SessionHeader20SizeMax {
			return ErrUnpackedDataTooShort
		}
		h.OEMIANA, _, _ = unpackUint32L(msg, 2)
		h.OEMPayloadID, _, _ = unpackUint16L(msg, 6)
		sessionIDIndex = 8
	} else {
		sessionIDIndex = 2
	}

	h.SessionID, _, _ = unpackUint32L(msg, sessionIDIndex)
	h.Sequence, _, _ = unpackUint32L(msg, sessionIDIndex+4)
	h.PayloadLength, _, _ = unpackUint16L(msg, sessionIDIndex+8)
	return nil
}

type Session20 struct {
	SessionHeader20 *SessionHeader20

	// for encrypted packets, it should contain Confidentiality Header, Encrypted Payload, and Confidentiality Trailer.
	SessionPayload []byte

	// For IPMI v2.0 RMCP+ packets, the IPMI Session Trailer is absent whenever the Session ID is 0000_0000h, or whenever bit 6 in the payload type field indicates the packet is unauthenticated.
	SessionTrailer *SessionTrailer
}

func (s *Session20) Pack() []byte {
	out := s.SessionHeader20.Pack()
	out = append(out, s.SessionPayload...)

	if s.SessionTrailer != nil {
		out = append(out, s.SessionTrailer.Pack()...)
	}
	return out
}

func (s *Session20) Unpack(msg []byte) error {
	sessionHeader := &SessionHeader20{}
	if err := sessionHeader.Unpack(msg); err != nil {
		return fmt.Errorf("unpack SessionHeader failed, err: %s", err)
	}
	s.SessionHeader20 = sessionHeader

	var sessionHeaderSize int
	if sessionHeader.PayloadType == PayloadTypeOEM {
		sessionHeaderSize = SessionHeader20SizeMax
	} else {
		sessionHeaderSize = SessionHeader20SizeMin
	}

	payloadLength := int(s.SessionHeader20.PayloadLength)
	if len(msg) < sessionHeaderSize+payloadLength {
		return ErrUnpackedDataTooShort
	}
	s.SessionPayload, _, _ = unpackBytes(msg, sessionHeaderSize, payloadLength)

	s.SessionTrailer = nil
	sessionTrailerIndex := sessionHeaderSize + payloadLength
	if s.SessionHeader20.PayloadAuthenticated && s.SessionHeader20.SessionID != 0 {

		padSize := genSessionTrailerPadLength(sessionHeader.Pack(), s.SessionPayload)
		sessionTrailer := &SessionTrailer{}
		_, err := sessionTrailer.Unpack(msg, sessionTrailerIndex, padSize)
		if err != nil {
			return fmt.Errorf("unpack SessionTrailer failed, err: %s", err)
		}

		s.SessionTrailer = sessionTrailer
	}

	return nil
}

// For IPMI v2.0 RMCP+ packets, the IPMI Session Trailer is absent whenever the Session ID is 0000_0000h, or whenever bit 6 in the payload type field indicates the packet is unauthenticated
type SessionTrailer struct {
	// IPMI 2.0 and ASF only

	// Added as needed to cause the number of bytes in the data range covered by the AuthCode (Integrity Data) field to be a multiple of 4 bytes (DWORD). If present, each Integrity Pad byte is set to FFh.
	IntegrityPAD []byte

	// indicates how many pad bytes were added so that the amount of non-pad data can be determined.
	PadLength uint8

	// Reserved in IPMI v2.0. Set to 07h for RMCP+ packets defined in this specification.
	NextHeader uint8

	// For IPMI v2.0 (RMCP+) if this field is present, then it is calculated according to the Integrity Algorithm that was negotiated during the session open process. See Table 13-, Integrity Algorithm Numbers.
	// This field is absent when the packet is unauthenticated.
	AuthCode []byte // Integrity Data
}

func (s *SessionTrailer) Pack() []byte {
	msg := make([]byte, len(s.IntegrityPAD)+2+len(s.AuthCode))
	packBytes(s.IntegrityPAD, msg, 0)
	packUint8(s.PadLength, msg, len(s.IntegrityPAD))
	packUint8(s.NextHeader, msg, len(s.IntegrityPAD)+1)
	packBytes(s.AuthCode, msg, len(s.IntegrityPAD)+2)
	return msg
}

func (s *SessionTrailer) Unpack(msg []byte, off int, padSize int) (int, error) {
	var err error
	s.IntegrityPAD, off, _ = unpackBytes(msg, off, padSize)
	s.PadLength, off, _ = unpackUint8(msg, off)
	s.NextHeader, off, _ = unpackUint8(msg, off)
	s.AuthCode, off, _ = unpackBytesMost(msg, off, 16)
	return off, err
}

type SessionState uint8

const (
	SessionStatePreSession          SessionState = 0x00
	SessionStateOpenSessionSent     SessionState = 0x01
	SessionStateOpenSessionReceived SessionState = 0x02
	SessionStateRakp1Sent           SessionState = 0x03
	SessionStateRakp2Received       SessionState = 0x04
	SessionStateRakp3Sent           SessionState = 0x05
	SessionStateActive              SessionState = 0x06
	SessionStateCloseSent           SessionState = 0x07
)

func (c *Client) genSession15(rawPayload []byte) (*Session15, error) {
	//
	// Session Header
	//
	sessionHeader := &SessionHeader15{
		AuthType:      AuthTypeNone,
		Sequence:      0,
		SessionID:     0,
		AuthCode:      nil, // AuthCode would be filled afterward
		PayloadLength: uint8(len(rawPayload)),
	}

	if c.session.v15.active && c.session.authType != AuthTypeNone {
		sessionHeader.AuthType = c.session.authType
		sessionHeader.Sequence = c.session.v15.inSeq
		sessionHeader.SessionID = c.session.v15.sessionID

		authCode := c.genAuthCodeForMultiSession(rawPayload)

		c.DebugBytes("session 15 authcode", authCode, 16)
		sessionHeader.AuthCode = authCode
	}

	return &Session15{
		SessionHeader15: sessionHeader,
		Payload:         rawPayload,
	}, nil
}

func (c *Client) genSession20(payloadType PayloadType, rawPayload []byte) (*Session20, error) {
	//
	// Session Header
	//
	sessionHeader := &SessionHeader20{
		AuthType:             AuthTypeRMCPPlus, // Auth Type / Format is always 0x06 for IPMI v2
		PayloadType:          payloadType,
		PayloadAuthenticated: false,
		PayloadEncrypted:     false,
		SessionID:            0,
		Sequence:             0,
		PayloadLength:        0, // PayloadLength would be updated later after encryption if necessary.
	}

	if c.session.v20.state == SessionStateActive {
		sessionHeader.PayloadAuthenticated = true
		sessionHeader.PayloadEncrypted = true
		sessionHeader.SessionID = c.session.v20.bmcSessionID // use bmc session id
		sessionHeader.Sequence = c.session.v20.sequence
	}

	//
	// Session Payload
	//
	sessionPayload := rawPayload
	if c.session.v20.state == SessionStateActive && sessionHeader.PayloadEncrypted {
		e, err := c.encryptPlayload(rawPayload, nil)
		if err != nil {
			return nil, fmt.Errorf("encrypt payload failed, err: %s", err)
		}
		sessionPayload = e
	}
	// now we can fill PayloadLength field of the SessionHeader
	sessionHeader.PayloadLength = uint16(len(sessionPayload))
	c.DebugBytes("sessionPayload(final)", sessionPayload, 16)

	sessionHeaderBytes := sessionHeader.Pack()

	c.DebugBytes("sessionHeader", sessionHeaderBytes, 16)
	//
	// Session Trailer
	//
	var sessionTrailer *SessionTrailer = nil
	var err error
	// For IPMI v2.0 RMCP+ packets, the IPMI Session Trailer is absent
	// whenever the Session ID is 0000_0000h, or the packet is unauthenticated
	if sessionHeader.PayloadAuthenticated && sessionHeader.SessionID != 0 {
		sessionTrailer, err = c.genSessionTrailer(sessionHeaderBytes, sessionPayload)
		if err != nil {
			return nil, fmt.Errorf("genSessionTrailer failed, err: %s", err)
		}
	}

	return &Session20{
		SessionHeader20: sessionHeader,
		SessionPayload:  sessionPayload,
		SessionTrailer:  sessionTrailer,
	}, nil
}
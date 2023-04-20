package ipmi

import (
	"fmt"
)

// GetDellOEMMACCountRequest 22.14b Get Dell OEM MAC addresses Command
type GetDellOEMMACCountRequest struct {
	embNIC int
}

type GetDellOEMMACCountResponse struct {
	legacy   bool
	nicCount int
}

func (req *GetDellOEMMACCountRequest) Command() Command {
	return CommandGetSystemInfoParameters
}

func (req *GetDellOEMMACCountRequest) Pack() []byte {
	out := make([]byte, 6)
	_, _ = packUint8(uint8(0), out, 0)
	_, _ = packUint8(uint8(req.embNIC), out, 1)
	_, _ = packUint8(uint8(0), out, 2)
	_, _ = packUint8(uint8(0), out, 3)
	_, _ = packUint8(uint8(0), out, 4)
	_, _ = packUint8(uint8(0), out, 5)
	return out
}

func (res *GetDellOEMMACCountResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShort
	}
	if res.legacy {
		// TODO: implement this
	} else {
		maxLen, _, _ := unpackUint8(msg, 1)
		res.nicCount = int(maxLen / 8)
	}
	return nil
}

func (*GetDellOEMMACCountResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetDellOEMMACCountResponse) Format() string {
	return fmt.Sprintf("NICs count: %d", res.nicCount)
}

func (c *DellOEMClient) GetDellOEMMACCount() (nicCount int, err error) {

	request := &GetDellOEMMACCountRequest{
		embNIC: c.embNIC,
	}

	response := &GetDellOEMMACCountResponse{
		legacy: c.idrac10gOrOlder,
	}
	err = c.Exchange(request, response)
	if err != nil {
		return 0, err
	}
	nicCount = response.nicCount

	return
}

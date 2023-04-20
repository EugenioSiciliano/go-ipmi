package ipmi

import (
	"fmt"
)

// GetSystemInfoRequest 22.14b Get System info Command
type GetSystemInfoRequest struct {
	param int
	block int
}

type GetSystemInfoResponse struct {
	Response []byte
}

func (req *GetSystemInfoRequest) Command() Command {
	return CommandGetSystemInfoParameters
}

func (req *GetSystemInfoRequest) Pack() []byte {
	out := make([]byte, 4)
	_, _ = packUint8(uint8(0), out, 0)
	_, _ = packUint8(uint8(req.param), out, 1)
	_, _ = packUint8(uint8(req.block), out, 2)
	_, _ = packUint8(uint8(0), out, 3)
	return out
}

func (res *GetSystemInfoResponse) Unpack(msg []byte) error {
	res.Response = msg
	return nil
}

func (*GetSystemInfoResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetSystemInfoResponse) Format() string {
	return fmt.Sprintf("%d bytes", len(res.Response))
}

func (c *Client) GetSystemInfo(param int, block int) (response *GetSystemInfoResponse, err error) {
	request := &GetSystemInfoRequest{param: param, block: block}
	response = &GetSystemInfoResponse{}
	err = c.Exchange(request, response)
	return
}

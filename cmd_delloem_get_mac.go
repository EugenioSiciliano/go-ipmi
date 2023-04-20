package ipmi

import (
	"fmt"
	"net"
)

// GetDellOEMMACRequest 22.14b Get Dell OEM MAC addresses Command
type GetDellOEMMACRequest struct {
	embNIC   int
	nicIndex int
}

type GetDellOEMMACResponse struct {
	legacy         bool
	bladSlotNumber int
	nicNumber      int
	mac            net.HardwareAddr
	enabled        bool
}

func (req *GetDellOEMMACRequest) Command() Command {
	return CommandGetSystemInfoParameters
}

func (req *GetDellOEMMACRequest) Pack() []byte {
	out := make([]byte, 6)
	_, _ = packUint8(uint8(0), out, 0)
	_, _ = packUint8(uint8(req.embNIC), out, 1)
	_, _ = packUint8(uint8(0), out, 2)
	_, _ = packUint8(uint8(0), out, 3)
	length := 8
	offset := req.nicIndex * length
	_, _ = packUint8(uint8(offset), out, 4)
	_, _ = packUint8(uint8(length), out, 5)
	return out
}

func (res *GetDellOEMMACResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShort
	}
	if res.legacy {
		// TODO: implement this
	} else {
		cursor := 1
		var (
			metadata       uint16
			bladSlotNumber int
			macType        int
			ethernetStatus int
			nicNumber      int
		)
		metadata, cursor, _ = unpackUint16(msg, cursor)

		bladSlotNumber = int((metadata & 0xf000) >> 12)
		macType = int((metadata & 0x0c00) >> 10)
		ethernetStatus = int((metadata & 0x0300) >> 8)
		nicNumber = int((metadata & 0x00f8) >> 3)

		// type ethernet
		if macType == 0 {
			res.bladSlotNumber = bladSlotNumber
			res.enabled = ethernetStatus == 0 // 0 = enabled
			res.nicNumber = nicNumber
			res.mac, _, _ = unpackBytes(msg, cursor, 6)
		}
	}
	return nil
}

func (*GetDellOEMMACResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetDellOEMMACResponse) Format() string {
	var enabled string
	if res.enabled {
		enabled = "Enabled"
	} else {
		enabled = "Disabled"
	}
	return fmt.Sprintf("Slot\tNIC Number\tMAC Address\t\tStatus\n%d\t%d\t\t%s\t%s",
		res.bladSlotNumber, res.nicNumber, res.mac.String(), enabled)
}

func (c *DellOEMClient) GetDellOEMMAC(nicIndex int) (
	bladSlotNumber int, nicNumber int, mac net.HardwareAddr, enabled bool, err error) {

	request := &GetDellOEMMACRequest{
		nicIndex: nicIndex,
		embNIC:   c.embNIC,
	}

	response := &GetDellOEMMACResponse{
		legacy: c.idrac10gOrOlder,
	}
	err = c.Exchange(request, response)
	if err != nil {
		return 0, 0, []byte{}, false, err
	}
	bladSlotNumber = response.bladSlotNumber
	nicNumber = response.nicNumber
	mac = response.mac
	enabled = response.enabled

	return
}

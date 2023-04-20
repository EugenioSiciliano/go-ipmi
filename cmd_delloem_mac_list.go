package ipmi

import (
	"fmt"
)

const (
	imcIdrac10g           = 0x08
	imcIdrac11gMonolithic = 0x0A
	imcIdrac11gModular    = 0x0B
	imcMaserLiteBMC       = 0x0D
	imcMaserLiteNU        = 0x0E
	imcIdrac12gMonolithic = 0x10
	imcIdrac12gModular    = 0x11
	imcIdrac13gMonolithic = 0x20
	imcIdrac13gModular    = 0x21
	imcIdrac13gDCS        = 0x22
	imcIdrac14gMonolithic = 0x30
	imcIdrac15gMonolithic = 0x40
)

func (c *DellOEMClient) GetDellOEMMACs() (macList []string, err error) {

	if c.embNIC >= 0 {
		nicCount, err := c.GetDellOEMMACCount()
		if err != nil {
			return nil, fmt.Errorf("error in getting NIC count")
		}

		for i := 0; i < nicCount; i++ {

			_, _, mac, _, err := c.GetDellOEMMAC(i)
			if err != nil {
				return nil, fmt.Errorf("error in getting NIC #%d", i+1)
			}
			macList = append(macList, mac.String())
		}
	}
	return
}

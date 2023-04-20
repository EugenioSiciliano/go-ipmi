package ipmi

import "fmt"

// DellOEMClient client for Dell specific OEM extensions
type DellOEMClient struct {
	*Client

	idrac10gOrOlder bool
	embNIC          int
}

// NewDellOEMClient creates a Dell-specific client
func NewDellOEMClient(client *Client) (dellClient *DellOEMClient, err error) {

	systemRes, err := client.GetSystemInfo(0xDD, 2)
	if err != nil {
		return nil, fmt.Errorf("GetSystemInfo failed, err: %s", err)
	}
	if len(systemRes.Response) < 11 {
		return nil, fmt.Errorf("empty GetSystemInfo response")
	}

	var (
		imcOlderGen bool
		imcTypeSet  bool
	)

	switch systemRes.Response[10] {
	case imcIdrac10g:
		// do legacy 10G operation
		imcOlderGen = true
		imcTypeSet = true
	case imcMaserLiteNU:
		fallthrough
	case imcMaserLiteBMC:
		fallthrough
	case imcIdrac11gModular:
		fallthrough
	case imcIdrac11gMonolithic:
		fallthrough
	case imcIdrac12gMonolithic:
		fallthrough
	case imcIdrac12gModular:
		fallthrough
	case imcIdrac13gMonolithic:
		fallthrough
	case imcIdrac13gModular:
		fallthrough
	case imcIdrac14gMonolithic:
		// do current 11G-14g operation
		imcOlderGen = false
		imcTypeSet = true
	case imcIdrac15gMonolithic:
		// do current 15g operation
		imcOlderGen = false
		imcTypeSet = true
	}

	var embNIC = -1
	if imcTypeSet {
		if imcOlderGen {
			// iDRAC 10G or older
			embNIC = 0xCB
		} else {
			// iDRAC 11G or newer
			embNIC = 0xDA
		}
	}

	dellClient = &DellOEMClient{
		Client:          client,
		embNIC:          embNIC,
		idrac10gOrOlder: imcOlderGen,
	}
	return
}

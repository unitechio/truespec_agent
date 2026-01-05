//go:build windows

package gpu

import (
	"strings"

	"github.com/StackExchange/wmi"
)

type win32VideoController struct {
	Name          string
	AdapterRAM    uint64
	DriverVersion string
	PNPDeviceID   string
}

func getGPUs() ([]GPUInfo, error) {
	var ctrls []win32VideoController
	err := wmi.Query(`
        SELECT Name, AdapterRAM, DriverVersion, PNPDeviceID
        FROM Win32_VideoController
    `, &ctrls)
	if err != nil {
		return nil, err
	}

	gpus := make([]GPUInfo, 0, len(ctrls))
	for i, c := range ctrls {
		vendor := detectVendor(c.Name)
		gpus = append(gpus, GPUInfo{
			Index:       i,
			Name:        c.Name,
			Vendor:      vendor,
			VRAMTotalMB: c.AdapterRAM / 1024 / 1024,
			Driver:      c.DriverVersion,
			BusID:       c.PNPDeviceID,
			Integrated:  isIntegrated(vendor),
		})
	}
	return gpus, nil
}

func containsFold(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}

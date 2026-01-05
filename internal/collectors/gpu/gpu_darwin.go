//go:build darwin

package gpu

import "github.com/jaypipes/ghw"

func getGPUs() ([]GPUInfo, error) {
	info, err := ghw.GPU()
	if err != nil {
		return nil, err
	}

	gpus := make([]GPUInfo, 0)
	for _, card := range info.GraphicsCards {
		gpus = append(gpus, GPUInfo{
			Name:   card.DeviceInfo.Product.Name,
			Vendor: card.DeviceInfo.Vendor.Name,
		})
	}

	return gpus, nil
}

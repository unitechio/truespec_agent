package gpu

type GPUInfo struct {
    Index       int
    Name        string
    Vendor      string
    VRAMTotalMB uint64
    VRAMUsedMB  uint64
    Driver      string
    BusID       string
    Integrated  bool
}

func GetGPUs() ([]GPUInfo, error) {
	return getGPUs()
}

package gpu

import "strings"

func detectVendor(name string) string {
    n := strings.ToLower(name)
    switch {
    case strings.Contains(n, "nvidia"):
        return "NVIDIA"
    case strings.Contains(n, "amd"), strings.Contains(n, "radeon"):
        return "AMD"
    case strings.Contains(n, "intel"):
        return "Intel"
    case strings.Contains(n, "apple"):
        return "Apple"
    default:
        return "Unknown"
    }
}

func isIntegrated(vendor string) bool {
    v := strings.ToLower(vendor)
    return v == "intel" || v == "apple"
}

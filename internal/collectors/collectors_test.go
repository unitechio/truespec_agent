package collectors

import (
	"context"
	"testing"
)

func TestSystemCollector(t *testing.T) {
	collector := &SystemCollector{}

	if collector.Name() != "system" {
		t.Errorf("Expected name 'system', got '%s'", collector.Name())
	}

	ctx := context.Background()
	data, err := collector.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	result, ok := data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map[string]interface{}")
	}

	// Check required fields
	requiredFields := []string{"hostname", "os", "platform", "uptime"}
	for _, field := range requiredFields {
		if _, exists := result[field]; !exists {
			t.Errorf("Missing required field: %s", field)
		}
	}
}

func TestCPUCollector(t *testing.T) {
	collector := &CPUCollector{}

	if collector.Name() != "cpu" {
		t.Errorf("Expected name 'cpu', got '%s'", collector.Name())
	}

	ctx := context.Background()
	data, err := collector.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	result, ok := data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map[string]interface{}")
	}

	// Check usage_percent exists and is valid
	if usage, exists := result["usage_percent"]; exists {
		if usageFloat, ok := usage.(float64); ok {
			if usageFloat < 0 || usageFloat > 100 {
				t.Errorf("Invalid CPU usage: %f", usageFloat)
			}
		}
	}
}

func TestMemoryCollector(t *testing.T) {
	collector := &MemoryCollector{}

	if collector.Name() != "memory" {
		t.Errorf("Expected name 'memory', got '%s'", collector.Name())
	}

	ctx := context.Background()
	data, err := collector.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	result, ok := data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map[string]interface{}")
	}

	// Check that total >= used
	total, totalOk := result["total_mb"].(uint64)
	used, usedOk := result["used_mb"].(uint64)

	if totalOk && usedOk {
		if used > total {
			t.Errorf("Used memory (%d) > Total memory (%d)", used, total)
		}
	}
}

func TestDiskCollector(t *testing.T) {
	collector := &DiskCollector{}

	if collector.Name() != "disk" {
		t.Errorf("Expected name 'disk', got '%s'", collector.Name())
	}

	ctx := context.Background()
	data, err := collector.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	result, ok := data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map[string]interface{}")
	}

	// Check partitions exist
	if _, exists := result["partitions"]; !exists {
		t.Error("Missing 'partitions' field")
	}
}

func TestNetworkCollector(t *testing.T) {
	collector := &NetworkCollector{CollectMAC: false}

	if collector.Name() != "network" {
		t.Errorf("Expected name 'network', got '%s'", collector.Name())
	}

	ctx := context.Background()
	data, err := collector.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	result, ok := data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map[string]interface{}")
	}

	// Check interfaces exist
	interfaces, exists := result["interfaces"]
	if !exists {
		t.Fatal("Missing 'interfaces' field")
	}

	// Verify MAC is not collected when disabled
	if ifaceList, ok := interfaces.([]map[string]interface{}); ok {
		for _, iface := range ifaceList {
			if _, hasMac := iface["mac"]; hasMac {
				t.Error("MAC address collected when CollectMAC=false")
			}
		}
	}
}

func TestNetworkCollectorWithMAC(t *testing.T) {
	collector := &NetworkCollector{CollectMAC: true}

	ctx := context.Background()
	data, err := collector.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	result, ok := data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map[string]interface{}")
	}

	// Verify MAC is collected when enabled
	interfaces, exists := result["interfaces"]
	if !exists {
		t.Fatal("Missing 'interfaces' field")
	}

	if ifaceList, ok := interfaces.([]map[string]interface{}); ok && len(ifaceList) > 0 {
		// At least one interface should have MAC
		foundMAC := false
		for _, iface := range ifaceList {
			if _, hasMac := iface["mac"]; hasMac {
				foundMAC = true
				break
			}
		}
		if !foundMAC {
			t.Error("No MAC addresses collected when CollectMAC=true")
		}
	}
}

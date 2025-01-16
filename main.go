package main

import (
	"fmt"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"log"
)

func main() {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		log.Fatalf("Unable to initialize NVML: %v", nvml.ErrorString(ret))
	}
	defer func() {
		ret := nvml.Shutdown()
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to shutdown NVML: %v", nvml.ErrorString(ret))
		}
	}()

	// Assuming the 0-th device is MIG-enabled
	device, ret := nvml.DeviceGetHandleByIndex(0)
	if ret != nvml.SUCCESS {
		log.Fatalf("Unable to get device at index 0: %v", nvml.ErrorString(ret))
	}

	// Create GPU Instance
	giProfileInfo, ret := device.GetGpuInstanceProfileInfo(nvml.GPU_INSTANCE_PROFILE_4_SLICE)
	if ret != nvml.SUCCESS {
		log.Fatalf("Unable to get GPU Instance Profile Info: %v", nvml.ErrorString(ret))
	}
	gi, ret := device.CreateGpuInstance(&giProfileInfo)
	if ret != nvml.SUCCESS {
		log.Fatalf("Unable to create GPU Instance: %v", nvml.ErrorString(ret))
	}
	fmt.Printf("[+] Created GPU Instance with Profile Info:\n    %+v\n", giProfileInfo)

	// Create Compute Instance
	ciProfileInfo, ret := gi.GetComputeInstanceProfileInfo(nvml.COMPUTE_INSTANCE_PROFILE_2_SLICE, nvml.COMPUTE_INSTANCE_ENGINE_PROFILE_SHARED)
	if ret != nvml.SUCCESS {
		log.Fatalf("Unable to get Compute Instance Profile Info: %v", nvml.ErrorString(ret))
	}
	_, ret = gi.CreateComputeInstance(&ciProfileInfo)
	if ret != nvml.SUCCESS {
		log.Fatalf("Unable to create Compute Instance: %v", nvml.ErrorString(ret))
	}
	fmt.Printf("[+] Created Compute Instance with Profile Info:\n    %+v\n", ciProfileInfo)
}

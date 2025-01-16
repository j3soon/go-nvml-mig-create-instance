# Create MIG Instances with NVML Go Bindings

Unofficial example on creating Multi-Instance GPU (MIG) instances with NVIDIA Management Library (NVML) Go bindings.

Prerequisites:
- [NVIDIA Driver](https://ubuntu.com/server/docs/nvidia-drivers-installation)
- [Docker](https://docs.docker.com/engine/install/ubuntu/)
- [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html)
- [Enable MIG Mode](https://docs.nvidia.com/datacenter/tesla/mig-user-guide/#enable-mig-mode)

Take A30 as an example:

1. Clone this repo and `cd` into it:
   ```sh
   git clone https://github.com/j3soon/go-nvml-mig-create-instance.git
   cd go-nvml-mig-create-instance
   ```
2. Launch docker container for Go:  
   ```sh
   docker run --rm -it --gpus all \
       -v $(pwd):/workspace \
       --cap-add=SYS_ADMIN \
       -e NVIDIA_MIG_CONFIG_DEVICES=all \
       golang
   # in the container
   cd /workspace
   ```
   Note: `--runtime=nvidia`, `-e NVIDIA_VISIBLE_DEVICES=all`, and `-e NVIDIA_DRIVER_CAPABILITIES=all` may be required depending on your environment and use cases.  
   Alternatively, you can install Go on your host machine and skip this step.
3. Run the example and observe results:
   ```sh
   go run main.go
   # List the available CIs and GIs
   nvidia-smi mig -lgi; nvidia-smi mig -lci;
   # Destroy all the CIs and GIs
   nvidia-smi mig -dci; nvidia-smi mig -dgi;
   ```

This should also work on A100/H100/H200 by substituting the MIG profile to [a supported one](https://docs.nvidia.com/datacenter/tesla/mig-user-guide/#supported-mig-profiles).

## Description

To create MIG GIs and CIs, we should get instance profile information and then create instances based on the profile.

### Example Code Snippet

```go
// Assuming the 0-th device is MIG-enabled
device, ret := nvml.DeviceGetHandleByIndex(0)
// Create GPU Instance
giProfileInfo, ret := device.GetGpuInstanceProfileInfo(nvml.GPU_INSTANCE_PROFILE_4_SLICE)
gi, ret := device.CreateGpuInstance(&giProfileInfo)
// Create Compute Instance
ciProfileInfo, ret := gi.GetComputeInstanceProfileInfo(nvml.COMPUTE_INSTANCE_PROFILE_2_SLICE, nvml.COMPUTE_INSTANCE_ENGINE_PROFILE_SHARED)
_, ret = gi.CreateComputeInstance(&ciProfileInfo)
```

> The following source code references are based on `go-nvml v0.12.4-1`

### Creating GPU Instances (GIs)

- After getting the device handle of the 0-th GPU, we want to create a GI based on `CreateGpuInstance`. Take a look at its Go binding (ref: [cpp](https://docs.nvidia.com/deploy/nvml-api/group__nvmlMultiInstanceGPU.html#group__nvmlMultiInstanceGPU_1ge2fc1ce20e869dbf249304e1abedd52f), [go](https://pkg.go.dev/github.com/NVIDIA/go-nvml/pkg/nvml#Device.CreateGpuInstance), [src](https://github.com/NVIDIA/go-nvml/blob/8a7f6b796317a8c7cd5347a25be164108967e7b4/gen/nvml/nvml.h#L10106-L10129)):

  ```
  CreateGpuInstance(*GpuInstanceProfileInfo) (GpuInstance, Return)
  ```

- We can see that it takes the reference of `GpuInstanceProfileInfo` as the argument. Take a look at its source (ref: [cpp](https://docs.nvidia.com/deploy/nvml-api/structnvmlGpuInstanceProfileInfo__t.html#structnvmlGpuInstanceProfileInfo__t), [go](https://pkg.go.dev/github.com/NVIDIA/go-nvml/pkg/nvml#GpuInstanceProfileInfo), [src](https://github.com/NVIDIA/go-nvml/blob/8a7f6b796317a8c7cd5347a25be164108967e7b4/gen/nvml/nvml.h#L9752-L9768)):

  ```
  /**
   * GPU instance profile information.
   */
  typedef struct nvmlGpuInstanceProfileInfo_st
  {
      unsigned int id;                  //!< Unique profile ID within the device
      unsigned int isP2pSupported;      //!< Peer-to-Peer support
      unsigned int sliceCount;          //!< GPU Slice count
      unsigned int instanceCount;       //!< GPU instance count
      unsigned int multiprocessorCount; //!< Streaming Multiprocessor count
      unsigned int copyEngineCount;     //!< Copy Engine count
      unsigned int decoderCount;        //!< Decoder Engine count
      unsigned int encoderCount;        //!< Encoder Engine count
      unsigned int jpegCount;           //!< JPEG Engine count
      unsigned int ofaCount;            //!< OFA Engine count
      unsigned long long memorySizeMB;  //!< Memory size in MBytes
  } nvmlGpuInstanceProfileInfo_t;
  ```

- We suspect that these information isn't meant to be filled by hand. We should check the source for using the `GetGpuInstanceProfileInfo` API to retrieve these information (ref: [cpp](https://docs.nvidia.com/deploy/nvml-api/group__nvmlMultiInstanceGPU.html#group__nvmlMultiInstanceGPU_1g3f1c9b8bbbe2188d9c0c4f8bdb021a4d), [go](https://pkg.go.dev/github.com/NVIDIA/go-nvml/pkg/nvml#Device.GetGpuInstanceProfileInfo), [src](https://github.com/NVIDIA/go-nvml/blob/8a7f6b796317a8c7cd5347a25be164108967e7b4/gen/nvml/nvml.h#L10002-L10022)):

  ```
  /**
   * Get GPU instance profile information
   *
   * Information provided by this API is immutable throughout the lifetime of a MIG mode.
   *
   * For Ampere &tm; or newer fully supported devices.
   * Supported on Linux only.
   *
   * @param device                               The identifier of the target device
   * @param profile                              One of the NVML_GPU_INSTANCE_PROFILE_*
   * @param info                                 Returns detailed profile information
   *
   * @return
   *         - \ref NVML_SUCCESS                 Upon success
   *         - \ref NVML_ERROR_UNINITIALIZED     If library has not been successfully initialized
   *         - \ref NVML_ERROR_INVALID_ARGUMENT  If \a device, \a profile or \a info are invalid
   *         - \ref NVML_ERROR_NOT_SUPPORTED     If \a device doesn't support MIG or \a profile isn't supported
   *         - \ref NVML_ERROR_NO_PERMISSION     If user doesn't have permission to perform the operation
   */
  nvmlReturn_t DECLDIR nvmlDeviceGetGpuInstanceProfileInfo(nvmlDevice_t device, unsigned int profile,
                                                           nvmlGpuInstanceProfileInfo_t *info);
  ```

- Seems like we need to pass a `NVML_GPU_INSTANCE_PROFILE_*` as the `profile` argument. Let's view the source (ref: [cpp](https://docs.nvidia.com/deploy/nvml-api/group__nvmlMultiInstanceGPU.html#group__nvmlMultiInstanceGPU_1g099e78d565c10a3a0300d26adebe9f81), [go](https://pkg.go.dev/github.com/NVIDIA/go-nvml/pkg/nvml#GPU_INSTANCE_PROFILE_1_SLICE), [src](https://github.com/NVIDIA/go-nvml/blob/8a7f6b796317a8c7cd5347a25be164108967e7b4/gen/nvml/nvml.h#L9712-L9728)):

  ```
  /**
   * GPU instance profiles.
   *
   * These macros should be passed to \ref nvmlDeviceGetGpuInstanceProfileInfo to retrieve the
   * detailed information about a GPU instance such as profile ID, engine counts.
   */
  #define NVML_GPU_INSTANCE_PROFILE_1_SLICE      0x0
  #define NVML_GPU_INSTANCE_PROFILE_2_SLICE      0x1
  #define NVML_GPU_INSTANCE_PROFILE_3_SLICE      0x2
  #define NVML_GPU_INSTANCE_PROFILE_4_SLICE      0x3
  #define NVML_GPU_INSTANCE_PROFILE_7_SLICE      0x4
  #define NVML_GPU_INSTANCE_PROFILE_8_SLICE      0x5
  #define NVML_GPU_INSTANCE_PROFILE_6_SLICE      0x6
  #define NVML_GPU_INSTANCE_PROFILE_1_SLICE_REV1 0x7
  #define NVML_GPU_INSTANCE_PROFILE_2_SLICE_REV1 0x8
  #define NVML_GPU_INSTANCE_PROFILE_1_SLICE_REV2 0x9
  #define NVML_GPU_INSTANCE_PROFILE_COUNT        0xA
  ```

  > Please note that the `NVML_GPU_INSTANCE_PROFILE_COUNT` here is only a trick to get the number of profiles. It is not meant to be used as a profile.

- We can see that our hypothesis is correct based on the comments. We use `NVML_GPU_INSTANCE_PROFILE_4_SLICE` in our example.

### Creating Compute Instances (CIs)

- After creating a GI, we want to create a CI based on `CreateComputeInstance`. Take a look at its Go binding (ref: [cpp](https://docs.nvidia.com/deploy/nvml-api/group__nvmlMultiInstanceGPU.html#group__nvmlMultiInstanceGPU_1ga79996bd9a91d34b2f486112653d0042), [go](https://pkg.go.dev/github.com/NVIDIA/go-nvml/pkg/nvml#Interface.GpuInstanceCreateComputeInstance), [src](https://github.com/NVIDIA/go-nvml/blob/8a7f6b796317a8c7cd5347a25be164108967e7b4/gen/nvml/nvml.h#L10352-L10378)):

  ```
  GpuInstanceCreateComputeInstance(GpuInstance, *ComputeInstanceProfileInfo) (ComputeInstance, Return)
  ```

- Similar to the case in creating GIs, we'll need a `ComputeInstanceProfileInfo`. Let's look at its source (ref: [cpp](https://docs.nvidia.com/deploy/nvml-api/structnvmlComputeInstanceProfileInfo__t.html#structnvmlComputeInstanceProfileInfo__t), [go](https://pkg.go.dev/github.com/NVIDIA/go-nvml/pkg/nvml#ComputeInstanceProfileInfo), [src](https://github.com/NVIDIA/go-nvml/blob/8a7f6b796317a8c7cd5347a25be164108967e7b4/gen/nvml/nvml.h#L9863-L9877)):

  ```
  /**
   * Compute instance profile information.
   */
  typedef struct nvmlComputeInstanceProfileInfo_st
  {
      unsigned int id;                    //!< Unique profile ID within the GPU instance
      unsigned int sliceCount;            //!< GPU Slice count
      unsigned int instanceCount;         //!< Compute instance count
      unsigned int multiprocessorCount;   //!< Streaming Multiprocessor count
      unsigned int sharedCopyEngineCount; //!< Shared Copy Engine count
      unsigned int sharedDecoderCount;    //!< Shared Decoder Engine count
      unsigned int sharedEncoderCount;    //!< Shared Encoder Engine count
      unsigned int sharedJpegCount;       //!< Shared JPEG Engine count
      unsigned int sharedOfaCount;        //!< Shared OFA Engine count
  } nvmlComputeInstanceProfileInfo_t;
  ```

- Similarly, let's check the source for `GetComputeInstanceProfileInfo` API (ref: [cpp](https://docs.nvidia.com/deploy/nvml-api/group__nvmlMultiInstanceGPU.html#group__nvmlMultiInstanceGPU_1ge5ed2a23a041a3395886a324a0d2947e), [go](https://pkg.go.dev/github.com/NVIDIA/go-nvml/pkg/nvml#GpuInstance.GetComputeInstanceProfileInfo), [src](https://github.com/NVIDIA/go-nvml/blob/8a7f6b796317a8c7cd5347a25be164108967e7b4/gen/nvml/nvml.h#L10241-L10263)):

  ```
  /**
   * Get compute instance profile information.
   *
   * Information provided by this API is immutable throughout the lifetime of a MIG mode.
   *
   * For Ampere &tm; or newer fully supported devices.
   * Supported on Linux only.
   *
   * @param gpuInstance                          The identifier of the target GPU instance
   * @param profile                              One of the NVML_COMPUTE_INSTANCE_PROFILE_*
   * @param engProfile                           One of the NVML_COMPUTE_INSTANCE_ENGINE_PROFILE_*
   * @param info                                 Returns detailed profile information
   *
   * @return
   *         - \ref NVML_SUCCESS                 Upon success
   *         - \ref NVML_ERROR_UNINITIALIZED     If library has not been successfully initialized
   *         - \ref NVML_ERROR_INVALID_ARGUMENT  If \a gpuInstance, \a profile, \a engProfile or \a info are invalid
   *         - \ref NVML_ERROR_NOT_SUPPORTED     If \a profile isn't supported
   *         - \ref NVML_ERROR_NO_PERMISSION     If user doesn't have permission to perform the operation
   */
  nvmlReturn_t DECLDIR nvmlGpuInstanceGetComputeInstanceProfileInfo(nvmlGpuInstance_t gpuInstance, unsigned int profile,
                                                                    unsigned int engProfile,
                                                                    nvmlComputeInstanceProfileInfo_t *info);
  ```

- We should pass a `NVML_COMPUTE_INSTANCE_PROFILE_*` as the first (`profile`) argument. Let's view the source (ref: [cpp](https://docs.nvidia.com/deploy/nvml-api/group__nvmlMultiInstanceGPU.html#group__nvmlMultiInstanceGPU_1g9f17acf87c071aaaeb79b4d3c9bccebb), [go](https://pkg.go.dev/github.com/NVIDIA/go-nvml/pkg/nvml#COMPUTE_INSTANCE_PROFILE_1_SLICE), [src](https://github.com/NVIDIA/go-nvml/blob/8a7f6b796317a8c7cd5347a25be164108967e7b4/gen/nvml/nvml.h#L9838-L9852)):

  ```
  /**
   * Compute instance profiles.
   *
   * These macros should be passed to \ref nvmlGpuInstanceGetComputeInstanceProfileInfo to retrieve the
   * detailed information about a compute instance such as profile ID, engine counts
   */
  #define NVML_COMPUTE_INSTANCE_PROFILE_1_SLICE       0x0
  #define NVML_COMPUTE_INSTANCE_PROFILE_2_SLICE       0x1
  #define NVML_COMPUTE_INSTANCE_PROFILE_3_SLICE       0x2
  #define NVML_COMPUTE_INSTANCE_PROFILE_4_SLICE       0x3
  #define NVML_COMPUTE_INSTANCE_PROFILE_7_SLICE       0x4
  #define NVML_COMPUTE_INSTANCE_PROFILE_8_SLICE       0x5
  #define NVML_COMPUTE_INSTANCE_PROFILE_6_SLICE       0x6
  #define NVML_COMPUTE_INSTANCE_PROFILE_1_SLICE_REV1  0x7
  #define NVML_COMPUTE_INSTANCE_PROFILE_COUNT         0x8
  ```

- We use `COMPUTE_INSTANCE_PROFILE_2_SLICE` for the first argument in our example. As for the second argument (`engProfile`), let's also look at the source (ref: [cpp](https://docs.nvidia.com/deploy/nvml-api/group__nvmlMultiInstanceGPU.html#group__nvmlMultiInstanceGPU_1gb080d08cc8508ffc3ae348d7a4e37aa5), [go](https://pkg.go.dev/github.com/NVIDIA/go-nvml/pkg/nvml#COMPUTE_INSTANCE_ENGINE_PROFILE_SHARED), [src](https://github.com/NVIDIA/go-nvml/blob/8a7f6b796317a8c7cd5347a25be164108967e7b4/gen/nvml/nvml.h#L9854-L9855)):

  ```
  #define NVML_COMPUTE_INSTANCE_ENGINE_PROFILE_SHARED 0x0 //!< All the engines except multiprocessors would be shared
  #define NVML_COMPUTE_INSTANCE_ENGINE_PROFILE_COUNT  0x1
  ```

  > Although we currently only have the ability to share GPU engines (Copy Engine (CE), NVENC, NVDEC, NVJPEG, Optical Flow Accelerator (OFA), etc.) between CIs within the same GI, this struct may be extended to support isolating these engines for each CI within the same GI.

## References

Some references I found useful during the investigation.

- [NVIDIA/go-nvml](https://github.com/NVIDIA/go-nvml)
  - [Quick Start](https://github.com/NVIDIA/go-nvml?tab=readme-ov-file#quick-start)
  - [CreateComputeInstance() shows “Not Supported” - Issue #65](https://github.com/NVIDIA/go-nvml/issues/65)
- [(Archived) NVML in NVIDIA/gpu-monitoring-tools](https://github.com/NVIDIA/gpu-monitoring-tools/blob/master/bindings/go/nvml/nvml.h)
- [Docker MIG Manager](https://github.com/j3soon/docker-mig-manager)

API References (Useful for searching API definitions):

- [NVML API Reference Guide](https://docs.nvidia.com/deploy/nvml-api/index.html)
  - [4.28. Multi Instance GPU Management](https://docs.nvidia.com/deploy/nvml-api/group__nvmlMultiInstanceGPU.html)
- [NVML Go Bindings](https://pkg.go.dev/github.com/NVIDIA/go-nvml/pkg/nvml)
- [NVML Go Bindings Source Code Definitions](https://github.com/NVIDIA/go-nvml/blob/v0.12.4-1/gen/nvml/nvml.h)

## Acknowledgement

Thanks **Hsu-Tzu Ting** for discussions.

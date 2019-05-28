// Copyright 2017 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"time"
	"os"
	"strings"
        "strconv"

	gpumanager "github.com/dkozlov/container-engine-accelerators/pkg/gpu/nvidia"
	"github.com/golang/glog"
)

const (
	// Device plugin settings.
	kubeletEndpoint                   = "kubelet.sock"
	pluginEndpointPrefix              = "nvidiaGPU"
	devDirectory                      = "/dev"
        envExtendedResourceName           = "DP_EXTENDED_RESOURCE_NAME"
        envExtendedResourceValuePerDevice = "DP_EXTENDED_RESOURCE_VALUE_PER_DEVICE"
)


func getExtendedResourceValuePerDevice() (extendedResourceValue uint) {
        extendedResourceValue = 1 // default value
        strNum, present := os.LookupEnv(envExtendedResourceValuePerDevice)
        if !present {
                return
        }
        rawExtendedResourceValue, err := strconv.Atoi(strNum)
        if err != nil {
                glog.Errorf("Fatal: Could not parse %s environment variable: %v\n", envExtendedResourceValuePerDevice, err)
        }
        if rawExtendedResourceValue < 1 {
                glog.Errorf("Fatal: invalid %s environment variable value: %v\n", envExtendedResourceValuePerDevice, rawExtendedResourceValue)
        }
        extendedResourceValue = uint(rawExtendedResourceValue)
        return
}

var (
	hostPathPrefix       = flag.String("host-path", "/home/kubernetes/bin/nvidia", "Path on the host that contains nvidia libraries. This will be mounted inside the container as '-container-path'")
	containerPathPrefix  = flag.String("container-path", "/usr/local/nvidia", "Path on the container that mounts '-host-path'")
	pluginMountPath      = flag.String("plugin-directory", "/device-plugin", "The directory path to create plugin socket")
	gpuDuplicationFactor = flag.Uint("gpu-duplication-factor",  getExtendedResourceValuePerDevice(), "The number of fake GPU device declared per real GPU device")
)

func main() {
	flag.Parse()
	glog.Infoln("device-plugin started")
	ngm := gpumanager.NewSharedNvidiaGPUManager(*hostPathPrefix, *containerPathPrefix, devDirectory, *gpuDuplicationFactor)
	// Keep on trying until success. This is required
	// because Nvidia drivers may not be installed initially.
	for {
		err := ngm.Start()
		if err == nil {
			break
		}
		// Use non-default level to avoid log spam.
		glog.V(3).Infof("nvidiaGPUManager.Start() failed: %v", err)
		time.Sleep(5 * time.Second)
	}
	ngm.Serve(*pluginMountPath, kubeletEndpoint, fmt.Sprintf("%s-%d.sock", strings.Replace(os.Getenv(envExtendedResourceName), "/", ".", -1), time.Now().Unix()))
}

/*
 * Copyright 2018-present Open Networking Foundation

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package core

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/opencord/voltha-go/common/log"
	"github.com/opencord/voltha-go/db/model"
	fu "github.com/opencord/voltha-go/rw_core/utils"
	ic "github.com/opencord/voltha-protos/go/inter_container"
	ofp "github.com/opencord/voltha-protos/go/openflow_13"
	"github.com/opencord/voltha-protos/go/voltha"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"reflect"
	"sync"
)

type DeviceAgent struct {
	deviceId         string
	deviceType       string
	isRootdevice     bool
	lastData         *voltha.Device
	adapterProxy     *AdapterProxy
	adapterMgr       *AdapterManager
	deviceMgr        *DeviceManager
	clusterDataProxy *model.Proxy
	deviceProxy      *model.Proxy
	exitChannel      chan int
	lockDevice       sync.RWMutex
	defaultTimeout   int64
}

//newDeviceAgent creates a new device agent along as creating a unique ID for the device and set the device state to
//preprovisioning
func newDeviceAgent(ap *AdapterProxy, device *voltha.Device, deviceMgr *DeviceManager, cdProxy *model.Proxy, timeout int64) *DeviceAgent {
	var agent DeviceAgent
	agent.adapterProxy = ap
	cloned := (proto.Clone(device)).(*voltha.Device)
	if cloned.Id == "" {
		cloned.Id = CreateDeviceId()
		cloned.AdminState = voltha.AdminState_PREPROVISIONED
		cloned.FlowGroups = &ofp.FlowGroups{Items: nil}
		cloned.Flows = &ofp.Flows{Items: nil}
	}
	if !device.GetRoot() && device.ProxyAddress != nil {
		// Set the default vlan ID to the one specified by the parent adapter.  It can be
		// overwritten by the child adapter during a device update request
		cloned.Vlan = device.ProxyAddress.ChannelId
	}
	agent.isRootdevice = device.Root
	agent.deviceId = cloned.Id
	agent.deviceType = cloned.Type
	agent.lastData = cloned
	agent.deviceMgr = deviceMgr
	agent.adapterMgr = deviceMgr.adapterMgr
	agent.exitChannel = make(chan int, 1)
	agent.clusterDataProxy = cdProxy
	agent.lockDevice = sync.RWMutex{}
	agent.defaultTimeout = timeout
	return &agent
}

// start save the device to the data model and registers for callbacks on that device if loadFromdB is false.  Otherwise,
// it will load the data from the dB and setup teh necessary callbacks and proxies.
func (agent *DeviceAgent) start(ctx context.Context, loadFromdB bool) error {
	agent.lockDevice.Lock()
	defer agent.lockDevice.Unlock()
	log.Debugw("starting-device-agent", log.Fields{"deviceId": agent.deviceId})
	if loadFromdB {
		if device := agent.clusterDataProxy.Get("/devices/"+agent.deviceId, 1, false, ""); device != nil {
			if d, ok := device.(*voltha.Device); ok {
				agent.lastData = proto.Clone(d).(*voltha.Device)
				agent.deviceType = agent.lastData.Adapter
			}
		} else {
			log.Errorw("failed-to-load-device", log.Fields{"deviceId": agent.deviceId})
			return status.Errorf(codes.NotFound, "device-%s", agent.deviceId)
		}
		log.Debugw("device-loaded-from-dB", log.Fields{"device": agent.lastData})
	} else {
		// Add the initial device to the local model
		if added := agent.clusterDataProxy.AddWithID("/devices", agent.deviceId, agent.lastData, ""); added == nil {
			log.Errorw("failed-to-add-device", log.Fields{"deviceId": agent.deviceId})
		}
	}

	agent.deviceProxy = agent.clusterDataProxy.CreateProxy("/devices/"+agent.deviceId, false)
	agent.deviceProxy.RegisterCallback(model.POST_UPDATE, agent.processUpdate)

	log.Debug("device-agent-started")
	return nil
}

// stop stops the device agent.  Not much to do for now
func (agent *DeviceAgent) stop(ctx context.Context) {
	agent.lockDevice.Lock()
	defer agent.lockDevice.Unlock()
	log.Debug("stopping-device-agent")
	agent.exitChannel <- 1
	log.Debug("device-agent-stopped")
}

// GetDevice retrieves the latest device information from the data model
func (agent *DeviceAgent) getDevice() (*voltha.Device, error) {
	agent.lockDevice.RLock()
	defer agent.lockDevice.RUnlock()
	if device := agent.clusterDataProxy.Get("/devices/"+agent.deviceId, 0, false, ""); device != nil {
		if d, ok := device.(*voltha.Device); ok {
			cloned := proto.Clone(d).(*voltha.Device)
			return cloned, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "device-%s", agent.deviceId)
}

// getDeviceWithoutLock is a helper function to be used ONLY by any device agent function AFTER it has acquired the device lock.
// This function is meant so that we do not have duplicate code all over the device agent functions
func (agent *DeviceAgent) getDeviceWithoutLock() (*voltha.Device, error) {
	if device := agent.clusterDataProxy.Get("/devices/"+agent.deviceId, 0, false, ""); device != nil {
		if d, ok := device.(*voltha.Device); ok {
			cloned := proto.Clone(d).(*voltha.Device)
			return cloned, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "device-%s", agent.deviceId)
}

// enableDevice activates a preprovisioned or a disable device
func (agent *DeviceAgent) enableDevice(ctx context.Context) error {
	agent.lockDevice.Lock()
	defer agent.lockDevice.Unlock()
	log.Debugw("enableDevice", log.Fields{"id": agent.deviceId})

	if device, err := agent.getDeviceWithoutLock(); err != nil {
		return status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		// First figure out which adapter will handle this device type.  We do it at this stage as allow devices to be
		// pre-provisionned with the required adapter not registered.   At this stage, since we need to communicate
		// with the adapter then we need to know the adapter that will handle this request
		if adapterName, err := agent.adapterMgr.getAdapterName(device.Type); err != nil {
			log.Warnw("no-adapter-registered-for-device-type", log.Fields{"deviceType": device.Type, "deviceAdapter": device.Adapter})
			return err
		} else {
			device.Adapter = adapterName
		}

		if device.AdminState == voltha.AdminState_ENABLED {
			log.Debugw("device-already-enabled", log.Fields{"id": agent.deviceId})
			return nil
		}
		// If this is a child device then verify the parent state before proceeding
		if !agent.isRootdevice {
			if parent := agent.deviceMgr.getParentDevice(device); parent != nil {
				if parent.AdminState == voltha.AdminState_DISABLED ||
					parent.AdminState == voltha.AdminState_DELETED ||
					parent.AdminState == voltha.AdminState_UNKNOWN {
					err = status.Error(codes.FailedPrecondition, fmt.Sprintf("incorrect-parent-state: %s %d", parent.Id, parent.AdminState))
					log.Warnw("incorrect-parent-state", log.Fields{"id": agent.deviceId, "error": err})
					return err
				}
			} else {
				err = status.Error(codes.Unavailable, fmt.Sprintf("parent-not-existent: %s ", device.Id))
				log.Warnw("parent-not-existent", log.Fields{"id": agent.deviceId, "error": err})
				return err
			}
		}
		if device.AdminState == voltha.AdminState_PREPROVISIONED {
			// First send the request to an Adapter and wait for a response
			if err := agent.adapterProxy.AdoptDevice(ctx, device); err != nil {
				log.Debugw("adoptDevice-error", log.Fields{"id": agent.lastData.Id, "error": err})
				return err
			}
		} else {
			// First send the request to an Adapter and wait for a response
			if err := agent.adapterProxy.ReEnableDevice(ctx, device); err != nil {
				log.Debugw("renableDevice-error", log.Fields{"id": agent.lastData.Id, "error": err})
				return err
			}
		}
		// Received an Ack (no error found above).  Now update the device in the model to the expected state
		cloned := proto.Clone(device).(*voltha.Device)
		cloned.AdminState = voltha.AdminState_ENABLED
		cloned.OperStatus = voltha.OperStatus_ACTIVATING
		if afterUpdate := agent.clusterDataProxy.Update("/devices/"+agent.deviceId, cloned, false, ""); afterUpdate == nil {
			return status.Errorf(codes.Internal, "failed-update-device:%s", agent.deviceId)
		}
	}
	return nil
}

func (agent *DeviceAgent) updateDeviceWithoutLockAsync(device *voltha.Device, ch chan interface{}) {
	if err := agent.updateDeviceWithoutLock(device); err != nil {
		ch <- status.Errorf(codes.Internal, "failure-updating-%s", agent.deviceId)
	}
	ch <- nil
}

func (agent *DeviceAgent) sendBulkFlowsToAdapters(device *voltha.Device, flows *voltha.Flows, groups *voltha.FlowGroups, ch chan interface{}) {
	if err := agent.adapterProxy.UpdateFlowsBulk(device, flows, groups); err != nil {
		log.Debugw("update-flow-bulk-error", log.Fields{"id": agent.lastData.Id, "error": err})
		ch <- err
	}
	ch <- nil
}

func (agent *DeviceAgent) sendIncrementalFlowsToAdapters(device *voltha.Device, flows *ofp.FlowChanges, groups *ofp.FlowGroupChanges, ch chan interface{}) {
	if err := agent.adapterProxy.UpdateFlowsIncremental(device, flows, groups); err != nil {
		log.Debugw("update-flow-incremental-error", log.Fields{"id": agent.lastData.Id, "error": err})
		ch <- err
	}
	ch <- nil
}

func (agent *DeviceAgent) addFlowsAndGroups(newFlows []*ofp.OfpFlowStats, newGroups []*ofp.OfpGroupEntry) error {
	if (len(newFlows) | len(newGroups)) == 0 {
		log.Debugw("nothing-to-update", log.Fields{"deviceId": agent.deviceId, "flows": newFlows, "groups": newGroups})
		return nil
	}

	agent.lockDevice.Lock()
	defer agent.lockDevice.Unlock()
	log.Debugw("addFlowsAndGroups", log.Fields{"deviceId": agent.deviceId, "flows": newFlows, "groups": newGroups})
	var existingFlows *voltha.Flows
	if device, err := agent.getDeviceWithoutLock(); err != nil {
		return status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		existingFlows = proto.Clone(device.Flows).(*voltha.Flows)
		existingGroups := proto.Clone(device.FlowGroups).(*ofp.FlowGroups)
		log.Debugw("addFlows", log.Fields{"deviceId": agent.deviceId, "flows": newFlows, "existingFlows": existingFlows, "groups": newGroups, "existingGroups": existingGroups})

		var updatedFlows []*ofp.OfpFlowStats
		var flowsToDelete []*ofp.OfpFlowStats
		var groupsToDelete []*ofp.OfpGroupEntry
		var updatedGroups []*ofp.OfpGroupEntry

		// Process flows
		for _, flow := range newFlows {
			updatedFlows = append(updatedFlows, flow)
		}

		for _, flow := range existingFlows.Items {
			if idx := fu.FindFlows(newFlows, flow); idx == -1 {
				updatedFlows = append(updatedFlows, flow)
			} else {
				flowsToDelete = append(flowsToDelete, flow)
			}
		}

		// Process groups
		for _, g := range newGroups {
			updatedGroups = append(updatedGroups, g)
		}

		for _, group := range existingGroups.Items {
			if fu.FindGroup(newGroups, group.Desc.GroupId) == -1 { // does not exist now
				updatedGroups = append(updatedGroups, group)
			} else {
				groupsToDelete = append(groupsToDelete, group)
			}
		}

		// Sanity check
		if (len(updatedFlows) | len(flowsToDelete) | len(updatedGroups) | len(groupsToDelete)) == 0 {
			log.Debugw("nothing-to-update", log.Fields{"deviceId": agent.deviceId, "flows": newFlows, "groups": newGroups})
			return nil

		}
		// Send update to adapters
		chAdapters := make(chan interface{})
		defer close(chAdapters)
		chdB := make(chan interface{})
		defer close(chdB)
		dType := agent.adapterMgr.getDeviceType(device.Type)
		if !dType.AcceptsAddRemoveFlowUpdates {

			if len(updatedGroups) != 0 && reflect.DeepEqual(existingGroups.Items, updatedGroups) && len(updatedFlows) != 0 && reflect.DeepEqual(existingFlows.Items, updatedFlows) {
				log.Debugw("nothing-to-update", log.Fields{"deviceId": agent.deviceId, "flows": newFlows, "groups": newGroups})
				return nil
			}
			go agent.sendBulkFlowsToAdapters(device, &voltha.Flows{Items: updatedFlows}, &voltha.FlowGroups{Items: updatedGroups}, chAdapters)

		} else {
			flowChanges := &ofp.FlowChanges{
				ToAdd:    &voltha.Flows{Items: newFlows},
				ToRemove: &voltha.Flows{Items: flowsToDelete},
			}
			groupChanges := &ofp.FlowGroupChanges{
				ToAdd:    &voltha.FlowGroups{Items: newGroups},
				ToRemove: &voltha.FlowGroups{Items: groupsToDelete},
				ToUpdate: &voltha.FlowGroups{Items: []*ofp.OfpGroupEntry{}},
			}
			go agent.sendIncrementalFlowsToAdapters(device, flowChanges, groupChanges, chAdapters)
		}

		// store the changed data
		device.Flows = &voltha.Flows{Items: updatedFlows}
		device.FlowGroups = &voltha.FlowGroups{Items: updatedGroups}
		go agent.updateDeviceWithoutLockAsync(device, chdB)

		if res := fu.WaitForNilOrErrorResponses(agent.defaultTimeout, chAdapters, chdB); res != nil {
			return status.Errorf(codes.Aborted, "errors-%s", res)
		}

		return nil
	}
}

//disableDevice disable a device
func (agent *DeviceAgent) disableDevice(ctx context.Context) error {
	agent.lockDevice.Lock()
	log.Debugw("disableDevice", log.Fields{"id": agent.deviceId})
	// Get the most up to date the device info
	if device, err := agent.getDeviceWithoutLock(); err != nil {
		agent.lockDevice.Unlock()
		return status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		if device.AdminState == voltha.AdminState_DISABLED {
			log.Debugw("device-already-disabled", log.Fields{"id": agent.deviceId})
			agent.lockDevice.Unlock()
			return nil
		}
		// First send the request to an Adapter and wait for a response
		if err := agent.adapterProxy.DisableDevice(ctx, device); err != nil {
			log.Debugw("disableDevice-error", log.Fields{"id": agent.lastData.Id, "error": err})
			agent.lockDevice.Unlock()
			return err
		}
		// Received an Ack (no error found above).  Now update the device in the model to the expected state
		cloned := proto.Clone(device).(*voltha.Device)
		cloned.AdminState = voltha.AdminState_DISABLED
		if afterUpdate := agent.clusterDataProxy.Update("/devices/"+agent.deviceId, cloned, false, ""); afterUpdate == nil {
			agent.lockDevice.Unlock()
			return status.Errorf(codes.Internal, "failed-update-device:%s", agent.deviceId)
		}
		agent.lockDevice.Unlock()
	}
	return nil
}

func (agent *DeviceAgent) rebootDevice(ctx context.Context) error {
	agent.lockDevice.Lock()
	defer agent.lockDevice.Unlock()
	log.Debugw("rebootDevice", log.Fields{"id": agent.deviceId})
	// Get the most up to date the device info
	if device, err := agent.getDeviceWithoutLock(); err != nil {
		return status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		if device.AdminState != voltha.AdminState_DISABLED {
			log.Debugw("device-not-disabled", log.Fields{"id": agent.deviceId})
			//TODO:  Needs customized error message
			return status.Errorf(codes.FailedPrecondition, "deviceId:%s, expected-admin-state:%s", agent.deviceId, voltha.AdminState_DISABLED)
		}
		// First send the request to an Adapter and wait for a response
		if err := agent.adapterProxy.RebootDevice(ctx, device); err != nil {
			log.Debugw("rebootDevice-error", log.Fields{"id": agent.lastData.Id, "error": err})
			return err
		}
	}
	return nil
}

func (agent *DeviceAgent) deleteDevice(ctx context.Context) error {
	agent.lockDevice.Lock()
	log.Debugw("deleteDevice", log.Fields{"id": agent.deviceId})
	// Get the most up to date the device info
	if device, err := agent.getDeviceWithoutLock(); err != nil {
		agent.lockDevice.Unlock()
		return status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		if (device.AdminState != voltha.AdminState_DISABLED) &&
			(device.AdminState != voltha.AdminState_PREPROVISIONED) {
			log.Debugw("device-not-disabled", log.Fields{"id": agent.deviceId})
			//TODO:  Needs customized error message
			agent.lockDevice.Unlock()
			return status.Errorf(codes.FailedPrecondition, "deviceId:%s, expected-admin-state:%s", agent.deviceId, voltha.AdminState_DISABLED)
		}
		if device.AdminState != voltha.AdminState_PREPROVISIONED {
			// Send the request to an Adapter and wait for a response
			if err := agent.adapterProxy.DeleteDevice(ctx, device); err != nil {
				log.Debugw("deleteDevice-error", log.Fields{"id": agent.lastData.Id, "error": err})
				agent.lockDevice.Unlock()
				return err
			}
		}
		if removed := agent.clusterDataProxy.Remove("/devices/"+agent.deviceId, ""); removed == nil {
			agent.lockDevice.Unlock()
			return status.Errorf(codes.Internal, "failed-update-device:%s", agent.deviceId)
		}
		agent.lockDevice.Unlock()
	}
	return nil
}

func (agent *DeviceAgent) downloadImage(ctx context.Context, img *voltha.ImageDownload) (*voltha.OperationResp, error) {
	agent.lockDevice.Lock()
	defer agent.lockDevice.Unlock()
	log.Debugw("downloadImage", log.Fields{"id": agent.deviceId})
	// Get the most up to date the device info
	if device, err := agent.getDeviceWithoutLock(); err != nil {
		return nil, status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		if device.AdminState != voltha.AdminState_ENABLED {
			log.Debugw("device-not-enabled", log.Fields{"id": agent.deviceId})
			return nil, status.Errorf(codes.FailedPrecondition, "deviceId:%s, expected-admin-state:%s", agent.deviceId, voltha.AdminState_ENABLED)
		}
		// Save the image
		clonedImg := proto.Clone(img).(*voltha.ImageDownload)
		clonedImg.DownloadState = voltha.ImageDownload_DOWNLOAD_REQUESTED
		cloned := proto.Clone(device).(*voltha.Device)
		if cloned.ImageDownloads == nil {
			cloned.ImageDownloads = []*voltha.ImageDownload{clonedImg}
		} else {
			cloned.ImageDownloads = append(cloned.ImageDownloads, clonedImg)
		}
		cloned.AdminState = voltha.AdminState_DOWNLOADING_IMAGE
		if afterUpdate := agent.clusterDataProxy.Update("/devices/"+agent.deviceId, cloned, false, ""); afterUpdate == nil {
			return nil, status.Errorf(codes.Internal, "failed-update-device:%s", agent.deviceId)
		}
		// Send the request to the adapter
		if err := agent.adapterProxy.DownloadImage(ctx, cloned, clonedImg); err != nil {
			log.Debugw("downloadImage-error", log.Fields{"id": agent.lastData.Id, "error": err, "image": img.Name})
			return nil, err
		}
	}
	return &voltha.OperationResp{Code: voltha.OperationResp_OPERATION_SUCCESS}, nil
}

// isImageRegistered is a helper method to figure out if an image is already registered
func isImageRegistered(img *voltha.ImageDownload, device *voltha.Device) bool {
	for _, image := range device.ImageDownloads {
		if image.Id == img.Id && image.Name == img.Name {
			return true
		}
	}
	return false
}

func (agent *DeviceAgent) cancelImageDownload(ctx context.Context, img *voltha.ImageDownload) (*voltha.OperationResp, error) {
	agent.lockDevice.Lock()
	defer agent.lockDevice.Unlock()
	log.Debugw("cancelImageDownload", log.Fields{"id": agent.deviceId})
	// Get the most up to date the device info
	if device, err := agent.getDeviceWithoutLock(); err != nil {
		return nil, status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		// Verify whether the Image is in the list of image being downloaded
		if !isImageRegistered(img, device) {
			return nil, status.Errorf(codes.FailedPrecondition, "deviceId:%s, image-not-registered:%s", agent.deviceId, img.Name)
		}

		// Update image download state
		cloned := proto.Clone(device).(*voltha.Device)
		for _, image := range cloned.ImageDownloads {
			if image.Id == img.Id && image.Name == img.Name {
				image.DownloadState = voltha.ImageDownload_DOWNLOAD_CANCELLED
			}
		}

		//If device is in downloading state, send the request to cancel the download
		if device.AdminState == voltha.AdminState_DOWNLOADING_IMAGE {
			if err := agent.adapterProxy.CancelImageDownload(ctx, device, img); err != nil {
				log.Debugw("cancelImageDownload-error", log.Fields{"id": agent.lastData.Id, "error": err, "image": img.Name})
				return nil, err
			}
			// Set the device to Enabled
			cloned.AdminState = voltha.AdminState_ENABLED
			if afterUpdate := agent.clusterDataProxy.Update("/devices/"+agent.deviceId, cloned, false, ""); afterUpdate == nil {
				return nil, status.Errorf(codes.Internal, "failed-update-device:%s", agent.deviceId)
			}
		}
	}
	return &voltha.OperationResp{Code: voltha.OperationResp_OPERATION_SUCCESS}, nil
}

func (agent *DeviceAgent) activateImage(ctx context.Context, img *voltha.ImageDownload) (*voltha.OperationResp, error) {
	agent.lockDevice.Lock()
	defer agent.lockDevice.Unlock()
	log.Debugw("activateImage", log.Fields{"id": agent.deviceId})
	// Get the most up to date the device info
	if device, err := agent.getDeviceWithoutLock(); err != nil {
		return nil, status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		// Verify whether the Image is in the list of image being downloaded
		if !isImageRegistered(img, device) {
			return nil, status.Errorf(codes.FailedPrecondition, "deviceId:%s, image-not-registered:%s", agent.deviceId, img.Name)
		}

		if device.AdminState == voltha.AdminState_DOWNLOADING_IMAGE {
			return nil, status.Errorf(codes.FailedPrecondition, "deviceId:%s, device-in-downloading-state:%s", agent.deviceId, img.Name)
		}
		// Update image download state
		cloned := proto.Clone(device).(*voltha.Device)
		for _, image := range cloned.ImageDownloads {
			if image.Id == img.Id && image.Name == img.Name {
				image.ImageState = voltha.ImageDownload_IMAGE_ACTIVATING
			}
		}
		// Set the device to downloading_image
		cloned.AdminState = voltha.AdminState_DOWNLOADING_IMAGE
		if afterUpdate := agent.clusterDataProxy.Update("/devices/"+agent.deviceId, cloned, false, ""); afterUpdate == nil {
			return nil, status.Errorf(codes.Internal, "failed-update-device:%s", agent.deviceId)
		}

		if err := agent.adapterProxy.ActivateImageUpdate(ctx, device, img); err != nil {
			log.Debugw("activateImage-error", log.Fields{"id": agent.lastData.Id, "error": err, "image": img.Name})
			return nil, err
		}
		// The status of the AdminState will be changed following the update_download_status response from the adapter
		// The image name will also be removed from the device list
	}
	return &voltha.OperationResp{Code: voltha.OperationResp_OPERATION_SUCCESS}, nil
}

func (agent *DeviceAgent) revertImage(ctx context.Context, img *voltha.ImageDownload) (*voltha.OperationResp, error) {
	agent.lockDevice.Lock()
	defer agent.lockDevice.Unlock()
	log.Debugw("revertImage", log.Fields{"id": agent.deviceId})
	// Get the most up to date the device info
	if device, err := agent.getDeviceWithoutLock(); err != nil {
		return nil, status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		// Verify whether the Image is in the list of image being downloaded
		if !isImageRegistered(img, device) {
			return nil, status.Errorf(codes.FailedPrecondition, "deviceId:%s, image-not-registered:%s", agent.deviceId, img.Name)
		}

		if device.AdminState != voltha.AdminState_ENABLED {
			return nil, status.Errorf(codes.FailedPrecondition, "deviceId:%s, device-not-enabled-state:%s", agent.deviceId, img.Name)
		}
		// Update image download state
		cloned := proto.Clone(device).(*voltha.Device)
		for _, image := range cloned.ImageDownloads {
			if image.Id == img.Id && image.Name == img.Name {
				image.ImageState = voltha.ImageDownload_IMAGE_REVERTING
			}
		}
		if afterUpdate := agent.clusterDataProxy.Update("/devices/"+agent.deviceId, cloned, false, ""); afterUpdate == nil {
			return nil, status.Errorf(codes.Internal, "failed-update-device:%s", agent.deviceId)
		}

		if err := agent.adapterProxy.RevertImageUpdate(ctx, device, img); err != nil {
			log.Debugw("revertImage-error", log.Fields{"id": agent.lastData.Id, "error": err, "image": img.Name})
			return nil, err
		}
	}
	return &voltha.OperationResp{Code: voltha.OperationResp_OPERATION_SUCCESS}, nil
}

func (agent *DeviceAgent) getImageDownloadStatus(ctx context.Context, img *voltha.ImageDownload) (*voltha.ImageDownload, error) {
	agent.lockDevice.Lock()
	defer agent.lockDevice.Unlock()
	log.Debugw("getImageDownloadStatus", log.Fields{"id": agent.deviceId})
	// Get the most up to date the device info
	if device, err := agent.getDeviceWithoutLock(); err != nil {
		return nil, status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		if resp, err := agent.adapterProxy.GetImageDownloadStatus(ctx, device, img); err != nil {
			log.Debugw("getImageDownloadStatus-error", log.Fields{"id": agent.lastData.Id, "error": err, "image": img.Name})
			return nil, err
		} else {
			return resp, nil
		}
	}
}

func (agent *DeviceAgent) updateImageDownload(img *voltha.ImageDownload) error {
	agent.lockDevice.Lock()
	defer agent.lockDevice.Unlock()
	log.Debugw("updateImageDownload", log.Fields{"id": agent.deviceId})
	// Get the most up to date the device info
	if device, err := agent.getDeviceWithoutLock(); err != nil {
		return status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		// Update the image as well as remove it if the download was cancelled
		cloned := proto.Clone(device).(*voltha.Device)
		clonedImages := make([]*voltha.ImageDownload, len(cloned.ImageDownloads))
		for _, image := range cloned.ImageDownloads {
			if image.Id == img.Id && image.Name == img.Name {
				if image.DownloadState != voltha.ImageDownload_DOWNLOAD_CANCELLED {
					clonedImages = append(clonedImages, img)
				}
			}
		}
		cloned.ImageDownloads = clonedImages
		// Set the Admin state to enabled if required
		if (img.DownloadState != voltha.ImageDownload_DOWNLOAD_REQUESTED &&
			img.DownloadState != voltha.ImageDownload_DOWNLOAD_STARTED) ||
			(img.ImageState != voltha.ImageDownload_IMAGE_ACTIVATING) {
			cloned.AdminState = voltha.AdminState_ENABLED
		}

		if afterUpdate := agent.clusterDataProxy.Update("/devices/"+agent.deviceId, cloned, false, ""); afterUpdate == nil {
			return status.Errorf(codes.Internal, "failed-update-device:%s", agent.deviceId)
		}
	}
	return nil
}

func (agent *DeviceAgent) getImageDownload(ctx context.Context, img *voltha.ImageDownload) (*voltha.ImageDownload, error) {
	agent.lockDevice.RLock()
	defer agent.lockDevice.RUnlock()
	log.Debugw("getImageDownload", log.Fields{"id": agent.deviceId})
	// Get the most up to date the device info
	if device, err := agent.getDeviceWithoutLock(); err != nil {
		return nil, status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		for _, image := range device.ImageDownloads {
			if image.Id == img.Id && image.Name == img.Name {
				return image, nil
			}
		}
		return nil, status.Errorf(codes.NotFound, "image-not-found:%s", img.Name)
	}
}

func (agent *DeviceAgent) listImageDownloads(ctx context.Context, deviceId string) (*voltha.ImageDownloads, error) {
	agent.lockDevice.RLock()
	defer agent.lockDevice.RUnlock()
	log.Debugw("listImageDownloads", log.Fields{"id": agent.deviceId})
	// Get the most up to date the device info
	if device, err := agent.getDeviceWithoutLock(); err != nil {
		return nil, status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		return &voltha.ImageDownloads{Items: device.ImageDownloads}, nil
	}
}

// getPorts retrieves the ports information of the device based on the port type.
func (agent *DeviceAgent) getPorts(ctx context.Context, portType voltha.Port_PortType) *voltha.Ports {
	log.Debugw("getPorts", log.Fields{"id": agent.deviceId, "portType": portType})
	ports := &voltha.Ports{}
	if device, _ := agent.deviceMgr.GetDevice(agent.deviceId); device != nil {
		for _, port := range device.Ports {
			if port.Type == portType {
				ports.Items = append(ports.Items, port)
			}
		}
	}
	return ports
}

// getSwitchCapability is a helper method that a logical device agent uses to retrieve the switch capability of a
// parent device
func (agent *DeviceAgent) getSwitchCapability(ctx context.Context) (*ic.SwitchCapability, error) {
	log.Debugw("getSwitchCapability", log.Fields{"deviceId": agent.deviceId})
	if device, err := agent.deviceMgr.GetDevice(agent.deviceId); device == nil {
		return nil, err
	} else {
		var switchCap *ic.SwitchCapability
		var err error
		if switchCap, err = agent.adapterProxy.GetOfpDeviceInfo(ctx, device); err != nil {
			log.Debugw("getSwitchCapability-error", log.Fields{"id": device.Id, "error": err})
			return nil, err
		}
		return switchCap, nil
	}
}

// getPortCapability is a helper method that a logical device agent uses to retrieve the port capability of a
// device
func (agent *DeviceAgent) getPortCapability(ctx context.Context, portNo uint32) (*ic.PortCapability, error) {
	log.Debugw("getPortCapability", log.Fields{"deviceId": agent.deviceId})
	if device, err := agent.deviceMgr.GetDevice(agent.deviceId); device == nil {
		return nil, err
	} else {
		var portCap *ic.PortCapability
		var err error
		if portCap, err = agent.adapterProxy.GetOfpPortInfo(ctx, device, portNo); err != nil {
			log.Debugw("getPortCapability-error", log.Fields{"id": device.Id, "error": err})
			return nil, err
		}
		return portCap, nil
	}
}

func (agent *DeviceAgent) packetOut(outPort uint32, packet *ofp.OfpPacketOut) error {
	//	Send packet to adapter
	if err := agent.adapterProxy.packetOut(agent.deviceType, agent.deviceId, outPort, packet); err != nil {
		log.Debugw("packet-out-error", log.Fields{"id": agent.lastData.Id, "error": err})
		return err
	}
	return nil
}

// processUpdate is a callback invoked whenever there is a change on the device manages by this device agent
func (agent *DeviceAgent) processUpdate(args ...interface{}) interface{} {
	//// Run this callback in its own go routine
	go func(args ...interface{}) interface{} {
		var previous *voltha.Device
		var current *voltha.Device
		var ok bool
		if len(args) == 2 {
			if previous, ok = args[0].(*voltha.Device); !ok {
				log.Errorw("invalid-callback-type", log.Fields{"data": args[0]})
				return nil
			}
			if current, ok = args[1].(*voltha.Device); !ok {
				log.Errorw("invalid-callback-type", log.Fields{"data": args[1]})
				return nil
			}
		} else {
			log.Errorw("too-many-args-in-callback", log.Fields{"len": len(args)})
			return nil
		}
		// Perform the state transition in it's own go routine
		if err := agent.deviceMgr.processTransition(previous, current); err != nil {
			log.Errorw("failed-process-transition", log.Fields{"deviceId": previous.Id,
				"previousAdminState": previous.AdminState, "currentAdminState": current.AdminState})
		}
		return nil
	}(args...)

	return nil
}

func (agent *DeviceAgent) updateDevice(device *voltha.Device) error {
	agent.lockDevice.Lock()
	defer agent.lockDevice.Unlock()
	log.Debugw("updateDevice", log.Fields{"deviceId": device.Id})
	cloned := proto.Clone(device).(*voltha.Device)
	afterUpdate := agent.clusterDataProxy.Update("/devices/"+device.Id, cloned, false, "")
	if afterUpdate == nil {
		return status.Errorf(codes.Internal, "%s", device.Id)
	}
	return nil
}

func (agent *DeviceAgent) updateDeviceWithoutLock(device *voltha.Device) error {
	log.Debugw("updateDevice", log.Fields{"deviceId": device.Id})
	cloned := proto.Clone(device).(*voltha.Device)
	afterUpdate := agent.clusterDataProxy.Update("/devices/"+device.Id, cloned, false, "")
	if afterUpdate == nil {
		return status.Errorf(codes.Internal, "%s", device.Id)
	}
	return nil
}

func (agent *DeviceAgent) updateDeviceStatus(operStatus voltha.OperStatus_OperStatus, connStatus voltha.ConnectStatus_ConnectStatus) error {
	agent.lockDevice.Lock()
	// Work only on latest data
	if storeDevice, err := agent.getDeviceWithoutLock(); err != nil {
		agent.lockDevice.Unlock()
		return status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		// clone the device
		cloned := proto.Clone(storeDevice).(*voltha.Device)
		// Ensure the enums passed in are valid - they will be invalid if they are not set when this function is invoked
		if s, ok := voltha.ConnectStatus_ConnectStatus_value[connStatus.String()]; ok {
			log.Debugw("updateDeviceStatus-conn", log.Fields{"ok": ok, "val": s})
			cloned.ConnectStatus = connStatus
		}
		if s, ok := voltha.OperStatus_OperStatus_value[operStatus.String()]; ok {
			log.Debugw("updateDeviceStatus-oper", log.Fields{"ok": ok, "val": s})
			cloned.OperStatus = operStatus
		}
		log.Debugw("updateDeviceStatus", log.Fields{"deviceId": cloned.Id, "operStatus": cloned.OperStatus, "connectStatus": cloned.ConnectStatus})
		// Store the device
		if afterUpdate := agent.clusterDataProxy.Update("/devices/"+agent.deviceId, cloned, false, ""); afterUpdate == nil {
			agent.lockDevice.Unlock()
			return status.Errorf(codes.Internal, "%s", agent.deviceId)
		}
		agent.lockDevice.Unlock()
		return nil
	}
}

func (agent *DeviceAgent) enablePorts() error {
	agent.lockDevice.Lock()
	defer agent.lockDevice.Unlock()
	if storeDevice, err := agent.getDeviceWithoutLock(); err != nil {
		return status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		// clone the device
		cloned := proto.Clone(storeDevice).(*voltha.Device)
		for _, port := range cloned.Ports {
			port.AdminState = voltha.AdminState_ENABLED
			port.OperStatus = voltha.OperStatus_ACTIVE
		}
		// Store the device
		if afterUpdate := agent.clusterDataProxy.Update("/devices/"+agent.deviceId, cloned, false, ""); afterUpdate == nil {
			return status.Errorf(codes.Internal, "%s", agent.deviceId)
		}
		return nil
	}
}

func (agent *DeviceAgent) disablePorts() error {
	agent.lockDevice.Lock()
	defer agent.lockDevice.Unlock()
	if storeDevice, err := agent.getDeviceWithoutLock(); err != nil {
		return status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		// clone the device
		cloned := proto.Clone(storeDevice).(*voltha.Device)
		for _, port := range cloned.Ports {
			port.AdminState = voltha.AdminState_DISABLED
			port.OperStatus = voltha.OperStatus_UNKNOWN
		}
		// Store the device
		if afterUpdate := agent.clusterDataProxy.Update("/devices/"+agent.deviceId, cloned, false, ""); afterUpdate == nil {
			return status.Errorf(codes.Internal, "%s", agent.deviceId)
		}
		return nil
	}
}

func (agent *DeviceAgent) updatePortState(portType voltha.Port_PortType, portNo uint32, operStatus voltha.OperStatus_OperStatus) error {
	agent.lockDevice.Lock()
	// Work only on latest data
	// TODO: Get list of ports from device directly instead of the entire device
	if storeDevice, err := agent.getDeviceWithoutLock(); err != nil {
		agent.lockDevice.Unlock()
		return status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		// clone the device
		cloned := proto.Clone(storeDevice).(*voltha.Device)
		// Ensure the enums passed in are valid - they will be invalid if they are not set when this function is invoked
		if _, ok := voltha.Port_PortType_value[portType.String()]; !ok {
			agent.lockDevice.Unlock()
			return status.Errorf(codes.InvalidArgument, "%s", portType)
		}
		for _, port := range cloned.Ports {
			if port.Type == portType && port.PortNo == portNo {
				port.OperStatus = operStatus
				// Set the admin status to ENABLED if the operational status is ACTIVE
				// TODO: Set by northbound system?
				if operStatus == voltha.OperStatus_ACTIVE {
					port.AdminState = voltha.AdminState_ENABLED
				}
				break
			}
		}
		log.Debugw("portStatusUpdate", log.Fields{"deviceId": cloned.Id})
		// Store the device
		if afterUpdate := agent.clusterDataProxy.Update("/devices/"+agent.deviceId, cloned, false, ""); afterUpdate == nil {
			agent.lockDevice.Unlock()
			return status.Errorf(codes.Internal, "%s", agent.deviceId)
		}
		agent.lockDevice.Unlock()
		return nil
	}
}

func (agent *DeviceAgent) updatePmConfigs(pmConfigs *voltha.PmConfigs) error {
	agent.lockDevice.Lock()
	defer agent.lockDevice.Unlock()
	log.Debug("updatePmConfigs")
	// Work only on latest data
	if storeDevice, err := agent.getDeviceWithoutLock(); err != nil {
		return status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		// clone the device
		cloned := proto.Clone(storeDevice).(*voltha.Device)
		cloned.PmConfigs = proto.Clone(pmConfigs).(*voltha.PmConfigs)
		// Store the device
		afterUpdate := agent.clusterDataProxy.Update("/devices/"+agent.deviceId, cloned, false, "")
		if afterUpdate == nil {
			return status.Errorf(codes.Internal, "%s", agent.deviceId)
		}
		return nil
	}
}

func (agent *DeviceAgent) addPort(port *voltha.Port) error {
	agent.lockDevice.Lock()
	defer agent.lockDevice.Unlock()
	log.Debugw("addLogicalPortToMap", log.Fields{"deviceId": agent.deviceId})
	// Work only on latest data
	if storeDevice, err := agent.getDeviceWithoutLock(); err != nil {
		return status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		// clone the device
		cloned := proto.Clone(storeDevice).(*voltha.Device)
		if cloned.Ports == nil {
			//	First port
			log.Debugw("addLogicalPortToMap-first-port-to-add", log.Fields{"deviceId": agent.deviceId})
			cloned.Ports = make([]*voltha.Port, 0)
		} else {
			for _, p := range cloned.Ports {
				if p.Type == port.Type && p.PortNo == port.PortNo {
					log.Debugw("port already exists", log.Fields{"port": *port})
					return nil
				}
			}

		}
		cp := proto.Clone(port).(*voltha.Port)
		// Set the admin state of the port to ENABLE if the operational state is ACTIVE
		// TODO: Set by northbound system?
		if cp.OperStatus == voltha.OperStatus_ACTIVE {
			cp.AdminState = voltha.AdminState_ENABLED
		}
		cloned.Ports = append(cloned.Ports, cp)
		// Store the device
		afterUpdate := agent.clusterDataProxy.Update("/devices/"+agent.deviceId, cloned, false, "")
		if afterUpdate == nil {
			return status.Errorf(codes.Internal, "%s", agent.deviceId)
		}
		return nil
	}
}

func (agent *DeviceAgent) addPeerPort(port *voltha.Port_PeerPort) error {
	agent.lockDevice.Lock()
	defer agent.lockDevice.Unlock()
	log.Debug("addPeerPort")
	// Work only on latest data
	if storeDevice, err := agent.getDeviceWithoutLock(); err != nil {
		return status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		// clone the device
		cloned := proto.Clone(storeDevice).(*voltha.Device)
		// Get the peer port on the device based on the port no
		for _, peerPort := range cloned.Ports {
			if peerPort.PortNo == port.PortNo { // found port
				cp := proto.Clone(port).(*voltha.Port_PeerPort)
				peerPort.Peers = append(peerPort.Peers, cp)
				log.Debugw("found-peer", log.Fields{"portNo": port.PortNo, "deviceId": agent.deviceId})
				break
			}
		}
		// Store the device
		afterUpdate := agent.clusterDataProxy.Update("/devices/"+agent.deviceId, cloned, false, "")
		if afterUpdate == nil {
			return status.Errorf(codes.Internal, "%s", agent.deviceId)
		}
		return nil
	}
}

// TODO: A generic device update by attribute
func (agent *DeviceAgent) updateDeviceAttribute(name string, value interface{}) {
	agent.lockDevice.Lock()
	defer agent.lockDevice.Unlock()
	if value == nil {
		return
	}
	var storeDevice *voltha.Device
	var err error
	if storeDevice, err = agent.getDeviceWithoutLock(); err != nil {
		return
	}
	updated := false
	s := reflect.ValueOf(storeDevice).Elem()
	if s.Kind() == reflect.Struct {
		// exported field
		f := s.FieldByName(name)
		if f.IsValid() && f.CanSet() {
			switch f.Kind() {
			case reflect.String:
				f.SetString(value.(string))
				updated = true
			case reflect.Uint32:
				f.SetUint(uint64(value.(uint32)))
				updated = true
			case reflect.Bool:
				f.SetBool(value.(bool))
				updated = true
			}
		}
	}
	log.Debugw("update-field-status", log.Fields{"deviceId": storeDevice.Id, "name": name, "updated": updated})
	//	Save the data
	cloned := proto.Clone(storeDevice).(*voltha.Device)
	if afterUpdate := agent.clusterDataProxy.Update("/devices/"+agent.deviceId, cloned, false, ""); afterUpdate == nil {
		log.Warnw("attribute-update-failed", log.Fields{"attribute": name, "value": value})
	}
	return
}

func (agent *DeviceAgent) simulateAlarm(ctx context.Context, simulatereq *voltha.SimulateAlarmRequest) error {
	agent.lockDevice.Lock()
	defer agent.lockDevice.Unlock()
	log.Debugw("simulateAlarm", log.Fields{"id": agent.deviceId})
	// Get the most up to date the device info
	if device, err := agent.getDeviceWithoutLock(); err != nil {
		return status.Errorf(codes.NotFound, "%s", agent.deviceId)
	} else {
		// First send the request to an Adapter and wait for a response
		if err := agent.adapterProxy.SimulateAlarm(ctx, device, simulatereq); err != nil {
			log.Debugw("simulateAlarm-error", log.Fields{"id": agent.lastData.Id, "error": err})
			return err
		}
	}
	return nil
}

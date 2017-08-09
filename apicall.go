package p2pubmachine

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/iij/p2pubapi"
	"github.com/iij/p2pubapi/protocol"
)

var endpoint = os.Getenv("IIJAPI_ENDPOINT")
var insecure, _ = strconv.ParseBool(os.Getenv("IIJAPI_INSECURE"))

func (d *Driver) callapi(arg protocol.CommonArg, res interface{}) (err error) {
	api := p2pubapi.NewAPI(d.AccessKey, d.SecretKey)
	if endpoint != "" {
		api.Endpoint = endpoint
	}
	api.Insecure = insecure
	if err = p2pubapi.Call(*api, arg, &res); err != nil {
		log.Error(err)
	}
	return
}

func (d *Driver) createvm() (err error) {
	log.Infof("creating VM")
	vmadd := protocol.VMAdd{
		GisServiceCode: d.GisServiceCode,
		OSType:         osType,
		Type:           d.ServerType,
		ServerGroup:    d.ServerGroup,
	}
	var res = protocol.VMAddResponse{}
	if err = d.callapi(vmadd, &res); err != nil {
		return
	}
	d.IvmServiceCode = res.ServiceCode
	return
}

func (d *Driver) getstatus(typ, servicecode string) (contract, resource string, err error) {
	switch typ {
	case "vm":
		if servicecode == "" {
			servicecode = d.IvmServiceCode
		}
		arg := protocol.VMGet{
			GisServiceCode: d.GisServiceCode,
			IvmServiceCode: servicecode,
		}
		var res = protocol.VMGetResponse{}
		if err = d.callapi(arg, &res); err != nil {
			return
		}
		contract = res.ContractStatus
		resource = res.ResourceStatus
	case "systemstorage":
		if servicecode == "" {
			servicecode = d.IbaServiceCode
		}
		arg := protocol.SystemStorageGet{
			GisServiceCode: d.GisServiceCode,
			IbaServiceCode: servicecode,
		}
		var res = protocol.SystemStorageGetResponse{}
		if err = d.callapi(arg, &res); err != nil {
			return
		}
		contract = res.ContractStatus
		resource = res.ResourceStatus
	case "datastorage":
		arg := protocol.StorageGet{
			GisServiceCode: d.GisServiceCode,
			IbgServiceCode: servicecode,
		}
		var res = protocol.StorageGetResponse{}
		if err = d.callapi(arg, &res); err != nil {
			return
		}
		contract = res.ContractStatus
		resource = res.ResourceStatus
	}
	return
}

func (d *Driver) waitstatus(typ, servicecode, cstatus, rstatus string) error {
	t := time.Now()
	defer func() {
		log.Infof("wait %s (%s/%s) done %v", typ, cstatus, rstatus, time.Since(t))
	}()
	for {
		log.Debugf("check status %s %s (%s/%s)", typ, servicecode, cstatus, rstatus)
		contract, resource, err := d.getstatus(typ, servicecode)
		if err != nil {
			return err
		}
		log.Debugf("check result: %s/%s (%s/%s)", contract, resource, cstatus, rstatus)
		if (rstatus == "" || resource == rstatus) && (cstatus == "" || contract == cstatus) {
			return nil
		}
		time.Sleep(10 * time.Second)
	}
}

func (d *Driver) createdisk() (err error) {
	log.Infof("creating VM disk")
	vmdisk := protocol.SystemStorageAdd{
		GisServiceCode: d.GisServiceCode,
		StorageGroup:   d.StorageGroup,
		Type:           d.ImageName,
	}
	var res = protocol.StorageAddResponse{}
	if err = d.callapi(vmdisk, &res); err != nil {
		return
	}
	d.IbaServiceCode = res.ServiceCode
	return
}

func (d *Driver) createdatadisk() (err error) {
	log.Infof("creating data disk")
	datadisk := protocol.StorageAdd{
		GisServiceCode: d.GisServiceCode,
		Type:           d.addStorageType,
		StorageGroup:   d.StorageGroup,
	}
	var res = protocol.StorageAddResponse{}
	if err = d.callapi(datadisk, &res); err != nil {
		return
	}
	d.IbbServiceCode = res.ServiceCode
	return
}

func (d *Driver) attachdisk() (err error) {
	log.Infof("attach system storage %s <- %s", d.IvmServiceCode, d.IbaServiceCode)
	vmattach := protocol.BootDeviceStorageConnect{
		GisServiceCode: d.GisServiceCode,
		IvmServiceCode: d.IvmServiceCode,
		IbaServiceCode: d.IbaServiceCode,
	}
	var res = protocol.BootDeviceStorageConnectResponse{}
	if err = d.callapi(vmattach, &res); err != nil {
		return
	}
	if d.IbbServiceCode != "" {
		argD := protocol.DataDeviceStorageConnect{
			GisServiceCode: d.GisServiceCode,
			IvmServiceCode: d.IvmServiceCode,
		}
		if strings.HasPrefix(d.IbbServiceCode, "ibb") {
			argD.IbbServiceCode = d.IbbServiceCode
		} else if strings.HasPrefix(d.IbbServiceCode, "ibg") {
			argD.IbgServiceCode = d.IbbServiceCode
		} else if strings.HasPrefix(d.IbbServiceCode, "iba") {
			argD.IbaServiceCode = d.IbbServiceCode
		}
		var resD = protocol.DataDeviceStorageConnectResponse{}
		if err = d.callapi(argD, &resD); err != nil {
			return
		}
	}
	return
}

func (d *Driver) attachip() (err error) {
	log.Infof("attach IP address %s", d.IvmServiceCode)
	vmip := protocol.GlobalAddressAllocate{
		GisServiceCode: d.GisServiceCode,
		IvmServiceCode: d.IvmServiceCode,
	}
	var res = protocol.GlobalAddressAllocateResponse{}
	if err = d.callapi(vmip, &res); err != nil {
		return
	}
	d.IPv4 = res.IPv4.IpAddress
	d.IPv6 = res.IPv6.IpAddress
	d.IPAddress = d.IPv4
	log.Infof("IP address: %s %s", d.IPv4, d.IPv6)
	return
}

func (d *Driver) getip() (v4 []string, v6 []string) {
	arg := protocol.VMGet{
		GisServiceCode: d.GisServiceCode,
		IvmServiceCode: d.IvmServiceCode,
	}
	var res = protocol.VMGetResponse{}
	if err := d.callapi(arg, &res); err != nil {
		return
	}
	for _, v := range res.NetworkList {
		for _, i := range v.IpAddressList {
			if i.IPv4.IpAddress != "" {
				v4 = append(v4, i.IPv4.IpAddress)
			}
			if i.IPv6.IpAddress != "" {
				v6 = append(v6, i.IPv6.IpAddress)
			}
		}
	}
	return
}

func (d *Driver) setpubkey(key string) (err error) {
	log.Infof("setting public key")
	pubkey := protocol.PublicKeyAdd{
		GisServiceCode: d.GisServiceCode,
		IbaServiceCode: d.IbaServiceCode,
		PublicKey:      key,
	}
	var res = protocol.PublicKeyAddResponse{}
	if err = d.callapi(pubkey, &res); err != nil {
		return err
	}
	return
}

// vmpower On/Off/Reset, and wait Stopped/Configuring/Starting/Running/Stopping/Locked
func (d *Driver) vmpower(arg, waitto string) error {
	log.Infof("VM power %s", arg)
	vmpower := protocol.VMPower{
		GisServiceCode: d.GisServiceCode,
		IvmServiceCode: d.IvmServiceCode,
		Power:          arg,
	}
	var res = protocol.VMPowerResponse{}
	if err := d.callapi(vmpower, &res); err != nil {
		return err
	}
	return d.waitstatus("vm", "", "", waitto)
}

func (d *Driver) removevm() error {
	log.Infof("removing vm and system storage")
	d.vmpower("Off", "Stopped")
	arg := protocol.VMCancel{
		GisServiceCode: d.GisServiceCode,
		IvmServiceCode: d.IvmServiceCode,
	}
	var res = protocol.VMCancelResponse{}
	if err := d.callapi(arg, &res); err != nil {
		return err
	}
	dargS := protocol.SystemStorageCancel{
		GisServiceCode: d.GisServiceCode,
		IbaServiceCode: d.IbaServiceCode,
	}
	var dresS = protocol.SystemStorageCancelResponse{}
	if err := d.callapi(dargS, &dresS); err != nil {
		return err
	}
	if d.IbbServiceCode != "" {
		dargD := protocol.StorageCancel{
			GisServiceCode: d.GisServiceCode,
			IbgServiceCode: d.IbbServiceCode,
		}
		var dresD = protocol.StorageCancelResponse{}
		if err := d.callapi(dargD, &dresD); err != nil {
			return err
		}
	}
	return nil
}

func (d *Driver) restore(iarServiceCode string, imageId string) error {
	log.Info("Restore system storage.")
	arg := protocol.Restore{
		GisServiceCode: d.GisServiceCode,
		IbaServiceCode: d.IbaServiceCode,
		ImageId: imageId,
		IarServiceCode: iarServiceCode,
		Image: "Archive",
	}
	var res = protocol.RestoreResponse{}
	return d.callapi(arg, &res)
}

func (d *Driver) setVMlabel(l string) error {
	log.Info("setting vm label", l)
	arg := protocol.VMLabelSet{
		GisServiceCode: d.GisServiceCode,
		IvmServiceCode: d.IvmServiceCode,
		Name:           l,
	}
	var res = protocol.VMLabelSetResponse{}
	return d.callapi(arg, &res)
}

func (d *Driver) setSysStlabel(l string) error {
	log.Info("setting system storage label", l)
	arg := protocol.SystemStorageLabelSet{
		GisServiceCode: d.GisServiceCode,
		IbaServiceCode: d.IbaServiceCode,
		Name:           l,
	}
	var res = protocol.SystemStorageLabelSetResponse{}
	return d.callapi(arg, &res)
}

func (d *Driver) setDataStlabel(l string) error {
	log.Info("setting data storage label", l)
	arg := protocol.StorageLabelSet{
		GisServiceCode: d.GisServiceCode,
		IbgServiceCode: d.IbbServiceCode,
		Name:           l,
	}
	var res = protocol.StorageLabelSetResponse{}
	return d.callapi(arg, &res)
}

func (d *Driver) sshCommand(cmd string) (res string, err error) {
	log.Debug("execute command:", cmd)
	return drivers.RunSSHCommandFromDriver(d, cmd)
}

func (d *Driver) sshCommands(cmd []string) (res []string, err error) {
	for _, v := range cmd {
		var r string
		if r, err = d.sshCommand(v); err != nil {
			return
		}
		res = append(res, r)
	}
	return
}

package p2pubmachine

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/iij/docker-machine-driver-p2pub/oscmd"
)

const (
	imageName    = "S30GB_UBUNTU14_64"
	storageGroup = ""
	serverGroup  = ""
	serverType   = "VB0-1"
	osType       = "Linux"
	addDev       = "/dev/vdb"
)

// Driver for P2PUB
type Driver struct {
	*drivers.BaseDriver
	GisServiceCode string
	AccessKey      string
	SecretKey      string
	IvmServiceCode string
	IbaServiceCode string
	IPv4           string
	IPv6           string
	DockerPort     int
	ServerType     string
	ImageName      string
	StorageGroup   string
	ServerGroup    string
	IbbServiceCode string
	privateMode    string
	addStorageType string
	baseImage      string
}

type openport struct {
	port  int
	proto string
}

var ports = []openport{
	{2377, "tcp"}, // swarm
	{7946, "udp"}, // swarm object store
	{7946, "tcp"}, // swarm object store
	{4789, "udp"}, // vxlan for overlay network
}

// NewDriver is constructor
func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

// DriverName returns name of driver(p2pub)
func (d *Driver) DriverName() string {
	return "p2pub"
}

// PreCreateCheck checks required options
func (d *Driver) PreCreateCheck() error {
	log.Debugf("precreate %#+v", d)
	if d.AccessKey == "" || d.SecretKey == "" || d.GisServiceCode == "" {
		return fmt.Errorf("p2pub requires accesskey/secretkey/gisservicecode")
	}
	if d.privateMode != "" && len(strings.Split(d.privateMode, ",")) != 2 {
		return fmt.Errorf("option format: --p2pub-private-only defgw,dns")
	}
	if d.baseImage != "" && len(strings.Split(d.baseImage, ",")) != 2 {                                           
		return fmt.Errorf("option format: --p2pub-custom-image iarservicecode,imageid")
	}
  	return nil
}

// Create machine
func (d *Driver) Create() (err error) {
	if d.IvmServiceCode == "" {
		if err = d.createvm(); err != nil {
			return
		}
		if err = d.waitstatus("vm", d.IvmServiceCode, "InService", "Stopped"); err != nil {
			return
		}
		if err = d.setVMlabel(d.MachineName); err != nil {
			log.Warn("set label of VM failed:", err)
		}
	}
	if d.IbaServiceCode == "" {
		if err = d.createdisk(); err != nil {
			return
		}
		if err = d.waitstatus("systemstorage", d.IbaServiceCode, "InService", "NotAttached"); err != nil {
			return
		}
		if err = d.setSysStlabel(d.MachineName); err != nil {
			log.Warn("set label of SystemStorage failed:", err)
		}
		if err = ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
			log.Error(err)
			return
		}
		var publicKey []byte
		publicKey, err = ioutil.ReadFile(d.GetSSHKeyPath() + ".pub")
		if err = d.setpubkey(string(publicKey)); err != nil {
			return
		}
		if err = d.waitstatus("systemstorage", d.IbaServiceCode, "InService", "NotAttached"); err != nil {
			return
		}
	}
	if d.baseImage != "" {
		restoreParams := strings.Split(d.baseImage, ",")
		if err = d.restore(restoreParams[0], restoreParams[1]); err != nil {
 			return
 		}
 		if err = d.waitstatus("systemstorage", d.IbaServiceCode, "InService", "NotAttached"); err != nil {
 			return
 		}
	}
	if d.IbbServiceCode == "" && d.addStorageType != "" {
		if err = d.createdatadisk(); err != nil {
			return
		}
		if err = d.waitstatus("datastorage", d.IbbServiceCode, "InService", "NotAttached"); err != nil {
			return
		}
		if err = d.setDataStlabel(d.MachineName); err != nil {
			log.Warn("set label of Storage failed:", err)
		}
	}
	if err = d.attachdisk(); err != nil {
		return
	}
	if err = d.waitstatus("systemstorage", d.IbaServiceCode, "InService", "Attached"); err != nil {
		return
	}
	if d.IbbServiceCode != "" {
		if err = d.waitstatus("datastorage", d.IbbServiceCode, "InService", "Attached"); err != nil {
			return
		}
	}
	if d.privateMode == "" {
		if err = d.attachip(); err != nil {
			return
		}
	} else {
		v4, v6 := d.getip()
		if len(v4) != 0 {
			d.IPv4 = v4[0]
		}
		if len(v6) != 0 {
			d.IPv6 = v6[0]
		}
		d.IPAddress = d.IPv4
		log.Info("Private mode:", d.privateMode, v4, v6, d.IPv4, d.IPv6)
	}
	if err = d.vmpower("On", "Running"); err != nil {
		return
	}
	var cmd oscmd.Oscmd
	prov, err := provision.DetectProvisioner(d)
	if err != nil {
		return
	}
	osr, err := prov.GetOsReleaseInfo()
	if err != nil {
		return
	}
	log.Info(osr.PrettyName)
	switch osr.ID {
	case "centos":
		cmd = oscmd.CentOS{}
	case "rhel":
		cmd = oscmd.RedHat{}
	case "ubuntu":
		cmd = oscmd.Ubuntu{}
	case "debian":
		cmd = oscmd.Debian{}
	default:
		cmd = oscmd.CentOS{}
	}
	log.Debugf("%T", cmd)
	var res []string
	if res, err = d.osInit(cmd); err != nil {
		log.Error(err)
		return
	}
	if err = drivers.WaitForSSH(d); err != nil {
		return
	}
	log.Debug("execute:", res)
	_, err = d.sshCommands(res)
	return
}

func (d *Driver) osInit(cmd oscmd.Oscmd) (res []string, err error) {
	log.Infof("open port %d (for docker)", d.DockerPort)
	res = cmd.OpenFW(d.DockerPort, "tcp")
	log.Debug("cmd1:", res)
	for _, v := range ports {
		log.Infof("open port %d/%s", v.port, v.proto)
		res = append(res, cmd.OpenFW(v.port, v.proto)...)
	}

	netinfo := strings.Split(d.privateMode, ",")
	if len(netinfo) == 2 {
		log.Infof("set default gateway: %s", netinfo[0])
		res = append(res, cmd.DefGW(netinfo[0])...)
		log.Infof("set dns: %s", netinfo[1])
		res = append(res, cmd.DNS([]string{netinfo[1]})...)
		log.Debug("cmd3:", res)
	}
	return
}

// Kill machine
func (d *Driver) Kill() (err error) {
	return d.vmpower("Off", "Stopped")
}

// Restart machine
func (d *Driver) Restart() (err error) {
	return d.vmpower("Reset", "Running")
}

// GetURL returns URL for DOCKER_HOST
func (d *Driver) GetURL() (string, error) {
	if err := drivers.MustBeRunning(d); err != nil {
		return "", err
	}
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, strconv.Itoa(d.DockerPort))), nil
}

// GetState returns state of machine
func (d *Driver) GetState() (st state.State, err error) {
	var res string
	if _, res, err = d.getstatus("vm", ""); err != nil {
		return state.Error, err
	}
	switch res {
	case "Stopped":
		return state.Stopped, nil
	case "Configuring":
		return state.Paused, nil
	case "Starting":
		return state.Starting, nil
	case "Running":
		return state.Running, nil
	case "Stopping":
		return state.Stopping, nil
	case "Locked":
		return state.Paused, nil
	}
	return state.Error, fmt.Errorf("unknown state: %s", res)
}

// GetCreateFlags defines option for docker-machine command
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   "p2pub-access-key",
			EnvVar: "IIJAPI_ACCESS_KEY",
			Usage:  "p2pub access key",
		},
		mcnflag.StringFlag{
			Name:   "p2pub-secret-key",
			EnvVar: "IIJAPI_SECRET_KEY",
			Usage:  "p2pub secret key",
		},
		mcnflag.StringFlag{
			Name:   "p2pub-gis",
			EnvVar: "GisServiceCode,GISSERVICECODE",
			Usage:  "p2pub gis service code",
		},
		mcnflag.StringFlag{
			Name:   "p2pub-ivm",
			EnvVar: "IvmServiceCode,IVMSERVICECODE",
			Usage:  "p2pub ivm service code (for VM)",
		},
		mcnflag.StringFlag{
			Name:   "p2pub-iba",
			EnvVar: "IbaServiceCode,IBASERVICECODE",
			Usage:  "p2pub iba service code (for system storage)",
		},
		mcnflag.StringFlag{
			Name:   "p2pub-ibb",
			EnvVar: "IbbServiceCode,IBBSERVICECODE",
			Usage:  "p2pub ibb service code (for data storage)",
		},
		mcnflag.IntFlag{
			Name:  "p2pub-docker-port",
			Value: 2376,
			Usage: "p2pub Docker Port",
		},
		mcnflag.StringFlag{
			Name:  "p2pub-server-type",
			Value: serverType,
			Usage: "p2pub serverType (http://manual.iij.jp/p2/pubapi/59949011.html)",
		},
		mcnflag.StringFlag{
			Name:  "p2pub-server-group",
			Value: serverGroup,
			Usage: "p2pub serverGroup [A|B](http://manual.iij.jp/p2/pubapi/59939383.html)",
		},
		mcnflag.StringFlag{
			Name:  "p2pub-system-storage",
			Value: imageName,
			Usage: "p2pub system storage (http://manual.iij.jp/p2/pubapi/59949023.html)",
		},
		mcnflag.StringFlag{
			Name:  "p2pub-storage-group",
			Value: storageGroup,
			Usage: "p2pub storage group [Y|Z](http://manual.iij.jp/p2/pubapi/59939812.html)",
		},
		mcnflag.StringFlag{
			Name:  "p2pub-data-storage",
			Value: "",
			Usage: "p2pub data storage (http://manual.iij.jp/p2/pubapi/59949023.html)",
		},
		mcnflag.StringFlag{
			Name:  "p2pub-private-only",
			Usage: "uses private network (does not attach global IP and setup defgw/dns)",
		},
		mcnflag.StringFlag{
			Name:  "p2pub-custom-image",
			Usage: "create system storage from custom image (http://manual.iij.jp/p2/pubapi/59940054.html)",
		},
	}
}

// SetConfigFromFlags set option to driver structure
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.AccessKey = flags.String("p2pub-access-key")
	d.SecretKey = flags.String("p2pub-secret-key")
	d.GisServiceCode = flags.String("p2pub-gis")
	d.IvmServiceCode = flags.String("p2pub-ivm")
	d.IbaServiceCode = flags.String("p2pub-iba")
	d.IbbServiceCode = flags.String("p2pub-ibb")
	d.DockerPort = flags.Int("p2pub-docker-port")
	d.ServerType = flags.String("p2pub-server-type")
	d.ServerGroup = flags.String("p2pub-server-group")
	d.StorageGroup = flags.String("p2pub-storage-group")
	d.ImageName = flags.String("p2pub-system-storage")
	d.privateMode = flags.String("p2pub-private-only")
	d.addStorageType = flags.String("p2pub-data-storage")
	d.baseImage = flags.String("p2pub-custom-image")
	d.SetSwarmConfigFromFlags(flags)
	if d.AccessKey == "" || d.SecretKey == "" || d.GisServiceCode == "" {
		return fmt.Errorf("p2pub driver requires --p2pub-{access,secret}-key, --p2pub-gis option: %+v", d)
	}
	return nil
}

// GetSSHHostname returns IP address of machine
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// Remove machine
func (d *Driver) Remove() error {
	return d.removevm()
}

// Start machine
func (d *Driver) Start() error {
	return d.vmpower("On", "Running")
}

// Stop machine
func (d *Driver) Stop() error {
	return d.vmpower("Off", "Stopped")
}

func (d *Driver) swarmPort() (int, error) {
	u, err := url.Parse(d.SwarmHost)
	if err != nil {
		return 0, err
	}
	_, p, err := net.SplitHostPort(u.Host)
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(p)
	return port, err
}

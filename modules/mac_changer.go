package modules

import (
	"fmt"
	"net"
	"runtime"
	"strings"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/log"
	bnet "github.com/evilsocket/bettercap-ng/net"
	"github.com/evilsocket/bettercap-ng/session"
)

type MacChanger struct {
	session.SessionModule
	originalMac net.HardwareAddr
	fakeMac     net.HardwareAddr
}

func NewMacChanger(s *session.Session) *MacChanger {
	mc := &MacChanger{
		SessionModule: session.NewSessionModule("mac.changer", s),
	}

	mc.AddParam(session.NewStringParameter("mac.changer.address",
		session.ParamRandomMAC,
		"[a-fA-F0-9]{2}:[a-fA-F0-9]{2}:[a-fA-F0-9]{2}:[a-fA-F0-9]{2}:[a-fA-F0-9]{2}:[a-fA-F0-9]{2}",
		"Hardware address to apply to the interface."))

	mc.AddHandler(session.NewModuleHandler("mac.changer on", "",
		"Start mac changer module.",
		func(args []string) error {
			return mc.Start()
		}))

	mc.AddHandler(session.NewModuleHandler("mac.changer off", "",
		"Stop mac changer module and restore original mac address.",
		func(args []string) error {
			return mc.Stop()
		}))

	return mc
}

func (mc *MacChanger) Name() string {
	return "mac.changer"
}

func (mc *MacChanger) Description() string {
	return "Change active interface mac address."
}

func (mc *MacChanger) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (mc *MacChanger) Configure() (err error) {
	var changeTo string

	if err, changeTo = mc.StringParam("mac.changer.address"); err != nil {
		return err
	}

	changeTo = bnet.NormalizeMac(changeTo)
	if mc.fakeMac, err = net.ParseMAC(changeTo); err != nil {
		return err
	}

	mc.originalMac = mc.Session.Interface.HW

	return nil
}

func (mc *MacChanger) setMac(mac net.HardwareAddr) error {
	os := runtime.GOOS
	args := []string{}

	if strings.Contains(os, "bsd") || os == "darwin" {
		args = []string{mc.Session.Interface.Name(), "ether", mac.String()}
	} else if os == "linux" {
		args = []string{mc.Session.Interface.Name(), "hw", "ether", mac.String()}
	} else {
		return fmt.Errorf("OS %s not supported by mac.changer module.", os)
	}

	_, err := core.Exec("ifconfig", args)
	if err == nil {
		mc.Session.Interface.HW = mac
	}

	return err
}

func (mc *MacChanger) Start() error {
	if mc.Running() == true {
		return session.ErrAlreadyStarted
	} else if err := mc.Configure(); err != nil {
		return err
	} else if err := mc.setMac(mc.fakeMac); err != nil {
		return err
	}

	mc.SetRunning(true)
	log.Info("Interface mac address set to %s", core.Bold(mc.fakeMac.String()))

	return nil
}

func (mc *MacChanger) Stop() error {
	if mc.Running() == false {
		return session.ErrAlreadyStopped
	} else if err := mc.setMac(mc.originalMac); err != nil {
		return err
	}

	mc.SetRunning(false)
	log.Info("Interface mac address restored to %s", core.Bold(mc.originalMac.String()))
	return nil
}

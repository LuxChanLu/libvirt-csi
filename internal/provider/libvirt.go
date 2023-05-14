package provider

import (
	"net/url"

	"github.com/LuxChanLu/csi-libvirt/internal/provider/config"
	"github.com/digitalocean/go-libvirt"
	"github.com/digitalocean/go-libvirt/socket"
	"github.com/digitalocean/go-libvirt/socket/dialers"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func ProvideLibvirt(lc fx.Lifecycle, log *zap.Logger, dialer socket.Dialer, config *config.Config) *libvirt.Libvirt {
	virt := libvirt.NewWithDialer(dialer)
	lc.Append(fx.StartStopHook(func() error {
		if err := virt.Connect(); err != nil {
			return err
		}
		version, err := virt.ConnectGetLibVersion()
		if err != nil {
			return err
		}
		log.Info("libvirt connected", zap.Uint64("version", version))
		return nil
	}, virt.Disconnect))
	return virt
}

func ProvideLibvirtDialer(log *zap.Logger, config *config.Config) socket.Dialer {
	endpoint, err := url.Parse(config.LibvirtEndpoint)
	if err != nil {
		log.Fatal("unable to parse libvirt endpoint", zap.String("endpoint", config.LibvirtEndpoint))
	}
	var dialer socket.Dialer
	switch endpoint.Scheme {
	case "tcp":
		log.Info("connect to a tcp dialer", zap.String("hostname", endpoint.Hostname()), zap.String("port", endpoint.Port()))
		opts := []dialers.RemoteOption{}
		if endpoint.Port() != "" {
			opts = append(opts, dialers.UsePort(endpoint.Port()))
		}
		dialer = dialers.NewRemote(endpoint.Hostname(), opts...)
	case "unix":
		log.Info("connect to a unix dialer", zap.String("endpoint", endpoint.Path))
		dialer = dialers.NewLocal(dialers.WithSocket(endpoint.Path))
	default:
		log.Fatal("unimplemented protocol for libvirt", zap.String("protocol", endpoint.Scheme))
		return nil
	}
	return dialer
}

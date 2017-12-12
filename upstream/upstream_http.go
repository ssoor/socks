package upstream

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ssoor/fundadore/api"
	"github.com/ssoor/fundadore/log"
)

type URLUpstreamDialer struct {
	url string
	lock    sync.Mutex
	interval time.Duration

	dialer *UpstreamDialer
}

func NewUpstreamDialerByURL(url string, updateInterval  time.Duration) *URLUpstreamDialer {
	settings := Settings{

	}

	dialer := &URLUpstreamDialer{
		url:     url,
		interval: updateInterval * time.Second,
		dialer: NewUpstreamDialer(settings),
	}

	go dialer.backgroundUpdateServices()

	return dialer
}

func getSocksdSetting(url string) (setting Settings, err error) {
	var jsonData string

	if jsonData, err = api.GetURL(url); err != nil {
		return setting, errors.New(fmt.Sprint("Query setting interface failed, err: ", err))
	}

	if err = json.Unmarshal([]byte(jsonData), &setting); err != nil {
		return setting, errors.New("Unmarshal setting interface failed.")
	}

	return setting, nil
}

func (u *URLUpstreamDialer) backgroundUpdateServices() {
	var err error
	var setting Settings

	for {
		log.Info("Setting messenger server information:")
		
		log.Info("\tNext flush time :", u.interval)

		if setting, err = getSocksdSetting(u.url); nil == err {
			log.Info("\tDial timeout time :", setting.DialTimeout)
			log.Info("\tDNS cache timeout time :", setting.DNSCacheTime)
			
			for _, upstream := range setting.Upstreams {
				log.Info("\tUpstream :", upstream.Address)
			}
	
			u.lock.Lock()
			u.dialer = NewUpstreamDialer(setting)
			u.lock.Unlock()
		}

		time.Sleep(u.interval)
	}
}

func (u *URLUpstreamDialer) Dial(network, address string) (conn net.Conn, err error) {
	u.lock.Lock()
	defer u.lock.Unlock()
	
	return u.dialer.Dial(network, address)
}

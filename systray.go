package flutter

import (
	"encoding/json"
	"github.com/getlantern/systray"
	"github.com/go-flutter-desktop/go-flutter/plugin"
	"io/ioutil"
	"log"
	"os"
)

const (
	SYSTRAY_CHANNEL_NAME = "flutter/systray"
)

type MenuItem struct {
	title   string
	tooltip string
}

type SystrayPlugin struct {
	systrayChannel *plugin.BasicMessageChannel
	config         SystrayConfig
	menus          map[MenuItem]func(p *SystrayPlugin, m *messenger) error
}

var systrayPlugin = &SystrayPlugin{}

func (p *SystrayPlugin) InitPlugin(m plugin.BinaryMessenger) error {
	messenger := m.(*messenger)
	p.systrayChannel = plugin.NewBasicMessageChannel(messenger, SYSTRAY_CHANNEL_NAME, &SystrayCodec{})
	p.systrayChannel.HandleFunc(func(message interface{}) (reply interface{}, err error) {
		log.Println(message)
		return message, nil
	})
	onReady := func() {
		file, err := os.Open(p.config.IconPath)
		if err != nil {
			log.Fatal("icon_path of SystrayConfig err", err)
		}
		bs, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal("read bytes from file err", err)
		}
		systray.SetIcon(bs)
		systray.SetTitle(p.config.Title)
		systray.SetTooltip(p.config.Tooltip)

		for i, f := range p.menus {
			menuItem := systray.AddMenuItem(i.title, i.tooltip)
			go func() {
				for {
					select {
					case <-menuItem.ClickedCh:
						log.Println(i.title, "systray click")
						err := f(p, messenger)
						if err != nil {
							log.Println(i.title, "systray err")
						}
					}
				}
			}()
		}

	}
	onExit := func() {
		log.Println("app exit")
	}
	systray.Run(onReady, onExit)
	return nil
}

type SystrayMessage struct {
	MessageType string `json:"message_type"`
}

type SystrayCodec struct {
}

func (s *SystrayCodec) EncodeMessage(message interface{}) (binaryMessage []byte, err error) {
	return json.Marshal(message)
}
func (s *SystrayCodec) DecodeMessage(binaryMessage []byte) (message interface{}, err error) {
	var sm SystrayMessage
	err = json.Unmarshal(binaryMessage, &sm)
	return sm, err
}

type SystrayConfig struct {
	IconPath string `json:"icon_path"`
	Title    string `json:"title"`
	Tooltip  string `json:"tooltip"`
}

func SystrayOption(sc SystrayConfig, menus map[MenuItem]func(p *SystrayPlugin, m *messenger) error) Option {
	return func(c *config) {
		for _, plugin := range c.plugins {
			p, ok := plugin.(*SystrayPlugin)
			if ok {
				p.config = sc
				p.menus = menus
				break
			}
		}
	}
}

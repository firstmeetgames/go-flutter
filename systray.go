package flutter

import (
	"encoding/json"
	"github.com/getlantern/systray"
	"github.com/go-flutter-desktop/go-flutter/embedder"
	"github.com/go-flutter-desktop/go-flutter/plugin"
	"io/ioutil"
	"log"
	"os"
)

const (
	SYSTRAY_CHANNEL_NAME = "flutter/systray"
)

type MenuItem struct {
	Title   string
	Tooltip string
}

type SystrayPlugin struct {
	systrayChannel *plugin.BasicMessageChannel
	Engine         *embedder.FlutterEngine
	Config         SystrayConfig
}

func (p *SystrayPlugin) InitPlugin(m plugin.BinaryMessenger) error {
	messenger := m.(*messenger)
	engine := messenger.engine
	p.Engine = engine
	p.systrayChannel = plugin.NewBasicMessageChannel(messenger, SYSTRAY_CHANNEL_NAME, &SystrayCodec{})
	p.systrayChannel.HandleFunc(func(message interface{}) (reply interface{}, err error) {
		log.Println(message)
		return message, nil
	})
	handle := func(item *systray.MenuItem, menu Menu) {
		for {
			<-item.ClickedCh
			log.Println(menu.MenuItem.Title, "systray click")
			err := menu.MenuHandler(p, m)
			log.Println(err)
			if err != nil {
				log.Println(menu.MenuItem.Title, "systray err")
				os.Exit(0)
				return
			}
		}
	}
	onReady := func() {
		file, err := os.Open(p.Config.IconPath)
		if err != nil {
			log.Fatal("icon_path of SystrayConfig err", err)
		}
		bs, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal("read bytes from file err", err)
		}
		systray.SetIcon(bs)
		systray.SetTitle(p.Config.Title)
		systray.SetTooltip(p.Config.Tooltip)

		for _, menu := range p.Config.Menus {
			menuItem := systray.AddMenuItem(menu.MenuItem.Title, menu.MenuItem.Tooltip)
			go handle(menuItem, menu)
		}
	}
	onExit := func() {
		log.Println("app exit")
	}
	go func() {
		systray.Run(onReady, onExit)
	}()
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

type MenuHandler func(p *SystrayPlugin, m plugin.BinaryMessenger) error

type Menu struct {
	MenuItem    MenuItem    `json:"menu_item"`
	MenuHandler MenuHandler `json:"menu_handler"`
}
type SystrayConfig struct {
	IconPath string `json:"icon_path"`
	Title    string `json:"title"`
	Tooltip  string `json:"tooltip"`
	Menus    []Menu `json:"menus"`
}

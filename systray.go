package flutter

import (
	"github.com/getlantern/systray"
	"github.com/go-flutter-desktop/go-flutter/embedder"
	"github.com/go-flutter-desktop/go-flutter/plugin"
	"io/ioutil"
	"log"
	"os"
)

//const (
//	SYSTRAY_CHANNEL_NAME = "flutter/systray"
//)

type MenuItem struct {
	Title   string
	Tooltip string
}

type SystrayPlugin struct {
	//systrayChannel *plugin.BasicMessageChannel
	Config SystrayConfig
}

func (p *SystrayPlugin) InitPlugin(m plugin.BinaryMessenger) error {
	messenger := m.(*messenger)
	engine := messenger.engine
	//p.systrayChannel = plugin.NewBasicMessageChannel(messenger, SYSTRAY_CHANNEL_NAME, &SystrayCodec{})
	//p.systrayChannel.HandleFunc(func(message interface{}) (reply interface{}, err error) {
	//	log.Println(message)
	//	return message, nil
	//})
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

		for i, f := range p.Config.Menus {
			menuItem := systray.AddMenuItem(i.Title, i.Tooltip)
			go func() {
				for {
					select {
					case <-menuItem.ClickedCh:
						log.Println(i.Title, "systray click")
						err := f(p, messenger, engine)
						if err != nil {
							log.Println(i.Title, "systray err")
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

/*type SystrayMessage struct {
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
}*/

type SystrayConfig struct {
	IconPath string `json:"icon_path"`
	Title    string `json:"title"`
	Tooltip  string `json:"tooltip"`
	Menus    map[MenuItem]func(p *SystrayPlugin, m plugin.BinaryMessenger, engine *embedder.FlutterEngine) error
}

package save

import (
	"io"
	"reflect"
	"strings"

	"charm.land/bubbles/v2/key"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

const (
	keyMapFileName = "keymap.yaml"
)

var (
	_ yaml.Marshaler   = (*KeyMap)(nil)
	_ yaml.Unmarshaler = (*KeyMap)(nil)
)

type KeyMap struct {
	// General
	Up      key.Binding `yaml:"up"`
	Down    key.Binding `yaml:"down"`
	Escape  key.Binding `yaml:"escape"`
	Confirm key.Binding `yaml:"confirm"`
	Help    key.Binding `yaml:"help"`

	// App Binds
	Quit       key.Binding `yaml:"quit"`
	Create     key.Binding `yaml:"create"`
	Remove     key.Binding `yaml:"remove"`
	CloseTab   key.Binding `yaml:"close_tab"`
	DumpScreen key.Binding `yaml:"dump_screen"`

	// Tab Binds
	Next     key.Binding `yaml:"next"`
	Previous key.Binding `yaml:"previous"`
	TabJump  key.Binding `yaml:"tab_jump"`

	// Quick Join
	QuickJoin key.Binding `yaml:"quick_join"`

	// Chat Binds
	InsertMode   key.Binding `yaml:"insert_mode"`
	InspectMode  key.Binding `yaml:"inspect_mode"`
	ChatPopUp    key.Binding `yaml:"chat_pop_up"`
	ChannelPopUp key.Binding `yaml:"channel_pop_up"`
	DumpChat     key.Binding `yaml:"dump_chat"`
	QuickTimeout key.Binding `yaml:"quick_timeout"`
	CopyMessage  key.Binding `yaml:"copy_message"`
	SearchMode   key.Binding `yaml:"search_mode"`

	// Account Binds
	MarkLeader key.Binding `yaml:"mark_leader"`
}

func (c *KeyMap) MarshalYAML() (interface{}, error) {
	data := map[string][]string{}

	for i := 0; i < reflect.ValueOf(c).Elem().NumField(); i++ {
		field := reflect.TypeOf(c).Elem().Field(i)
		value := reflect.ValueOf(c).Elem().Field(i)

		if value.IsZero() {
			continue
		}

		fieldName := field.Tag.Get("yaml")
		if fieldName == "" {
			fieldName = field.Name
		}

		data[fieldName] = value.Interface().(key.Binding).Keys()
	}

	return data, nil
}

func (c *KeyMap) UnmarshalYAML(value *yaml.Node) error {
	target := map[string][]string{}
	if err := value.Decode(&target); err != nil {
		return err
	}

	val := reflect.ValueOf(c).Elem()

	for targetField, binds := range target {
		for i := 0; i < val.NumField(); i++ {
			fieldName := val.Type().Field(i).Tag.Get("yaml")
			if fieldName == "" {
				fieldName = val.Type().Field(i).Name
			}

			if fieldName == targetField {
				keyBind := reflect.ValueOf(c).Elem().Field(i).Interface().(key.Binding)
				keyBind.SetKeys(binds...)
				keyBind.SetHelp(strings.Join(binds, "/"), keyBind.Help().Desc) // overwrite help with old description but new keys
				reflect.ValueOf(c).Elem().Field(i).Set(reflect.ValueOf(keyBind))
			}
		}
	}

	return nil
}

func BuildDefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "escape"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Help: key.NewBinding(
			key.WithHelp("", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Create: key.NewBinding(
			key.WithKeys("ctrl+t"),
			key.WithHelp("ctrl+t", "open new tab/add account"),
		),
		Remove: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "remove"),
		),
		CloseTab: key.NewBinding(
			key.WithKeys("ctrl+q"),
			key.WithHelp("ctrl+q", "close current tab"),
		),
		DumpScreen: key.NewBinding(
			key.WithHelp("", "dump screen"),
		),
		QuickJoin: key.NewBinding(
			key.WithHelp("", "quick join channel"),
		),
		Next: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next item"),
		),
		Previous: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "previous item"),
		),
		TabJump: key.NewBinding(
			key.WithKeys("alt+1", "alt+2", "alt+3", "alt+4", "alt+5", "alt+6", "alt+7", "alt+8", "alt+9"),
			key.WithHelp("alt+1..9", "jump to tab N"),
		),
		InsertMode: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "insert mode"),
		),
		InspectMode: key.NewBinding(
			key.WithHelp("", "user inspect mode"),
		),
		ChatPopUp: key.NewBinding(
			key.WithHelp("", "twitch chat browser pop up"),
		),
		ChannelPopUp: key.NewBinding(
			key.WithHelp("", "twitch channel pop up"),
		),
		MarkLeader: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "mark account as main account"),
		),
		QuickTimeout: key.NewBinding(
			key.WithHelp("", "quick timeout"),
		),
		DumpChat: key.NewBinding(
			key.WithHelp("", "dump chat"),
		),
		CopyMessage: key.NewBinding(
			key.WithHelp("", "copy selected message"),
		),
		SearchMode: key.NewBinding(
			key.WithHelp("", "start search mode in chat window"),
		),
	}
}

func CreateReadKeyMap() (KeyMap, error) {
	f, err := openCreateConfigFile(afero.NewOsFs(), keyMapFileName)
	if err != nil {
		return KeyMap{}, err
	}

	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return KeyMap{}, err
	}

	// Config was empty, return default config and write a default one to disk
	if stat.Size() == 0 {
		m := BuildDefaultKeyMap()
		b, err := yaml.Marshal(&m)
		if err != nil {
			return KeyMap{}, err
		}

		if _, err := f.Write(b); err != nil {
			return KeyMap{}, err
		}

		return m, nil
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return KeyMap{}, err
	}

	// Config was not empty, read it and return it
	readableMap := BuildDefaultKeyMap()
	if err := yaml.Unmarshal(b, &readableMap); err != nil {
		return KeyMap{}, err
	}

	return readableMap, nil
}

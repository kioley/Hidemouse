package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"golang.design/x/hotkey"
)

type Settings []int

var settings Settings = []int{1, 0, 90, 4}

func main() {
	settings.setFromFile("settings.ini")
	freeze := ""
	if settings[1] == 1 {
		freeze = "f"
	}
	quit := hot(&freeze, settings[2], settings[3:])

	a := app.New()
	w := a.NewWindow("Hidemouse")

	w.Resize(fyne.NewSize(400, 200))

	if desk, ok := a.(desktop.App); ok {
		m := fyne.NewMenu("Hidemouse",
			fyne.NewMenuItem("Settings", func() {
				w.Show()
			}))
		desk.SetSystemTrayMenu(m)
	}

	hideSettingsCheck := widget.NewCheck("Show settings at startup", func(hide bool) {
		if hide {
			settings[0] = 1
		} else {
			settings[0] = 0
		}
		settings.writeToFile("settings.ini")
	})
	if settings[0] == 1 {
		hideSettingsCheck.SetChecked(true)
	}

	freezeCursorCheck := widget.NewCheck("Freeze hidden cursor", func(hide bool) {
		if hide {
			settings[1] = 1
			freeze = "f"
		} else {
			settings[1] = 0
			freeze = ""
		}
		settings.writeToFile("settings.ini")
	})
	if freeze == "f" {
		freezeCursorCheck.SetChecked(true)
	}

	modsCheck := widget.NewCheckGroup([]string{
		"Shift",
		"Ctrl",
		"Alt",
	}, func(checks []string) {
		settings = settings[0:3]
		for _, v := range checks {
			switch v {
			case "Shift":
				settings = append(settings, 4)
			case "Ctrl":
				settings = append(settings, 2)
			case "Alt":
				settings = append(settings, 1)
			}
		}
		quit <- true
		<-quit
		quit = hot(&freeze, settings[2], settings[3:])
		settings.writeToFile("settings.ini")
	})
	var mods []string
	for _, v := range settings[2:] {
		switch v {
		case 4:
			mods = append(mods, "Shift")
		case 2:
			mods = append(mods, "Ctrl")
		case 1:
			mods = append(mods, "Alt")
		}
	}
	modsCheck.SetSelected(mods)

	selector := widget.NewSelect([]string{
		"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L",
		"M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X",
		"Y", "Z", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0",
	}, func(key string) {
		settings[2] = int([]byte(key)[0])
		quit <- true
		<-quit
		quit = hot(&freeze, settings[2], settings[3:])
		settings.writeToFile("settings.ini")
	})
	selector.PlaceHolder = string(rune(settings[2]))

	hotkeysCardCont := container.NewVBox(
		modsCheck,
		selector,
	)
	hotkeysCard := widget.NewCard("", "Keyboard shortcut:", hotkeysCardCont)

	mainCardCont := container.NewVBox(
		hideSettingsCheck,
		freezeCursorCheck,
		hotkeysCard,
	)
	mainCard := widget.NewCard("Settings", "", mainCardCont)

	quitButton := widget.NewButton("     Quit the program     ", func() {
		a.Quit()
	})
	// quitButton.Resize(fyne.NewSize(400, 20))

	trayButton := widget.NewButton("      Collapse to tray      ", func() {
		w.Hide()
	})

	buttonsContainer := container.NewHBox(trayButton, layout.NewSpacer(), quitButton)

	cont := container.NewVBox(
		mainCard,
		buttonsContainer,
	)
	w.SetContent(cont)

	w.SetCloseIntercept(func() {
		w.Hide()
	})

	if settings[0] == 0 {
		a.Run()
	} else {
		w.ShowAndRun()
	}

}

func hot(freez *string, key int, mods []int) chan bool {
	quit := make(chan bool)

	var m []hotkey.Modifier
	for _, v := range mods {
		m = append(m, hotkey.Modifier(v))
	}
	hk := hotkey.New(m, hotkey.Key(key))

	if err := hk.Register(); err != nil {
		fmt.Println(err)
	}

	hotkeyChan := hk.Keydown()

	go func() {
		for {
			select {
			case <-hotkeyChan:
				path, err := exec.LookPath(`.\nomousy.exe`)

				if err != nil {
					fmt.Println(err)
					return
				}

				cmd := exec.Command(path, *freez, "h")

				cmd.Start()
			case <-quit:
				if err := hk.Unregister(); err != nil {
					fmt.Println(err)
				}
				quit <- true
				return
			}
		}
	}()
	return quit
}

func (o *Settings) setFromFile(fileName string) {
	dataFromFile, err := os.ReadFile(fileName)

	if err != nil {
		fmt.Println(err)
		return
	}
	options := strings.Split(string(dataFromFile), " ")

	*o = nil
	for _, v := range options {
		digit, err := strconv.Atoi(v)
		if err != nil {
			fmt.Println(err)
		}
		*o = append(*o, digit)
	}

}

func (s Settings) writeToFile(name string) {
	var str string
	for k, v := range settings {
		str += strconv.Itoa(v)
		if k != len(settings)-1 {
			str += " "
		}
	}
	os.WriteFile(name, []byte(str), 0600)
}

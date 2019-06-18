// +build windows

package main

//捕获热键，参考：https://stackoverflow.com/questions/38646794/implement-a-global-hotkey-in-golang

import (
	"bytes"
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

const (
	ModAlt = 1 << iota
	ModCtrl
	ModShift
	ModWin
)

type Hotkey struct {
	Id        int // Unique id
	Modifiers int // Mask of modifiers
	KeyCode   int // Key code, e.g. 'A'
}

type MSG struct {
	HWND   uintptr
	UINT   uintptr
	WPARAM int16
	LPARAM int64
	DWORD  int32
	POINT  struct{ X, Y int64 }
}

// String returns a human-friendly display name of the hotkey
// such as "Hotkey[Id: 1, Alt+Ctrl+O]"
func (h *Hotkey) String() string {
	mod := &bytes.Buffer{}
	if h.Modifiers&ModAlt != 0 {
		mod.WriteString("Alt+")
	}
	if h.Modifiers&ModCtrl != 0 {
		mod.WriteString("Ctrl+")
	}
	if h.Modifiers&ModShift != 0 {
		mod.WriteString("Shift+")
	}
	if h.Modifiers&ModWin != 0 {
		mod.WriteString("Win+")
	}
	return fmt.Sprintf("Hotkey[Id: %d, %s%c]", h.Id, mod, h.KeyCode)
}

func HotkeyHandel() {
	user32 := syscall.MustLoadDLL("user32")
	reghotkey := user32.MustFindProc("RegisterHotKey")
	keys := map[int16]*Hotkey{
		1: &Hotkey{1, ModShift + ModCtrl, 'H'}, // Shift+CTRL+H 隐藏窗口
		2: &Hotkey{2, ModShift + ModCtrl, 'S'}, // Shift+CTRL+S 显示窗口
	}
	// Register hotkeys:
	for _, v := range keys {
		r1, _, err := reghotkey.Call(
			0, uintptr(v.Id), uintptr(v.Modifiers), uintptr(v.KeyCode))
		if r1 == 1 {
			//fmt.Println("Registered", v)
		} else {
			fmt.Println("热键注册失败：", v, ", error:", err)
		}
	}
	peekmsg := user32.MustFindProc("PeekMessageW")

	for {
		var msg = &MSG{}
		peekmsg.Call(uintptr(unsafe.Pointer(msg)), 0, 0, 0, 1)

		// Registered id is in the WPARAM field:
		if id := msg.WPARAM; id != 0 {
			//fmt.Println("Hotkey pressed:", keys[id])
			if id == 1 { // Shift+CTRL+H 隐藏窗口
				HideConsole()
			} else if id == 2 { // Shift+CTRL+S 显示窗口
				ShowConsole()
			}
		}

		time.Sleep(time.Millisecond * 50)
	}
}

// +build windows

package main

//显示和隐藏控制台，参考：https://stackoverflow.com/questions/23250505/how-do-i-create-an-executable-from-golang-that-doesnt-open-a-command-cmd-wind

import (
	"github.com/gonutz/ide/w32"
)

//隐藏console
func HideConsole() {
	ShowConsoleAsync(w32.SW_HIDE)
}

//显示console
func ShowConsole() {
	ShowConsoleAsync(w32.SW_SHOW)
}

func ShowConsoleAsync(commandShow uintptr) {
	console := w32.GetConsoleWindow()
	if console != 0 {
		_, consoleProcID := w32.GetWindowThreadProcessId(console)
		if w32.GetCurrentProcessId() == consoleProcID {
			w32.ShowWindowAsync(console, commandShow)
		}
	}
}

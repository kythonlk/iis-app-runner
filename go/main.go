package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/rivo/tview"
)

type App struct {
	mu        sync.RWMutex
	running   bool
	logBuffer []string
}

func (a *App) Start(ctx interface{}, logPath string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.running = true
	a.logBuffer = append(a.logBuffer, "App started")
	return nil
}

func (a *App) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.running = false
	a.logBuffer = append(a.logBuffer, "App stopped")
	return nil
}

type AppManager struct {
	apps    map[string]*App
	logPath string
}

type UI struct {
	app     *tview.Application
	list    *tview.List
	logView *tview.TextView
	status  *tview.TextView
	manager *AppManager
}

func StartUI(manager *AppManager) {
	ui := &UI{
		app:     tview.NewApplication(),
		list:    tview.NewList(),
		logView: tview.NewTextView(),
		status:  tview.NewTextView(),
		manager: manager,
	}

	ui.list.SetBorder(true).SetTitle("Applications").SetTitleAlign(tview.AlignLeft)
	ui.logView.SetBorder(true).SetTitle("Logs").SetTitleAlign(tview.AlignLeft)
	ui.status.SetBorder(true).SetTitle("Status").SetTitleAlign(tview.AlignLeft)

	for name := range manager.apps {
		ui.list.AddItem(name, "Select to view logs", 0, nil)
	}

	ui.list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		ui.handleAppSelection(index, mainText, secondaryText, shortcut)
	})

	flex := tview.NewFlex().
		AddItem(ui.list, 0, 1, true).
		AddItem(ui.logView, 0, 2, false).
		AddItem(ui.status, 0, 1, false)

	ui.app.SetRoot(flex, true)

	if err := ui.app.Run(); err != nil {
		log.Fatalf("Error running UI: %v", err)
	}
}

func (ui *UI) handleAppSelection(index int, mainText, secondaryText string, shortcut rune) {
	if app, exists := ui.manager.apps[mainText]; exists {
		app.mu.RLock()
		status := "Stopped"
		if app.running {
			status = "Running"
		}
		logs := make([]string, len(app.logBuffer))
		copy(logs, app.logBuffer)
		app.mu.RUnlock()

		ui.app.QueueUpdateDraw(func() {
			ui.logView.Clear()
			for _, line := range logs {
				fmt.Fprintln(ui.logView, line)
			}
			ui.status.SetText(fmt.Sprintf("App: %s | Status: %s", mainText, status))
		})
	}
}

func (ui *UI) restartApp(name string) {
	if app, exists := ui.manager.apps[name]; exists {
		ui.app.QueueUpdateDraw(func() {
			ui.status.SetText(fmt.Sprintf("Restarting app: %s...", name))
		})

		app.mu.Lock()
		if app.running {
			if err := app.Stop(); err != nil {
				ui.app.QueueUpdateDraw(func() {
					ui.status.SetText(fmt.Sprintf("Failed to stop app: %s | Error: %v", name, err))
				})
				app.mu.Unlock()
				return
			}
		}
		app.mu.Unlock()

		if err := app.Start(nil, ui.manager.logPath); err != nil {
			ui.app.QueueUpdateDraw(func() {
				ui.status.SetText(fmt.Sprintf("Failed to restart app: %s | Error: %v", name, err))
			})
			return
		}

		ui.app.QueueUpdateDraw(func() {
			ui.status.SetText(fmt.Sprintf("Successfully restarted app: %s", name))
		})
	}
}

func main() {
	manager := &AppManager{
		apps: map[string]*App{
			"App1": {logBuffer: []string{"App1 initialized"}},
			"App2": {logBuffer: []string{"App2 initialized"}},
		},
		logPath: "/var/log/apps",
	}

	StartUI(manager)
}

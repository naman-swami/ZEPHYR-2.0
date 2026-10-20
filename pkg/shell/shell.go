package shell

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

// ANSI Colors
const (
	ColorCyan   = "\033[36m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorRed    = "\033[31m"
	ColorBold   = "\033[1m"
	ColorReset  = "\033[0m"
)

// RunInteractiveREPL launches an interactive workspace console session.
func RunInteractiveREPL(workspaceRoot string, exePath string) {
	fmt.Printf("\n%s================================================================================%s\n", ColorCyan, ColorReset)
	fmt.Printf("  %s⚡ ZEPHYR 2.0 INTERACTIVE WORKSPACE TERMINAL%s  [PID: %d]\n", ColorCyan+ColorBold, ColorReset, os.Getpid())
	fmt.Printf("  Workspace : %s%s%s\n", ColorBold, workspaceRoot, ColorReset)
	fmt.Printf("  Status    : Active session (Multi-instance isolated, local .taskcache)\n")
	fmt.Printf("%s================================================================================%s\n", ColorCyan, ColorReset)
	fmt.Println("Available commands:")
	fmt.Println("  run [target]       Execute build targets (or default pipeline)")
	fmt.Println("  doctor             Audit workspace health, DAG validity, and system integrity")
	fmt.Println("  organize <subcmd>  Workspace layout & AI token optimizer (analyze, plan, apply, undo)")
	fmt.Println("  vault <subcmd>     AI credential isolation shield (audit, cloak, merge)")
	fmt.Println("  heal <file>        Self-healing AST code repair")
	fmt.Println("  graph              Render visual dependency tree")
	fmt.Println("  clean, prune       Clear local cache")
	fmt.Println("  list               List configured tasks")
	fmt.Println("  help               Display full command reference")
	fmt.Println("  cls, clear         Clear screen")
	fmt.Println("  exit, quit         Close this terminal instance")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("%szephyr [%s]> %s", ColorCyan, filepath.Base(workspaceRoot), ColorReset)
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmdName := parts[0]
		cmdArgs := parts[1:]

		switch cmdName {
		case "exit", "quit", "q":
			fmt.Printf("%sClosing ZEPHYR terminal session.%s\n", ColorYellow, ColorReset)
			return

		case "cls", "clear":
			if runtime.GOOS == "windows" {
				c := exec.Command("cmd", "/c", "cls")
				c.Stdout = os.Stdout
				_ = c.Run()
			} else {
				fmt.Print("\033[H\033[2J")
			}
			continue

		default:
			// Execute through zephyr binary in current directory
			cmd := exec.Command(exePath, append([]string{cmdName}, cmdArgs...)...)
			cmd.Dir = workspaceRoot
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			_ = cmd.Run()
			fmt.Println()
		}
	}
}

// RegisterWindowsExplorer integrates ZEPHYR into Windows File Explorer address bar and context menu.
func RegisterWindowsExplorer(exePath string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("explorer registration is only supported on Windows")
	}

	absExe, err := filepath.Abs(exePath)
	if err != nil {
		return err
	}
	exeDir := filepath.Dir(absExe)

	fmt.Printf("%s[SHELL REGISTRATION]%s Registering ZEPHYR in Windows Explorer...\n", ColorCyan, ColorReset)

	// 1. Register in HKCU App Paths so typing 'zephyr' in File Explorer address bar launches zephyr.exe
	appPathsKey := `HKCU\Software\Microsoft\Windows\CurrentVersion\App Paths\zephyr.exe`
	if out, err := exec.Command("reg", "add", appPathsKey, "/ve", "/t", "REG_SZ", "/d", absExe, "/f").CombinedOutput(); err != nil {
		return fmt.Errorf("failed to register App Paths: %s (%w)", string(out), err)
	}
	_ = exec.Command("reg", "add", appPathsKey, "/v", "Path", "/t", "REG_SZ", "/d", exeDir, "/f").Run()

	// 2. Also register 'zephyr-run.exe' alias pointing to zephyr.exe
	appPathsRunKey := `HKCU\Software\Microsoft\Windows\CurrentVersion\App Paths\zephyr-run.exe`
	_ = exec.Command("reg", "add", appPathsRunKey, "/ve", "/t", "REG_SZ", "/d", absExe, "/f").Run()

	// 3. Create a lightweight zephyr.cmd launcher in user Local AppData WindowsApps or User bin
	userProfile := os.Getenv("USERPROFILE")
	windowsAppsDir := filepath.Join(userProfile, "AppData", "Local", "Microsoft", "WindowsApps")
	if _, err := os.Stat(windowsAppsDir); err == nil {
		shimContent := fmt.Sprintf("@echo off\r\n\"%s\" %%*\r\n", absExe)
		_ = os.WriteFile(filepath.Join(windowsAppsDir, "zephyr.cmd"), []byte(shimContent), 0755)
	}

	// 4. Add Windows Explorer Background Right-Click Context Menu ("Open ZEPHYR Terminal Here")
	menuKey := `HKCU\Software\Classes\Directory\Background\shell\Zephyr`
	_ = exec.Command("reg", "add", menuKey, "/ve", "/t", "REG_SZ", "/d", "Open ZEPHYR Terminal Here", "/f").Run()
	_ = exec.Command("reg", "add", menuKey, "/v", "Icon", "/t", "REG_SZ", "/d", absExe, "/f").Run()
	cmdKey := `HKCU\Software\Classes\Directory\Background\shell\Zephyr\command`
	_ = exec.Command("reg", "add", cmdKey, "/ve", "/t", "REG_SZ", "/d", fmt.Sprintf(`"%s" shell`, absExe), "/f").Run()

	fmt.Printf("%s[SUCCESS]%s ZEPHYR is now registered with Windows File Explorer!\n", ColorGreen, ColorReset)
	fmt.Println("  ✓ File Explorer Address Bar : Type 'zephyr' in any folder address bar to launch terminal.")
	fmt.Println("  ✓ Right-Click Context Menu  : Right-click empty space in any folder -> 'Open ZEPHYR Terminal Here'.")
	fmt.Println("  ✓ Multi-Instance Ready      : Open as many separate terminal instances as you want simultaneously.")

	return nil
}

var (
	kernel32DLL               = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleProcessList = kernel32DLL.NewProc("GetConsoleProcessList")
)

// IsSoleConsoleProcess checks if ZEPHYR is the only process attached to this console buffer.
// If true, it was launched from File Explorer or the Run dialog, not from an existing shell.
func IsSoleConsoleProcess() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	pids := make([]uint32, 2)
	r, _, _ := procGetConsoleProcessList.Call(
		uintptr(unsafe.Pointer(&pids[0])),
		uintptr(len(pids)),
	)
	return r == 1
}

// GuardConsoleExit pauses the console before closing if launched directly from File Explorer.
func GuardConsoleExit(exitCode int) {
	if runtime.GOOS == "windows" && IsSoleConsoleProcess() {
		fmt.Println("\n================================================================================")
		if exitCode == 0 {
			fmt.Printf("%s[ZEPHYR]%s Execution completed successfully.\n", ColorGreen, ColorReset)
		} else {
			fmt.Printf("%s[ZEPHYR]%s Execution finished with exit code %d.\n", ColorRed, ColorReset, exitCode)
		}
		fmt.Printf("Press Enter to close this window, or type a command to continue: ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			input := strings.TrimSpace(scanner.Text())
			if input != "" {
				exePath, _ := os.Executable()
				wd, _ := os.Getwd()
				RunInteractiveREPL(wd, exePath)
			}
		}
	}
}


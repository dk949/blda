package main

import "encoding/json"
import "fmt"
import "io/fs"
import "io/ioutil"
import "os"
import "os/exec"
import "path/filepath"
import "path"
import "strings"
import "syscall"

const CFG_FILE = ".bldr"
const VERSION = "1.4.1"

type ActionType int

const (
    Normal ActionType = iota
    OnlyList
    DryRun
)

func progName() string {
    return path.Base(os.Args[0])
}

func fileInfoMap(vs []fs.FileInfo, f func(fs.FileInfo) string) []string {
    vsm := make([]string, len(vs))
    for i, v := range vs {
        vsm[i] = f(v)
    }
    return vsm
}

func stringMap(vs []string, f func(string) string) []string {
    vsm := make([]string, len(vs))
    for i, v := range vs {
        vsm[i] = f(v)
    }
    return vsm
}

func contains(list []string, a string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

func exitWithError(msg string) {
    fmt.Fprintf(os.Stderr, "ERROR: %s\n", msg)
    os.Exit(1)

}

func usage(msg string) {
    if msg != "" {
        fmt.Fprintf(os.Stderr, "%s\n\n", msg)
    }
    fmt.Fprintf(os.Stderr, "Usage: %s [-h] [-v] [--ls] [--dry-run] [ACTION]\n", progName())
    os.Exit(1)
}

func help() {
    fmt.Printf("Usage: %s [-h] [-v] [--ls] [--dry-run] [ACTION]\n\n", progName())

    fmt.Printf("Run commands stored in the .%s file at the root of teh project\n\n", progName())
    fmt.Println("positional arguments:")
    fmt.Print("  ACTION                target to run\n\n")

    fmt.Println("options:")
    fmt.Print("  -h, --help            show this message and exit\n")
    fmt.Print("  -l, --ls              list available actions\n")
    fmt.Print("  -r, --dry-run         print command which would run for a given action.\n\n")
    fmt.Print("  -v, --version         print program veriosn and exit\n")
    os.Exit(0)
}

func version() {
    fmt.Printf("%s, verison %s\n", progName(), VERSION)
    os.Exit(0)
}

func getActions() (ActionType, []string) {
    if len(os.Args) < 2 {
        usage("")
    }

    var actions []string
    dry := false

    for _, arg := range os.Args[1:] {
        if arg[0] == '-' {
            switch arg {
            case "-h", "--help":
                help()
            case "-v", "--version":
                version()
            case "-l", "--ls":
                return OnlyList, []string{}
            case "-r", "--dry-run":
                dry = true
                continue
            default:
                usage(fmt.Sprintf("Unknonw flag %s", arg))
            }
        }
        actions = append(actions, arg)
    }
    if len(actions) == 0 {
        usage("At least one action has to be specified")
    }

    if dry {
        return DryRun, actions
    }
    return Normal, actions
}

func getRoot() string {
    path, err := os.Getwd()
    if err != nil {
        exitWithError(fmt.Sprint(err))
    }
    ppath := filepath.Dir(path)

    for path != ppath {
        files, err := ioutil.ReadDir(path)
        if err != nil {
            exitWithError("Encounterederror reading directoy")
        }
        if contains(fileInfoMap(files, func(s fs.FileInfo) string { return s.Name() }), CFG_FILE) {
            return path
        }
        path = ppath
        ppath = filepath.Dir(path)
    }
    exitWithError(fmt.Sprintf("Could not find %s in any parent directory", CFG_FILE))
    return ""
}

func readConfig(root string) map[string]interface{} {
    jsonFile, err := os.Open(filepath.Join(root, CFG_FILE))
    if err != nil {
        exitWithError(fmt.Sprintf("Coudl not open file %s:\n\t%s", CFG_FILE, err))
    }
    defer jsonFile.Close()

    byteValue, err := ioutil.ReadAll(jsonFile)
    if err != nil {
        exitWithError(fmt.Sprintf("Failed to read bytes of %s:\n\t%s", CFG_FILE, err))
    }

    var result map[string]interface{}
    err = json.Unmarshal([]byte(byteValue), &result)

    if err != nil {
        exitWithError(fmt.Sprintf("%s appeas to not be valid JSON:\n\t%s", CFG_FILE, err))
    }

    return result
}

func dryRun(action string, config map[string]interface{}) {
    if val, ok := config[action]; ok {
        fields := strings.Fields(val.(string))
        if fields[0] == progName() {
            for _, a := range fields[1:] {
                dryRun(a, config)
            }
        } else {
            fmt.Println(val)
        }
    } else {
        exitWithError(fmt.Sprintf("%s could not be found in %s", action, CFG_FILE))
    }
}

func runAction(action string, root string, config map[string]interface{}) int {
    if val, ok := config[action]; ok {
        args := stringMap(strings.Fields(val.(string)), func(arg string) string {
            if arg[0] == '$' {
                return os.Getenv(arg[1:])
            }
            return arg
        })
        cmd := exec.Command(args[0], args[1:]...)
        cmd.Dir = root
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        err := cmd.Run()
        if err == nil {
            return 0
        }
        if exiterr, ok := err.(*exec.ExitError); ok {
            if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
                return status.ExitStatus()
            }
        } else {
            exitWithError(fmt.Sprintf("no such command %s", args[0]))
        }
    } else {
        exitWithError(fmt.Sprintf("%s could not be found in %s", action, CFG_FILE))
    }
    return 0
}

func runActions(actionType ActionType, actions []string, root string, config map[string]interface{}) {
    switch actionType {
    case OnlyList:
        for k, v := range config {
            fmt.Printf("%s : %s\n", k, v)
        }
    case DryRun:
        for _, action := range actions {
            dryRun(action, config)
        }
    case Normal:
        for _, action := range actions {
            if status := runAction(action, root, config); status != 0 {
                os.Exit(status)
            }
        }
    }
    os.Exit(0)
}

func main() {
    actionType, actions := getActions()
    root := getRoot()
    config := readConfig(root)
    runActions(actionType, actions, root, config)
}

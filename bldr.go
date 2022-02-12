package main

import "encoding/json"
import "fmt"
import "io/fs"
import "io/ioutil"
import "os"
import "os/exec"
import "path/filepath"
import "strings"
import "syscall"

const CFG_FILE = ".bldr"

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

func printHelp(msg string) {
    fmt.Fprintf(os.Stderr, "%s\n\n", msg)
    fmt.Fprintf(os.Stderr, "Usage: bldr [ACTION]\n")
    os.Exit(1)
}

func getActions() []string {
    if len(os.Args) < 2 {
        printHelp("")
    }

    var actions []string

    for _, arg := range os.Args[1:] {
        if arg[0] == '-' {
            printHelp("Flags are not currently accepted")
        }
        actions = append(actions, arg)
    }
    if len(actions) == 0 {
        printHelp("At least one action has to be specified")
    }

    return actions
}

func getRoot() string {
    path, err := os.Getwd()
    if err != nil {
        fmt.Println(err)
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
        if exiterr, ok := err.(*exec.ExitError); ok {
            if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
                return status.ExitStatus()
            }
        } else {
            return 0
        }
    } else {
        exitWithError(fmt.Sprintf("%s could not be found in %s", action, CFG_FILE))
    }
    return 0
}

func runActions(actions []string, root string, config map[string]interface{}) {
    for _, action := range actions {
        if status := runAction(action, root, config); status != 0 {
            os.Exit(status)
        }
    }
    os.Exit(0)
}

func main() {
    actions := getActions()
    root := getRoot()
    config := readConfig(root)
    runActions(actions, root, config)
}

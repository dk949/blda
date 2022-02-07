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

func Map(vs []fs.FileInfo, f func(fs.FileInfo) string) []string {
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

func getAction() string {
    if len(os.Args) < 2 {
        printHelp("")
    }

    var action string

    for _, arg := range os.Args[1:] {
        if arg[0] == '-' {
            printHelp("Flags are not currently accepted")
        }
        if action == "" {
            action = arg
        } else {
            printHelp("Only one action can be used at once")
        }
    }
    if action == "" {
        printHelp("An action has to be specified")
    }

    return action
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
        if contains(Map(files, func(s fs.FileInfo) string { return s.Name() }), CFG_FILE) {
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

func runAction(action string, root string, config map[string]interface{}) {
    if val, ok := config[action]; ok {
        args := strings.Fields(val.(string))
        cmd := exec.Command(args[0], args[1:]...)
        cmd.Dir = root
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        err := cmd.Run()
        if exiterr, ok := err.(*exec.ExitError); ok {
            if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
                os.Exit(status.ExitStatus())
            }
        } else {
            os.Exit(0)
        }
    } else {
        exitWithError(fmt.Sprintf("%s could not be found in %s", action, CFG_FILE))
    }

}

func main() {
    action := getAction()
    root := getRoot()
    config := readConfig(root)
    runAction(action, root, config)
}

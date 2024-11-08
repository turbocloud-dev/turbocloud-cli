package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"sync"
	"text/template"
)

func createTemplate(name, t string) *template.Template {
	return template.Must(template.New(name).Parse(t))
}

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}

func executeScriptString(scriptString string) error {

	scriptContents := []byte(scriptString)

	currentUser, err := user.Current()
	if err != nil {
		fmt.Println("Cannot get home directory:", err)
	}

	homeDir := currentUser.HomeDir

	id, err := NanoId(7)
	if err != nil {
		fmt.Println("Cannot generate new NanoId for Deployment:", err)
		return err
	}
	fileName := homeDir + "/" + id + ".sh"

	err = os.WriteFile(fileName, scriptContents, 0644)
	if err != nil {
		fmt.Printf(" Cannot save script: %s\n", err.Error())
		return err
	}

	cmd := exec.Command("/bin/sh", fileName)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	cmd.Start()

	var wg sync.WaitGroup
	outch := make(chan string, 10)

	scannerStdout := bufio.NewScanner(stdout)
	wg.Add(1)
	go func() {
		for scannerStdout.Scan() {
			text := scannerStdout.Text()
			if strings.TrimSpace(text) != "" {
				outch <- text
			}
		}
		wg.Done()
	}()
	scannerStderr := bufio.NewScanner(stderr)
	wg.Add(1)
	go func() {
		for scannerStderr.Scan() {
			text := scannerStderr.Text()
			if strings.TrimSpace(text) != "" {
				outch <- text
			}
		}
		wg.Done()
	}()

	go func() {
		wg.Wait()
		close(outch)
	}()

	for t := range outch {
		fmt.Println(t)
	}

	wg.Wait()

	err = os.Remove(fileName) //remove the script file
	if err != nil {
		fmt.Printf(" Cannot remove script: %s\n", err.Error())
	}

	return nil
}

func YesNoPrompt(label string, def bool) bool {

	r := bufio.NewReader(os.Stdin)
	var s string

	for {
		fmt.Fprintf(os.Stderr, "%s ", label)
		s, _ = r.ReadString('\n')
		s = strings.TrimSpace(s)
		if s == "" {
			return def
		}
		s = strings.ToLower(s)
		if s == "y" || s == "yes" {
			return true
		}
		if s == "n" || s == "no" {
			return false
		}
	}
}

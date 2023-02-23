package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

var ops int64
var publicKey *ssh.PublicKeys
var keyError error
var sema chan struct{}
var gitch chan string
var buildch chan string
var stopch chan struct{}
var wg sync.WaitGroup

// TO DO: Add a channel for successfull and failed builds respectively
func main() {

	root := os.Args[1:]
	if len(root) == 0 {
		root = []string{"."}
	}

	pre()
	startTime := time.Now()
	fileOperation(root[0])
	startCoordinator()
	wg.Wait()
	post()
	executionTime := time.Since(startTime)
	fmt.Printf("The total execution time is: %s", executionTime.String())

}
func startCoordinator() {
	go func() {
		for {
			select {
			case path := <-gitch:
				fmt.Println(path)
				go pull(path)
			case path := <-buildch:
				fmt.Println(path)
				go build(path)
			case <-stopch:
				os.Exit(1)
			}

		}
	}()
}
func pre() {
	initSSHKey()
	sema = make(chan struct{}, 5)
	gitch = make(chan string)
	buildch = make(chan string)
}
func post() {
	//kills gradle daemons
	out, err := exec.Command("bash", "-c", "pkill -f '.*GradleDaemon.*'").Output()
	fmt.Println(string(out))
	if err != nil {
		fmt.Println(fmt.Errorf("the error is %s", err))
	}

	fmt.Printf("The number of goroutines created for this task is: %d\n", ops)
	//closes channels
	close(sema)
	close(gitch)
	close(buildch)
	// close(stopch)
}

func initSSHKey() {
	sshPath := os.Getenv("HOME") + "/.ssh/id_ed25519"
	fmt.Println(sshPath)
	sshKey, _ := ioutil.ReadFile(sshPath)
	publicKey, keyError = ssh.NewPublicKeys("git", []byte(sshKey), "")
	if keyError != nil {
		fmt.Println(keyError)
		os.Exit(1)
	}
}

func fileOperation(absolutePath string) {
	f, err := os.ReadDir(absolutePath)
	if err != nil {
		fmt.Println("Directory does not exist", err)
		os.Exit(1)
	}
	for _, fd := range f {

		if fd.IsDir() {
			if fd.Name() == ".git" {
				gitch <- absolutePath
				continue
			}
			wg.Add(1)
			atomic.AddInt64(&ops, 1)
			// fmt.Println("The name of the file is " + fd.Name())
			go func(fdp fs.DirEntry) {
				defer func() {
					wg.Done()
					<-sema
				}()
				sema <- struct{}{}
				fullPath, err := filepath.Abs(absolutePath)
				if err != nil {
					err = fmt.Errorf("INCORRECT DIRECTORY ADRESS")
					fmt.Println(err.Error())
					os.Exit(1)
				}
				path := filepath.Join(fullPath, fdp.Name())
				fileOperation(path)
			}(fd)
		} else if !strings.Contains(absolutePath, "generated") && fd.Name() == "gradlew" {
			abs, err := filepath.Abs(absolutePath)
			if err != nil {
				fmt.Println(fmt.Errorf("Error :%s", err))
			}
			buildch <- abs
		}
	}
}

func pull(dir string) {
	defer wg.Done()
	r, _ := git.PlainOpen(dir)

	wg.Add(1)
	w, _ := r.Worktree()
	ch := make(chan error)
	go func() {
		err := w.Pull(&git.PullOptions{RemoteName: "origin", Progress: os.Stdout, Auth: publicKey})
		if err != nil {
			fmt.Println(fmt.Errorf("Error is %s", err))
			ch <- err
		}
	}()

	t := time.Tick(time.Millisecond * 2000)
	select {
	case <-ch:
	case <-t:
	}

}

func build(path string) {
	defer wg.Done()
	wg.Add(1)
	fmt.Println(fmt.Printf("The build process for the project %s begins", path))
	cmnd := filepath.Join(path, "gradlew")
	out, err := exec.Command(cmnd, "-p", path, "build").Output()
	fmt.Println(string(out))
	if err != nil {
		fmt.Println(fmt.Errorf("the error is %s", err))
	}
}
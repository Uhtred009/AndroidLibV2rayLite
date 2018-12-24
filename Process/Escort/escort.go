package Escort

import (
	"os"
	"os/exec"

	"log"
)
import "github.com/2dust/AndroidLibV2rayLite/CoreI"

func (v *Escorting) EscortRun(proc string, pt []string, forgiveable bool, tapfd int, additionalEnv string) {
	log.Println(proc)
	log.Println(pt)
	count := 42
	for count > 0 {
		cmd := exec.Command(proc, pt...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if len(additionalEnv) > 0 {
			//additionalEnv := "FOO=bar"
			newEnv := append(os.Environ(), additionalEnv)
			cmd.Env = newEnv
		}
		
		if tapfd != 0 {
			file := os.NewFile(uintptr(tapfd), "/dev/tap0")
			var files []*os.File
			cmd.ExtraFiles = append(files, file)
		}

		err := cmd.Start()
		
		if err != nil {
			log.Println(err)
		}
		*v.escortProcess = append(*v.escortProcess, cmd.Process)
		log.Println("EscortRun Waiting....")
		err = cmd.Wait()
		log.Println("EscortRun Exit")
		log.Println(err)
		if v.status.IsRunning {
			log.Println("Unexpected Exit")
			count--
		} else {
			return
		}
	}

	if v.status.IsRunning && !forgiveable {
		v.unforgivnesschan <- 0
	}

}

func (v *Escorting) escortBeg(proc string, pt []string, forgiveable bool) {
	go v.EscortRun(proc, pt, forgiveable, 0)
}

func (v *Escorting) unforgivenessCloser() {
	log.Println("unforgivenessCloser() <-v.unforgivnesschan")
	<-v.unforgivnesschan
	/*if v.status.IsRunning {
		//TODO:v.caller.StopLoop()
		log.Println("Closed As unforgivenessCloser decided so.")

	}*/
	remain := true
	for remain {
		select {
		case <-v.unforgivnesschan:
			log.Println("unforgivenessCloser() removing reminder unforgivness sign")
			break
		default:
			remain = false
		}
	}
	log.Println("unforgivenessCloser() quit")
} 


func (v *Escorting) EscortingUPV() {
	if v.escortProcess != nil {
		return
	}
	v.escortProcess = new([](*os.Process))
	go v.unforgivenessCloser()
}

func (v *Escorting) EscortingDown() {
	log.Println("escortingDown() Killing all escorted process ")
	if v.escortProcess == nil {
		return
	}
	for _, pr := range *v.escortProcess {
		pr.Kill()
	}
	log.Println("escortingDown() v.unforgivnesschan <- 0")
	select {
	case v.unforgivnesschan <- 0:
	}
	v.escortProcess = nil
}

func (v *Escorting) SetStatus(st *CoreI.Status) {
	v.status = st
}

func NewEscort() *Escorting {
	return &Escorting{unforgivnesschan: make(chan int)}
}

type Escorting struct {
	escortProcess    *[](*os.Process)
	unforgivnesschan chan int
	status           *CoreI.Status
}

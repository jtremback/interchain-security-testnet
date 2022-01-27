package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"reflect"
)

func main() {
	s := System{
		config: DefautlSystemConfig(),
	}

	s.startDocker()
	// println("docker started?")

	for _, step := range exampleSteps1 {
		s.runStep(step)
	}

	println("test completed")
}

func (s System) runStep(step Step) {
	switch action := step.action.(type) {
	case StartChainAction:
		s.startChain(action)
	case SendTokensAction:
		s.sendTokens(action)
	case SubmitGovProposalAction:
		s.submitGovProposal(action)
	case VoteGovProposalAction:
		s.voteGovProposal(action)
	}

	// Check state
	if !reflect.DeepEqual(step.state, s.getState()) {
		log.Fatal(`actual state ` + fmt.Sprint(s.getState()) + ` not equal to model state ` + fmt.Sprint(step.state))
	}

	// println(`actual state ` + fmt.Sprint(s.getState()) + ` equal to model state ` + fmt.Sprint(step.state))
}

func (s System) getState() State {
	return State{
		chain0: ChainState{
			// TODO: build map from chain validators list
			valBalances: map[uint]uint{
				0: s.getBalance(0, 0),
				1: s.getBalance(0, 1),
				// TODO: deal with validator2
			},
		},
		// TODO: deal with chain1
	}
}

func (s System) startDocker() {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	ports_string := ""

	for _, port := range s.config.exposePorts {
		ports_string = ports_string + " -p " + fmt.Sprint(port) + ":" + fmt.Sprint(port)
	}

	cmd := exec.Command("/bin/bash", path+"/start-docker.sh", s.config.containerName, s.config.instanceName, ports_string)

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(cmdReader)

	for scanner.Scan() {
		out := scanner.Text()
		// fmt.Println("startDocker: " + out)
		if out == "beacon!!!!!!!!!!" {
			return
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

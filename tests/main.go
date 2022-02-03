package main

import (
	"bufio"
	"encoding/json"
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

	actualState := s.getState()
	modelState := step.state

	marshal := func(x interface{}) string {
		bz, err := json.Marshal(x)
		if err != nil {
			log.Fatal(err)
		}

		return string(bz)
	}

	// Check state
	if !reflect.DeepEqual(actualState, modelState) {
		log.Fatal(`actual state ` + marshal(actualState) + ` not equal to model state ` + marshal(modelState))
	}

	println(`actual state ` + marshal(actualState) + ` equal to model state ` + marshal(modelState))
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

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"reflect"

	"github.com/kylelemons/godebug/pretty"
)

func main() {
	s := DefaultSystemConfig()

	s.startDocker()
	// println("docker started?")

	for _, step := range exampleSteps1 {
		s.runStep(step)
	}

	println("test completed")
}

func (s System) runStep(step Step) {
	fmt.Printf("%#v\n", step.action)
	switch action := step.action.(type) {
	case StartChainAction:
		s.startChain(action)
	case SendTokensAction:
		s.sendTokens(action)
	case SubmitTextProposalAction:
		s.submitTextProposal(action)
	case SubmitConsumerProposalAction:
		s.submitConsumerProposal(action)
	case VoteGovProposalAction:
		s.voteGovProposal(action)
	case StartConsumerChainAction:
		s.startConsumerChain(action)
	case AddChainToRelayerAction:
		s.addChainToRelayer(action)
	case AddIbcChannelAction:
		s.addIbcChannel(action)
	case RelayPacketsAction:
		s.relayPackets(action)
	}

	modelState := step.state
	actualState := s.getState(step.state)

	// Check state
	if !reflect.DeepEqual(actualState, modelState) {
		pretty.Print("actual state", actualState)
		pretty.Print("model state", modelState)
		log.Fatal(`actual state (-) not equal to model state (+): ` + pretty.Compare(actualState, modelState))
	}

	pretty.Print(actualState)
}

func (s System) startDocker() {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	ports_string := ""

	for _, port := range s.containerConfig.exposePorts {
		ports_string = ports_string + " -p " + fmt.Sprint(port) + ":" + fmt.Sprint(port)
	}

	cmd := exec.Command("/bin/bash", path+"/start-docker.sh", s.containerConfig.containerName, s.containerConfig.instanceName, ports_string)

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

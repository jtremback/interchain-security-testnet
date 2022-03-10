package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"reflect"

	"github.com/kylelemons/godebug/pretty"
)

var verbose = false

func main() {
	s := DefaultSystemConfig()

	s.startDocker()

	for _, step := range happyPathSteps {
		s.runStep(step, verbose)
	}

	println("test successful")
}

type ActionEntry struct {
	Executor func(interface{}, bool)
	Action   interface{}
}

type StepJSON struct {
	ActionType string
	Action     json.RawMessage
	state      State
}

func (s System) serializedStepsDraft() {
	actionEntries := map[string]ActionEntry{
		"StartChainAction": {
			Executor: s.startChain,
			Action:   StartChainAction{},
		},
	}

	steps := `{
        "actionType": "StartChainAction",
        "action": {
            "chain": 0,
            "validators": [{
                    "id": 1,
                    "stake": 500000000,
                    "allocation": 10000000000
                },
                {
                    "id": 0,
                    "stake": 500000000,
                    "allocation": 10000000000
                },
                {
                    "id": 2,
                    "stake": 500000000,
                    "allocation": 10000000000
                }
            ]
        },
        "state": {
            "0": {
                "ValBalances": {
                    "0": 9500000000,
                    "1": 9500000000
                }
            }
        }
    }`

	pStep := StepJSON{}

	// Unmarshal to get action type so that we can do the rest
	json.Unmarshal([]byte(steps), &pStep)

	action := actionEntries[pStep.ActionType].Action
	json.Unmarshal(pStep.Action, &action)

	actionEntries[pStep.ActionType].Executor(action, true)

	modelState := pStep.state
	actualState := s.getState(pStep.state)

	// Check state
	if !reflect.DeepEqual(actualState, modelState) {
		pretty.Print("actual state", actualState)
		pretty.Print("model state", modelState)
		log.Fatal(`actual state (-) not equal to model state (+): ` + pretty.Compare(actualState, modelState))
	}

	pretty.Print(actualState)
}

func (s System) runStep(step Step, verbose bool) {
	fmt.Printf("%#v\n", step.action)
	switch action := step.action.(type) {
	case StartChainAction:
		s.startChain(action, verbose)
	case SendTokensAction:
		s.sendTokens(action, verbose)
	case SubmitTextProposalAction:
		s.submitTextProposal(action, verbose)
	case SubmitConsumerProposalAction:
		s.submitConsumerProposal(action, verbose)
	case VoteGovProposalAction:
		s.voteGovProposal(action, verbose)
	case StartConsumerChainAction:
		s.startConsumerChain(action, verbose)
	case AddChainToRelayerAction:
		s.addChainToRelayer(action, verbose)
	case AddIbcConnectionAction:
		s.addIbcConnection(action, verbose)
	case AddIbcChannelAction:
		s.addIbcChannel(action, verbose)
	case RelayPacketsAction:
		s.relayPackets(action, verbose)
	case DelegateTokensAction:
		s.delegateTokens(action, verbose)
	default:
		log.Fatal(fmt.Sprintf(`unknown action: %#v`, action))
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
		log.Fatal(err)
	}

	cmd := exec.Command("/bin/bash", path+"/start-docker.sh", s.containerConfig.containerName, s.containerConfig.instanceName)

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
		if verbose {
			fmt.Println("startDocker: " + out)
		}
		if out == "beacon!!!!!!!!!!" {
			return
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

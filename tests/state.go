package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"time"

	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	"github.com/tidwall/gjson"
)

type State map[uint]ChainState

type ChainState struct {
	ValBalances *map[uint]uint
	Proposals   *map[uint]Proposal
}

type Proposal interface {
	isProposal()
}
type TextProposal struct {
	Title       string
	Description string
	Deposit     uint
	Status      string
}

func (p TextProposal) isProposal() {}

type ConsumerProposal struct {
	Deposit       uint
	Chain         uint
	SpawnTime     time.Time
	InitialHeight clienttypes.Height
	Status        string
}

func (p ConsumerProposal) isProposal() {}

type ConsumerGenesis struct {
}

func (s System) getState(modelState State) State {
	systemState := State{}
	for k, modelState := range modelState {
		println("getting state for chain", k)
		systemState[k] = s.getChainState(k, modelState)
	}

	return systemState
}

func (s System) getChainState(chain uint, modelState ChainState) ChainState {
	chainState := ChainState{}

	if modelState.ValBalances != nil {
		valBalances := s.getBalances(chain, *modelState.ValBalances)
		chainState.ValBalances = &valBalances
	}

	if modelState.Proposals != nil {
		proposals := s.getProposals(chain, *modelState.Proposals)
		chainState.Proposals = &proposals
	}

	return chainState
}

func (s System) getBalances(chain uint, modelState map[uint]uint) map[uint]uint {
	systemState := map[uint]uint{}
	for k, _ := range modelState {
		systemState[k] = s.getBalance(chain, k)
	}

	return systemState
}

func (s System) getProposals(chain uint, modelState map[uint]Proposal) map[uint]Proposal {
	systemState := map[uint]Proposal{}
	for k, _ := range modelState {
		systemState[k] = s.getProposal(chain, k)
	}

	return systemState
}

func (s System) getBalance(chain uint, validator uint) uint {
	bz, err := exec.Command("docker", "exec", s.containerConfig.instanceName, s.containerConfig.binaryName,

		"query", "bank", "balances",
		s.validatorConfigs[validator].delAddress,

		`--node`, "tcp://"+s.chainConfigs[chain].ipPrefix+".0:26658",
		`-o`, `json`,
	).CombinedOutput()

	if err != nil {
		log.Fatal(err, "\n", string(bz))
	}

	amount := gjson.Get(string(bz), `balances.#(denom=="stake").amount`)
	println("getting balance for chain, val", chain, validator)
	println("amount", amount.Uint())
	println("chainID", s.chainConfigs[chain].chainId)
	println("queryValHome", s.getQueryValidatorHome(chain))

	return uint(amount.Uint())
}

var noProposalRegex = regexp.MustCompile(`doesn't exist: key not found`)

// interchain-securityd query gov proposals
func (s System) getProposal(chain uint, proposal uint) Proposal {
	bz, err := exec.Command("docker", "exec", s.containerConfig.instanceName, s.containerConfig.binaryName,

		"query", "gov", "proposal",
		fmt.Sprint(proposal),

		`--node`, "tcp://"+s.chainConfigs[chain].ipPrefix+".0:26658",
		`-o`, `json`,
	).CombinedOutput()

	//TODO: throw error for proposal not found

	prop := TextProposal{}

	if err != nil {
		if noProposalRegex.Match(bz) {
			return prop
		}

		log.Fatal(err, "\n", string(bz))
	}

	propType := gjson.Get(string(bz), `content.@type`).String()
	deposit := gjson.Get(string(bz), `total_deposit.#(denom=="stake").amount`).Uint()
	status := gjson.Get(string(bz), `status`).String()

	switch propType {
	case "/cosmos.gov.v1beta1.TextProposal":
		title := gjson.Get(string(bz), `content.title`).String()
		description := gjson.Get(string(bz), `content.description`).String()

		return TextProposal{
			Deposit:     uint(deposit),
			Status:      status,
			Title:       title,
			Description: description,
		}
	case "/interchain_security.ccv.parent.v1.CreateChildChainProposal":
		chainId := gjson.Get(string(bz), `content.chain_id`).String()
		spawnTime := gjson.Get(string(bz), `content.spawn_time`).Time()

		var chain uint
		for i, conf := range s.chainConfigs {
			if conf.chainId == chainId {
				chain = uint(i)
				break
			}
		}

		return ConsumerProposal{
			Deposit:   uint(deposit),
			Status:    status,
			Chain:     chain,
			SpawnTime: spawnTime.UTC(),
			InitialHeight: clienttypes.Height{
				RevisionNumber: gjson.Get(string(bz), `content.initial_height.revision_number`).Uint(),
				RevisionHeight: gjson.Get(string(bz), `content.initial_height.revision_height`).Uint(),
			},
		}

	}

	log.Fatal("unknown proposal type", string(bz))

	return nil
}

// func (s System) getConsumerGenesis(chain uint, validator uint, consumerChainId string) uint {
// 	bz, err := exec.Command("docker", "exec", s.containerConfig.instanceName, s.containerConfig.binaryName,

// 		"query", "parent", "child-genesis",
// 		consumerChainId,

// 		`--chain-id`, s.chainConfigs[chain].chainId,
// 		`--home`, s.getQueryValidatorHome(chain),
// 		`-o`, `json`,
// 	).CombinedOutput()

// 	if err != nil {
// 		log.Fatal(err, "\n", string(bz))
// 	}

// 	amount := gjson.Get(string(bz), `balances.#(denom=="stake").amount`)

// 	return uint(amount.Uint())
// }

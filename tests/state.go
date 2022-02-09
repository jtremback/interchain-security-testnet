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
}

func (p TextProposal) isProposal() {}

type ConsumerProposal struct {
	Deposit       uint
	ChainId       string
	SpawnTime     time.Time
	InitialHeight clienttypes.Height
}

func (p ConsumerProposal) isProposal() {}

func (s System) getState(modelState State) State {
	systemState := State{}
	for k, modelState := range modelState {
		systemState[k] = s.getChainState(modelState)
	}

	return systemState
}

func (s System) getChainState(modelState ChainState) ChainState {
	chainState := ChainState{}

	if modelState.ValBalances != nil {
		valBalances := s.getBalances(0, *modelState.ValBalances)
		chainState.ValBalances = &valBalances
	}

	if modelState.Proposals != nil {
		proposals := s.getProposals(0, 0, *modelState.Proposals)
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

func (s System) getProposals(chain uint, validator uint, modelState map[uint]Proposal) map[uint]Proposal {
	systemState := map[uint]Proposal{}
	for k, _ := range modelState {
		systemState[k] = s.getProposal(chain, validator, k)
	}

	return systemState
}

func (s System) getBalance(chain uint, validator uint) uint {
	bz, err := exec.Command("docker", "exec", "interchain-security-instance", "interchain-securityd", "query", "bank", "balances",
		s.config.validatorAttrs[validator].delAddress,
		`--chain-id`, s.config.chainAttrs[chain].chainId,
		`--home`, `/provider/validator`+fmt.Sprint(validator),
		`-o`, `json`,
	).CombinedOutput()

	if err != nil {
		log.Fatal(err, "\n", string(bz))
	}

	amount := gjson.Get(string(bz), `balances.#(denom=="stake").amount`)

	return uint(amount.Uint())
}

var noProposalRegex = regexp.MustCompile(`doesn't exist: key not found`)

// interchain-securityd query gov proposals
func (s System) getProposal(chain uint, validator uint, proposal uint) Proposal {
	bz, err := exec.Command("docker", "exec", "interchain-security-instance", "interchain-securityd", "query", "gov", "proposal",
		fmt.Sprint(proposal),
		`--chain-id`, s.config.chainAttrs[chain].chainId,
		`--home`, `/provider/validator`+fmt.Sprint(validator),
		`-o`, `json`,
	).CombinedOutput()

	prop := TextProposal{}

	if err != nil {
		if noProposalRegex.Match(bz) {
			return prop
		}

		log.Fatal(err, "\n", string(bz))
	}

	propType := gjson.Get(string(bz), `content.@type`).String()

	switch propType {
	case "/cosmos.gov.v1beta1.TextProposal":
		title := gjson.Get(string(bz), `content.title`).String()
		description := gjson.Get(string(bz), `content.description`).String()
		deposit := gjson.Get(string(bz), `total_deposit.#(denom=="stake").amount`).Uint()

		return TextProposal{
			Title:       title,
			Description: description,
			Deposit:     uint(deposit),
		}
	case "/interchain_security.ccv.parent.v1.CreateChildChainProposal":
		deposit := gjson.Get(string(bz), `total_deposit.#(denom=="stake").amount`).Uint()
		chainId := gjson.Get(string(bz), `content.chain_id`).String()
		spawnTime := gjson.Get(string(bz), `content.spawn_time`).Time()

		return ConsumerProposal{
			Deposit:   uint(deposit),
			ChainId:   chainId,
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

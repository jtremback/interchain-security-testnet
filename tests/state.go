package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

type State map[uint]ChainState

type ChainState struct {
	ValBalances *map[uint]uint
	Proposals   *map[uint]Proposal
	ValPowers   *map[uint]uint
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

	if modelState.ValPowers != nil {
		powers := s.getValPowers(chain, *modelState.ValPowers)
		chainState.ValPowers = &powers
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

func (s System) getValPowers(chain uint, modelState map[uint]uint) map[uint]uint {
	systemState := map[uint]uint{}
	for k, _ := range modelState {
		systemState[k] = s.getValPower(chain, k)
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

type TmValidatorSetYaml struct {
	Total      string `yaml:"total"`
	Validators []struct {
		Address     string    `yaml:"address"`
		VotingPower string    `yaml:"voting_power"`
		PubKey      ValPubKey `yaml:"pub_key"`
	}
}

type ValPubKey struct {
	Value string `yaml:"value"`
}

func (s System) getValPower(chain uint, validator uint) uint {
	bz, err := exec.Command("docker", "exec", s.containerConfig.instanceName, s.containerConfig.binaryName,

		"query", "tendermint-validator-set",

		`--node`, "tcp://"+s.chainConfigs[chain].ipPrefix+".0:26658",
	).CombinedOutput()

	//TODO: throw error for proposal not found

	valset := TmValidatorSetYaml{}

	err = yaml.Unmarshal(bz, &valset)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	total, err := strconv.Atoi(valset.Total)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	if total != len(valset.Validators) {
		log.Fatalf("Total number of validators %v does not match number of validators in list %v. Probably a query pagination issue.",
			valset.Total, uint(len(valset.Validators)))
	}

	for _, val := range valset.Validators {
		if val.Address == s.validatorConfigs[validator].valconsAddress {
			votingPower, err := strconv.Atoi(val.VotingPower)
			if err != nil {
				log.Fatalf("error: %v", err)
			}

			return uint(votingPower)
		}
	}

	log.Fatalf("Validator %v not in tendermint validator set", validator)

	return 0
}

// block_height: "4662"
// total: "3"
// validators:
// - address: cosmosvalcons19g577x902tuxgl725w0cedjkz9pkanqvkwwq3m
//   proposer_priority: "88"
//   pub_key:
//     type: tendermint/PubKeyEd25519
//     value: 8cDvCpTF99eF+2j8wNGomWXS0j4tfFI9XlMAddLT7zk=
//   voting_power: "501"
// - address: cosmosvalcons1g0ljre2k3m4adfz9fjlf3wcqnxgednydzsmkt9
//   proposer_priority: "-44"
//   pub_key:
//     type: tendermint/PubKeyEd25519
//     value: 4/a0hWG9tCU9sxhs8fQsHjDSZxVTmp59eqNJBvFz6Xw=
//   voting_power: "500"
// - address: cosmosvalcons1m5eahy3m8edkzy7cl0nm8jx6qt2zhkr0zy500u
//   proposer_priority: "-44"
//   pub_key:
//     type: tendermint/PubKeyEd25519
//     value: iamj/w3/1g96isN2DpS3c3OHxvfLHHqkJttxea4jEyQ=
//   voting_power: "500"

// block_height: "48"
// total: "3"
// validators:
// - address: cosmosvalcons1x4n0ger88vhsd96cmt5xlsjaqsqhhv28yrlhjv
//   proposer_priority: "154"
//   pub_key:
//     type: tendermint/PubKeyEd25519
//     value: XrLjKdc4mB2gfqplvnoySjSJq2E90RynUwaO3WhJutk=
//   voting_power: "511"
// - address: cosmosvalcons14ysnwpck6khrev6ftuxelgsw7rnqe3rkyu4rg9
//   proposer_priority: "-77"
//   pub_key:
//     type: tendermint/PubKeyEd25519
//     value: xYBNRAScX+h3mE/VKq+rTAwcNNeUzxG83J8cDPFwBHc=
//   voting_power: "500"
// - address: cosmosvalcons1mukrq2v8c9y3mm6y0l0eezx4fts8a05eykr2v5
//   proposer_priority: "-77"
//   pub_key:
//     type: tendermint/PubKeyEd25519
//     value: Qf7hUv1IhLeMKdCK6J4fVRajur6I0IV2wARrzwNDK8o=
//   voting_power: "500"

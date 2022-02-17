package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"

	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
)

type StartChainAction struct {
	chain          uint
	validators     []uint
	genesisChanges string
	skipGentx      bool
	copyConfigs    string
}

type StartConsumerChainAction struct {
	consumerChain uint
	providerChain uint
	validators    []uint
}

type SendTokensAction struct {
	chain  uint
	from   uint
	to     uint
	amount uint
}

type SubmitTextProposalAction struct {
	chain       uint
	from        uint
	deposit     uint
	propType    string
	title       string
	description string
}

type SubmitConsumerProposalAction struct {
	chain         uint
	from          uint
	deposit       uint
	consumerChain uint
	spawnTime     time.Time
	initialHeight clienttypes.Height
}

type VoteGovProposalAction struct {
	chain      uint
	from       []uint
	vote       []string
	propNumber uint
}

// TODO: import this directly from the module once it is merged
type CreateChildChainProposalJSON struct {
	Title         string             `json:"title"`
	Description   string             `json:"description"`
	ChainId       string             `json:"chain_id"`
	InitialHeight clienttypes.Height `json:"initial_height"`
	GenesisHash   []byte             `json:"genesis_hash"`
	BinaryHash    []byte             `json:"binary_hash"`
	SpawnTime     time.Time          `json:"spawn_time"`
	Deposit       string             `json:"deposit"`
}

func (s System) sendTokens(action SendTokensAction) {
	bz, err := exec.Command("docker", "exec", s.containerConfig.instanceName, s.containerConfig.binaryName,

		"tx", "bank", "send",
		s.validatorConfigs[action.from].delAddress,
		s.validatorConfigs[action.to].delAddress,
		fmt.Sprint(action.amount)+`stake`,

		`--chain-id`, s.chainConfigs[action.chain].chainId,
		`--home`, s.getTxValidatorHome(action.chain, action.from),
		`--node`, "tcp://"+s.chainConfigs[action.chain].ipPrefix+".0:26658",
		`--keyring-backend`, `test`,
		`-b`, `block`,
		`-y`,
	).CombinedOutput()

	if err != nil {
		log.Fatal(err, "\n", string(bz))
	}
}

func (s System) startChain(
	action StartChainAction,
) {
	chainConfig := s.chainConfigs[action.chain]
	var mnemonics []string

	for _, val := range action.validators {
		mnemonics = append(mnemonics, s.validatorConfigs[val].mnemonic)
	}

	mnz, err := json.Marshal(mnemonics)
	if err != nil {
		log.Fatal(err)
	}

	var genesisChanges string
	if action.genesisChanges != "" {
		genesisChanges = chainConfig.genesisChanges + " | " + action.genesisChanges
	} else {
		genesisChanges = chainConfig.genesisChanges
	}

	cmd := exec.Command("docker", "exec", s.containerConfig.instanceName, "/bin/bash",
		"/testnet-scripts/start-chain/start-chain.sh", s.containerConfig.binaryName, string(mnz), chainConfig.chainId, chainConfig.ipPrefix,
		fmt.Sprint(chainConfig.rpcPort), fmt.Sprint(chainConfig.grpcPort), genesisChanges,
		chainConfig.initialAllocation, chainConfig.stakeAmount, fmt.Sprint(action.skipGentx), action.copyConfigs)

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
		fmt.Println("startChain: " + out)
		if out == "done!!!!!!!!" {
			return
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func (s System) submitTextProposal(
	action SubmitTextProposalAction,
) {
	// docker exec interchain-security-instance interchain-securityd tx gov submit-proposal --title="Test Proposal" --description="My awesome proposal" --type Text --deposit 10000000stake --from validator1 --chain-id provider --home /provider/validator1 --keyring-backend test
	bz, err := exec.Command("docker", "exec", s.containerConfig.instanceName, s.containerConfig.binaryName,

		"tx", "gov", "submit-proposal",
		`--title`, action.title,
		`--description`, action.description,
		`--type`, action.propType,
		`--deposit`, fmt.Sprint(action.deposit)+`stake`,

		`--from`, `validator`+fmt.Sprint(action.from),
		`--chain-id`, s.chainConfigs[action.chain].chainId,
		`--home`, `/provider/validator`+fmt.Sprint(action.from),
		`--node`, "tcp://"+s.chainConfigs[action.chain].ipPrefix+".0:26658",
		`--keyring-backend`, `test`,
		`-b`, `block`,
		`-y`,
	).CombinedOutput()

	if err != nil {
		log.Fatal(err, "\n", string(bz))
	}
}

func (s System) submitConsumerProposal(
	action SubmitConsumerProposalAction,
) {
	prop := CreateChildChainProposalJSON{
		Title:         "Create a chain",
		Description:   "Gonna be a great chain",
		ChainId:       s.chainConfigs[action.consumerChain].chainId,
		InitialHeight: action.initialHeight,
		GenesisHash:   []byte("gen_hash"),
		BinaryHash:    []byte("bin_hash"),
		SpawnTime:     action.spawnTime,
		Deposit:       fmt.Sprint(action.deposit) + `stake`,
	}

	bz, err := json.Marshal(prop)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: cleanup file
	bz, err = exec.Command("docker", "exec", s.containerConfig.instanceName, "/bin/bash", "-c", fmt.Sprintf(`echo '%s' > %s`, string(bz), "/temp-proposal.json")).CombinedOutput()

	if err != nil {
		log.Fatal(err, "\n", string(bz))
	}

	bz, err = exec.Command("docker", "exec", s.containerConfig.instanceName, s.containerConfig.binaryName,

		"tx", "gov", "submit-proposal", "create-child-chain",
		"/temp-proposal.json",

		`--from`, `validator`+fmt.Sprint(action.from),
		`--chain-id`, s.chainConfigs[action.chain].chainId,
		`--home`, `/provider/validator`+fmt.Sprint(action.from),
		`--node`, "tcp://"+s.chainConfigs[action.chain].ipPrefix+".0:26658",
		`--keyring-backend`, `test`,
		`-b`, `block`,
		`-y`,
	).CombinedOutput()

	if err != nil {
		log.Fatal(err, "\n", string(bz))
	}
}

func (s System) voteGovProposal(
	action VoteGovProposalAction,
) {
	var wg sync.WaitGroup
	for i, val := range action.from {
		wg.Add(1)
		vote := action.vote[i]
		go func(val uint, vote string) {
			defer wg.Done()
			bz, err := exec.Command("docker", "exec", s.containerConfig.instanceName, s.containerConfig.binaryName,

				"tx", "gov", "vote",
				fmt.Sprint(action.propNumber), vote,

				`--from`, `validator`+fmt.Sprint(val),
				`--chain-id`, s.chainConfigs[action.chain].chainId,
				`--home`, `/provider/validator`+fmt.Sprint(val),
				`--node`, "tcp://"+s.chainConfigs[action.chain].ipPrefix+".0:26658",
				`--keyring-backend`, `test`,
				`-b`, `block`,
				`-y`,
			).CombinedOutput()

			if err != nil {
				log.Fatal(err, "\n", string(bz))
			}
		}(val, vote)
	}

	wg.Wait()
	time.Sleep(time.Duration(s.chainConfigs[action.chain].votingWaitTime) * time.Second)
}

func (s System) startConsumerChain(action StartConsumerChainAction) {
	bz, err := exec.Command("docker", "exec", s.containerConfig.instanceName, s.containerConfig.binaryName,

		"query", "parent", "child-genesis",
		s.chainConfigs[action.consumerChain].chainId,

		`--node`, "tcp://"+s.chainConfigs[action.providerChain].ipPrefix+".0:26658",
		`-o`, `json`,
	).CombinedOutput()

	if err != nil {
		log.Fatal(err, "\n", string(bz))
	}

	s.startChain(StartChainAction{
		chain:          1,
		validators:     action.validators,
		genesisChanges: ".app_state.ccvchild = " + string(bz),
		skipGentx:      true,
		copyConfigs:    s.chainConfigs[action.providerChain].chainId,
	})
}

func (s System) getQueryValidatorHome(chain uint) string {
	// Get first subdirectory of the directory of this chain, which will be the home directory of one of the validators
	bz, err := exec.Command("docker", "exec", s.containerConfig.instanceName, "bash", "-c", `cd /`+s.chainConfigs[chain].chainId+`; ls -d */ | awk '{print $1}' | head -n 1`).CombinedOutput()

	if err != nil {
		log.Fatal(err, "\n", string(bz))
	}

	return `/` + s.chainConfigs[chain].chainId + `/` + string(bz)
}

func (s System) getTxValidatorHome(chain uint, validator uint) string {
	return `/` + s.chainConfigs[chain].chainId + `/validator` + fmt.Sprint(validator)
}

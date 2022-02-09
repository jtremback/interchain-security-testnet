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

type System struct {
	config Config
}

type StartChainAction struct {
	chain      uint
	validators []uint
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
	chainId       string
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
	// docker exec interchain-security-instance interchain-securityd tx bank send cosmos19pe9pg5dv9k5fzgzmsrgnw9rl9asf7ddwhu7lm cosmos1dkas8mu4kyhl5jrh4nzvm65qz588hy9qcz08la 1stake --home /provider/validator1 --keyring-backend test --chain-id provider -y
	bz, err := exec.Command("docker", "exec", "interchain-security-instance", "interchain-securityd", "tx", "bank", "send",
		s.config.validatorAttrs[action.from].delAddress,
		s.config.validatorAttrs[action.to].delAddress,
		fmt.Sprint(action.amount)+`stake`,
		`--chain-id`, s.config.chainAttrs[action.chain].chainId,
		`--home`, `/provider/validator`+fmt.Sprint(action.from),
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
	c := s.config.chainAttrs[action.chain]
	var mnemonics []string

	for _, val := range action.validators {
		mnemonics = append(mnemonics, s.config.validatorAttrs[val].mnemonic)
	}

	mnz, err := json.Marshal(mnemonics)
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("docker", "exec", s.config.instanceName, "/bin/bash",
		s.config.startChainScript, s.config.binaryName, string(mnz), c.chainId, c.ipPrefix,
		fmt.Sprint(c.rpcPort), fmt.Sprint(c.grpcPort), c.genesisChanges, s.config.initialAllocation, s.config.stakeAmount)

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
		// fmt.Println("startChain: " + out)
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
	bz, err := exec.Command("docker", "exec", "interchain-security-instance", "interchain-securityd", "tx", "gov", "submit-proposal",
		`--title`, action.title,
		`--description`, action.description,
		`--type`, action.propType,
		`--deposit`, fmt.Sprint(action.deposit)+`stake`,

		`--from`, `validator`+fmt.Sprint(action.from),
		`--chain-id`, s.config.chainAttrs[action.chain].chainId,
		`--home`, `/provider/validator`+fmt.Sprint(action.from),
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
		ChainId:       action.chainId,
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
	bz, err = exec.Command("docker", "exec", "interchain-security-instance", "/bin/bash", "-c", fmt.Sprintf(`echo '%s' > %s`, string(bz), "/temp-proposal.json")).CombinedOutput()

	if err != nil {
		log.Fatal(err, "\n", string(bz))
	}

	bz, err = exec.Command("docker", "exec", "interchain-security-instance", "interchain-securityd", "tx", "gov", "submit-proposal", "create-child-chain",
		"/temp-proposal.json",

		`--from`, `validator`+fmt.Sprint(action.from),
		`--chain-id`, s.config.chainAttrs[action.chain].chainId,
		`--home`, `/provider/validator`+fmt.Sprint(action.from),
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
			bz, err := exec.Command("docker", "exec", "interchain-security-instance", "interchain-securityd", "tx", "gov", "vote",
				fmt.Sprint(action.propNumber), vote,

				`--from`, `validator`+fmt.Sprint(val),
				`--chain-id`, s.config.chainAttrs[action.chain].chainId,
				`--home`, `/provider/validator`+fmt.Sprint(val),
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
	time.Sleep(60 * time.Second)
}

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
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

type SubmitGovProposalAction struct {
	chain       uint
	from        uint
	deposit     uint
	propType    string
	title       string
	description string
}

type VoteGovProposalAction struct {
	chain      uint
	from       uint
	vote       string
	propNumber uint
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

func (s System) submitGovProposal(
	action SubmitGovProposalAction,
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

func (s System) voteGovProposal(
	action VoteGovProposalAction,
) {
	// docker exec interchain-security-instance interchain-securityd tx gov submit-proposal --title="Test Proposal" --description="My awesome proposal" --type Text --deposit 10000000stake --from validator1 --chain-id provider --home /provider/validator1 --keyring-backend test
	bz, err := exec.Command("docker", "exec", "interchain-security-instance", "interchain-securityd", "tx", "gov", "vote",
		`vote`, fmt.Sprint(action.propNumber), action.vote,

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

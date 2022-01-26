package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	"github.com/tidwall/gjson"
)

type System struct {
	config Config
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

func (s System) startChain(
	chain uint,
	validators []uint,
) {
	c := s.config.chainAttrs[chain]
	var mnemonics []string

	for _, val := range validators {
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

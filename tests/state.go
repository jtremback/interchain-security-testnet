package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/tidwall/gjson"
)

type State struct {
	Chain0 ChainState
	Chain1 ChainState
}

type ChainState struct {
	ValBalances map[uint]uint
	Proposals   map[uint]TextProposal
}

type TextProposal struct {
	Title       string
	Description string
	Deposit     uint
	From        uint
}

func (s System) getState() State {
	return State{
		Chain0: ChainState{
			// TODO: build map from chain validators list
			ValBalances: map[uint]uint{
				0: s.getBalance(0, 0),
				1: s.getBalance(0, 1),
				// TODO: deal with validator2
			},
		},
		// TODO: deal with chain1
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

// interchain-securityd query gov proposals
func (s System) getProposals(chain uint, validator uint) uint {
	bz, err := exec.Command("docker", "exec", "interchain-security-instance", "interchain-securityd", "query", "gov", "proposals",
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

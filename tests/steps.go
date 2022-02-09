package main

import (
	"time"

	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
)

type Step struct {
	action interface{}
	state  State
}

var now = time.Now().UTC()

var exampleSteps0 = []Step{
	{
		action: StartChainAction{
			chain:      0,
			validators: []uint{0, 1, 2},
		},
		state: State{
			0: ChainState{
				ValBalances: &map[uint]uint{
					0: 9500000000,
					1: 9500000000,
				},
			},
		},
	},
	{
		action: SendTokensAction{
			chain:  0,
			from:   0,
			to:     1,
			amount: 1,
		},
		state: State{
			0: ChainState{
				ValBalances: &map[uint]uint{
					0: 9499999999,
					1: 9500000001,
				},
			},
		},
	},
	{
		action: SubmitTextProposalAction{
			chain:       0,
			from:        0,
			deposit:     1000000,
			propType:    "Text",
			title:       "Prop title",
			description: "description",
		},
		state: State{
			0: ChainState{
				ValBalances: &map[uint]uint{
					0: 9498999999,
					1: 9500000001,
				},
				Proposals: &map[uint]Proposal{
					1: TextProposal{
						Title:       "Prop title",
						Description: "description",
						Deposit:     1000000,
					},
				},
			},
		},
	},
}

var exampleSteps1 = []Step{
	{
		action: StartChainAction{
			chain:      0,
			validators: []uint{0, 1, 2},
		},
		state: State{
			0: ChainState{
				ValBalances: &map[uint]uint{
					0: 9500000000,
					1: 9500000000,
				},
			},
		},
	},
	{
		action: SubmitConsumerProposalAction{
			chain:         0,
			from:          0,
			deposit:       10000001,
			chainId:       "consumer",
			spawnTime:     now,
			initialHeight: clienttypes.Height{0, 1},
		},
		state: State{
			0: ChainState{
				ValBalances: &map[uint]uint{
					0: 9489999999,
					1: 9500000000,
				},
				Proposals: &map[uint]Proposal{
					1: ConsumerProposal{
						Deposit:       10000001,
						ChainId:       "consumer",
						SpawnTime:     now,
						InitialHeight: clienttypes.Height{0, 1},
						Status:        "PROPOSAL_STATUS_VOTING_PERIOD",
					},
				},
			},
		},
	},
	{
		action: VoteGovProposalAction{
			chain:      0,
			from:       []uint{0, 1, 2},
			vote:       []string{"yes", "yes", "yes"},
			propNumber: 1,
		},
		state: State{
			0: ChainState{
				Proposals: &map[uint]Proposal{
					1: ConsumerProposal{
						Deposit:       10000001,
						ChainId:       "consumer",
						SpawnTime:     now,
						InitialHeight: clienttypes.Height{0, 1},
						Status:        "PROPOSAL_STATUS_PASSED",
					},
				},
			},
		},
	},
}

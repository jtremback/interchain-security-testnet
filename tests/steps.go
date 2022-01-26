package main

type State struct {
	chain0 ChainState
	chain1 ChainState
}

type ChainState struct {
	valBalances map[uint]uint
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

type Step struct {
	action interface{}
	state  State
}

var exampleSteps1 = []Step{
	{
		action: StartChainAction{
			chain:      0,
			validators: []uint{0, 1, 2},
		},
		state: State{
			chain0: ChainState{
				valBalances: map[uint]uint{
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
			chain0: ChainState{
				valBalances: map[uint]uint{
					0: 9499999999,
					1: 9500000001,
				},
			},
		},
	},
}

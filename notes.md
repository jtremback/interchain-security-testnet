	// Pass gov prop
	// Query IBC module for client state and consensus state????? <- NO but should be possible
	
	// Instead, create client state:
	// clientState = ibctmtypes.NewClientState(<provider chain id>, ibctmtypes.DefaultTrustLevel, <unbonding period- query provider staking params>, <half of the previous argument>,
	// 		time.Second*10, <height at which the proposal passed on provider +-1>, commitmenttypes.GetSDKSpecs(), []string{"upgrade", "upgradedIBCState"}, true, true),
	
	// Then, create consensus state:
	// ConsensusState = ibctmtypes.NewConsensusState(<time at which the proposal passed on provider +-1>, commitmenttypes.NewMerkleRoot(<AppHash from block header from when the proposal passed on the provider +-1>), <NextValidatorsHash from block header from when the proposal passed on the provider +-1>)
	// Populate consumer genesis - set enabled to true
	// start consumer chain

	// Set up relayer- Hermes may be best
	// https://hermes.informal.systems/config.html
	// - rpc endpoint
	// - private keys of an address with money
	// - probably mostly change id-key_name in config
	// - Massive headache: off by one errors

	consState = GetSelfConsensusState(ctx.blockheight (+-1???))


	clientState := k.GetTemplateClient(ctx)
	clientState.ChainId = ctx.chainid
	clientState.LatestHeight = ctx.blockheight (+-1???)
	clientState.TrustingPeriod = unbondingTime / 2
	clientState.UnbondingPeriod = unbondingTime

	gen.initialvalset = tm.validatorset -> utility func

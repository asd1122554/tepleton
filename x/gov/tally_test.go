package gov

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/tepleton/tepleton-sdk/types"
	wrsp "github.com/tepleton/tepleton/wrsp/types"
	"github.com/tepleton/tepleton/crypto"

	"github.com/tepleton/tepleton-sdk/x/stake"
)

func TestTallyNoOneVotes(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10)
	mapp.BeginBlock(wrsp.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, wrsp.Header{})
	stakeHandler := stake.NewHandler(sk)

	dummyDescription := stake.NewDescription("T", "E", "S", "T")
	val1CreateMsg := stake.NewMsgCreateValidator(addrs[0], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 5), dummyDescription)
	stakeHandler(ctx, val1CreateMsg)
	val2CreateMsg := stake.NewMsgCreateValidator(addrs[1], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 5), dummyDescription)
	stakeHandler(ctx, val2CreateMsg)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	passes, _ := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.False(t, passes)
}

func TestTallyOnlyValidatorsAllYes(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10)
	mapp.BeginBlock(wrsp.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, wrsp.Header{})
	stakeHandler := stake.NewHandler(sk)

	dummyDescription := stake.NewDescription("T", "E", "S", "T")
	val1CreateMsg := stake.NewMsgCreateValidator(addrs[0], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 5), dummyDescription)
	res := stakeHandler(ctx, val1CreateMsg)
	require.True(t, res.IsOK())
	val2CreateMsg := stake.NewMsgCreateValidator(addrs[1], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 5), dummyDescription)
	res = stakeHandler(ctx, val2CreateMsg)
	require.True(t, res.IsOK())

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionYes)
	require.Nil(t, err)

	passes, _ := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.True(t, passes)
}

func TestTallyOnlyValidators51No(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10)
	mapp.BeginBlock(wrsp.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, wrsp.Header{})
	stakeHandler := stake.NewHandler(sk)

	dummyDescription := stake.NewDescription("T", "E", "S", "T")
	val1CreateMsg := stake.NewMsgCreateValidator(addrs[0], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 5), dummyDescription)
	stakeHandler(ctx, val1CreateMsg)
	val2CreateMsg := stake.NewMsgCreateValidator(addrs[1], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 6), dummyDescription)
	stakeHandler(ctx, val2CreateMsg)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionNo)
	require.Nil(t, err)

	passes, _ := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.False(t, passes)
}

func TestTallyOnlyValidators51Yes(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10)
	mapp.BeginBlock(wrsp.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, wrsp.Header{})
	stakeHandler := stake.NewHandler(sk)

	dummyDescription := stake.NewDescription("T", "E", "S", "T")
	val1CreateMsg := stake.NewMsgCreateValidator(addrs[0], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 6), dummyDescription)
	stakeHandler(ctx, val1CreateMsg)
	val2CreateMsg := stake.NewMsgCreateValidator(addrs[1], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 6), dummyDescription)
	stakeHandler(ctx, val2CreateMsg)
	val3CreateMsg := stake.NewMsgCreateValidator(addrs[2], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 7), dummyDescription)
	stakeHandler(ctx, val3CreateMsg)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionNo)
	require.Nil(t, err)

	passes, _ := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.True(t, passes)
}

func TestTallyOnlyValidatorsVetoed(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10)
	mapp.BeginBlock(wrsp.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, wrsp.Header{})
	stakeHandler := stake.NewHandler(sk)

	dummyDescription := stake.NewDescription("T", "E", "S", "T")
	val1CreateMsg := stake.NewMsgCreateValidator(addrs[0], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 6), dummyDescription)
	stakeHandler(ctx, val1CreateMsg)
	val2CreateMsg := stake.NewMsgCreateValidator(addrs[1], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 6), dummyDescription)
	stakeHandler(ctx, val2CreateMsg)
	val3CreateMsg := stake.NewMsgCreateValidator(addrs[2], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 7), dummyDescription)
	stakeHandler(ctx, val3CreateMsg)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionNoWithVeto)
	require.Nil(t, err)

	passes, _ := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.False(t, passes)
}

func TestTallyOnlyValidatorsAbstainPasses(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10)
	mapp.BeginBlock(wrsp.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, wrsp.Header{})
	stakeHandler := stake.NewHandler(sk)

	dummyDescription := stake.NewDescription("T", "E", "S", "T")
	val1CreateMsg := stake.NewMsgCreateValidator(addrs[0], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 6), dummyDescription)
	stakeHandler(ctx, val1CreateMsg)
	val2CreateMsg := stake.NewMsgCreateValidator(addrs[1], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 6), dummyDescription)
	stakeHandler(ctx, val2CreateMsg)
	val3CreateMsg := stake.NewMsgCreateValidator(addrs[2], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 7), dummyDescription)
	stakeHandler(ctx, val3CreateMsg)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionAbstain)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionNo)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionYes)
	require.Nil(t, err)

	passes, _ := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.True(t, passes)
}

func TestTallyOnlyValidatorsAbstainFails(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10)
	mapp.BeginBlock(wrsp.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, wrsp.Header{})
	stakeHandler := stake.NewHandler(sk)

	dummyDescription := stake.NewDescription("T", "E", "S", "T")
	val1CreateMsg := stake.NewMsgCreateValidator(addrs[0], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 6), dummyDescription)
	stakeHandler(ctx, val1CreateMsg)
	val2CreateMsg := stake.NewMsgCreateValidator(addrs[1], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 6), dummyDescription)
	stakeHandler(ctx, val2CreateMsg)
	val3CreateMsg := stake.NewMsgCreateValidator(addrs[2], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 7), dummyDescription)
	stakeHandler(ctx, val3CreateMsg)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionAbstain)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionNo)
	require.Nil(t, err)

	passes, _ := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.False(t, passes)
}

func TestTallyOnlyValidatorsNonVoter(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10)
	mapp.BeginBlock(wrsp.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, wrsp.Header{})
	stakeHandler := stake.NewHandler(sk)

	dummyDescription := stake.NewDescription("T", "E", "S", "T")
	val1CreateMsg := stake.NewMsgCreateValidator(addrs[0], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 6), dummyDescription)
	stakeHandler(ctx, val1CreateMsg)
	val2CreateMsg := stake.NewMsgCreateValidator(addrs[1], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 6), dummyDescription)
	stakeHandler(ctx, val2CreateMsg)
	val3CreateMsg := stake.NewMsgCreateValidator(addrs[2], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 7), dummyDescription)
	stakeHandler(ctx, val3CreateMsg)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[1], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionNo)
	require.Nil(t, err)

	passes, nonVoting := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.False(t, passes)
	require.Equal(t, 1, len(nonVoting))
	require.Equal(t, addrs[0], nonVoting[0])
}

func TestTallyDelgatorOverride(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10)
	mapp.BeginBlock(wrsp.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, wrsp.Header{})
	stakeHandler := stake.NewHandler(sk)

	dummyDescription := stake.NewDescription("T", "E", "S", "T")
	val1CreateMsg := stake.NewMsgCreateValidator(addrs[0], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 5), dummyDescription)
	stakeHandler(ctx, val1CreateMsg)
	val2CreateMsg := stake.NewMsgCreateValidator(addrs[1], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 6), dummyDescription)
	stakeHandler(ctx, val2CreateMsg)
	val3CreateMsg := stake.NewMsgCreateValidator(addrs[2], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 7), dummyDescription)
	stakeHandler(ctx, val3CreateMsg)

	delegator1Msg := stake.NewMsgDelegate(addrs[3], addrs[2], sdk.NewCoin("steak", 30))
	stakeHandler(ctx, delegator1Msg)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[3], OptionNo)
	require.Nil(t, err)

	passes, _ := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.False(t, passes)
}

func TestTallyDelgatorInherit(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10)
	mapp.BeginBlock(wrsp.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, wrsp.Header{})
	stakeHandler := stake.NewHandler(sk)

	dummyDescription := stake.NewDescription("T", "E", "S", "T")
	val1CreateMsg := stake.NewMsgCreateValidator(addrs[0], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 5), dummyDescription)
	stakeHandler(ctx, val1CreateMsg)
	val2CreateMsg := stake.NewMsgCreateValidator(addrs[1], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 6), dummyDescription)
	stakeHandler(ctx, val2CreateMsg)
	val3CreateMsg := stake.NewMsgCreateValidator(addrs[2], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 7), dummyDescription)
	stakeHandler(ctx, val3CreateMsg)

	delegator1Msg := stake.NewMsgDelegate(addrs[3], addrs[2], sdk.NewCoin("steak", 30))
	stakeHandler(ctx, delegator1Msg)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionNo)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionNo)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionYes)
	require.Nil(t, err)

	passes, nonVoting := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.True(t, passes)
	require.Equal(t, 0, len(nonVoting))
}

func TestTallyDelgatorMultipleOverride(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10)
	mapp.BeginBlock(wrsp.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, wrsp.Header{})
	stakeHandler := stake.NewHandler(sk)

	dummyDescription := stake.NewDescription("T", "E", "S", "T")
	val1CreateMsg := stake.NewMsgCreateValidator(addrs[0], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 5), dummyDescription)
	stakeHandler(ctx, val1CreateMsg)
	val2CreateMsg := stake.NewMsgCreateValidator(addrs[1], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 6), dummyDescription)
	stakeHandler(ctx, val2CreateMsg)
	val3CreateMsg := stake.NewMsgCreateValidator(addrs[2], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 7), dummyDescription)
	stakeHandler(ctx, val3CreateMsg)

	delegator1Msg := stake.NewMsgDelegate(addrs[3], addrs[2], sdk.NewCoin("steak", 10))
	stakeHandler(ctx, delegator1Msg)
	delegator1Msg2 := stake.NewMsgDelegate(addrs[3], addrs[1], sdk.NewCoin("steak", 10))
	stakeHandler(ctx, delegator1Msg2)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[3], OptionNo)
	require.Nil(t, err)

	passes, _ := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.False(t, passes)
}

func TestTallyDelgatorMultipleInherit(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10)
	mapp.BeginBlock(wrsp.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, wrsp.Header{})
	stakeHandler := stake.NewHandler(sk)

	dummyDescription := stake.NewDescription("T", "E", "S", "T")
	val1CreateMsg := stake.NewMsgCreateValidator(addrs[0], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 25), dummyDescription)
	stakeHandler(ctx, val1CreateMsg)
	val2CreateMsg := stake.NewMsgCreateValidator(addrs[1], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 6), dummyDescription)
	stakeHandler(ctx, val2CreateMsg)
	val3CreateMsg := stake.NewMsgCreateValidator(addrs[2], crypto.GenPrivKeyEd25519().PubKey(), sdk.NewCoin("steak", 7), dummyDescription)
	stakeHandler(ctx, val3CreateMsg)

	delegator1Msg := stake.NewMsgDelegate(addrs[3], addrs[2], sdk.NewCoin("steak", 10))
	stakeHandler(ctx, delegator1Msg)
	delegator1Msg2 := stake.NewMsgDelegate(addrs[3], addrs[1], sdk.NewCoin("steak", 10))
	stakeHandler(ctx, delegator1Msg2)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionNo)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionNo)
	require.Nil(t, err)

	passes, _ := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.False(t, passes)
}

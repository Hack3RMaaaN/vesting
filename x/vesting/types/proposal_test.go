package types_test

import (
	"github.com/evmos/vesting/x/vesting/types"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ProposalTestSuite struct {
	suite.Suite
}

func TestProposalTestSuite(t *testing.T) {
	suite.Run(t, new(ProposalTestSuite))
}

func (suite *ProposalTestSuite) TestKeysTypes() {
	suite.Require().Equal("clawbackvesting", (&types.ClawbackProposal{}).ProposalRoute())
	suite.Require().Equal("Clawback", (&types.ClawbackProposal{}).ProposalType())
}

func (suite *ProposalTestSuite) TestClawbackProposal() {
	testCases := []struct {
		msg                string
		title              string
		description        string
		address            string
		destinationAddress string
		expectPass         bool
	}{
		// Valid tests
		{
			msg:         "Clawback proposal - valid address",
			title:       "test",
			description: "test desc",
			address:     "cosmos1p3ucd3ptpw902fluyjzhq3ffgq4ntddac9sa3s",
			expectPass:  true,
		},
		// Invalid - Missing params
		{
			msg:         "Clawback proposal - invalid missing title ",
			title:       "",
			description: "test desc",
			address:     "cosmos1p3ucd3ptpw902fluyjzhq3ffgq4ntddac9sa3s",
			expectPass:  false,
		},
		{
			msg:         "Clawback proposal - invalid missing description ",
			title:       "test",
			description: "",
			address:     "cosmos1p3ucd3ptpw902fluyjzhq3ffgq4ntddac9sa3s",
			expectPass:  false,
		},
		// Invalid address
		{
			msg:         "Clawback proposal - invalid address",
			title:       "test",
			description: "test desc",
			address:     "cosmos1p3ucd3ptpw902fluyjzhq3ffgq4ntddac9sass",
			expectPass:  false,
		},
		{
			msg:                "Clawback proposal - invalid destination addr",
			title:              "test",
			description:        "test desc",
			address:            "cosmos1p3ucd3ptpw902fluyjzhq3ffgq4ntddac9sa3s",
			destinationAddress: "125182ujaisch8hsgs",
			expectPass:         false,
		},
	}

	for i, tc := range testCases {
		tx := types.NewClawbackProposal(tc.title, tc.description, tc.address, tc.destinationAddress)
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
		}
	}
}

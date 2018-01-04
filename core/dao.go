
package core

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

// ValidateDAOHeaderExtraData validates the extra-data field of a block header to
// ensure it conforms to DAO hard-fork rules.
//
// DAO hard-fork extension to the header validity:
//   a) if the node is no-fork, do not accept blocks in the [fork, fork+10) range
//      with the fork specific extra-data set
//   b) if the node is pro-fork, require blocks in the specific range to have the
//      unique extra-data set.
func ValidateDAOHeaderExtraData(config *params.ChainConfig, header *types.Header) error {
	// Short circuit validation if the node doesn't care about the DAO fork
	if config.DAOForkBlock == nil {
		return nil
	}
	// Make sure the block is within the fork's modified extra-data range
	limit := new(big.Int).Add(config.DAOForkBlock, params.DAOForkExtraRange)
	if header.Number.Cmp(config.DAOForkBlock) < 0 || header.Number.Cmp(limit) >= 0 {
		return nil
	}
	// Depending whether we support or oppose the fork, validate the extra-data contents
	if config.DAOForkSupport {
		if !bytes.Equal(header.Extra, params.DAOForkBlockExtra) {
			return fmt.Errorf("DAO pro-fork bad block extra-data: 0x%x", header.Extra)
		}
	} else {
		if bytes.Equal(header.Extra, params.DAOForkBlockExtra) {
			return fmt.Errorf("DAO no-fork bad block extra-data: 0x%x", header.Extra)
		}
	}
	// All ok, header has the same extra-data we expect
	return nil
}

// ApplyDAOHardFork modifies the state database according to the DAO hard-fork
// rules, transferring all balances of a set of DAO accounts to a single refund
// contract.
func ApplyDAOHardFork(statedb *state.StateDB) {
	// Retrieve the contract to refund balances into
	if !statedb.Exist(params.DAORefundContract) {
		statedb.CreateAccount(params.DAORefundContract)
	}

	// Move every DAO account and extra-balance account funds into the refund contract
	for _, addr := range params.DAODrainList() {
		statedb.AddBalance(params.DAORefundContract, statedb.GetBalance(addr))
		statedb.SetBalance(addr, "0.00")	// new(big.Int))	Water Cherry
	}
}
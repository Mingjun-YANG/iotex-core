// Copyright (c) 2018 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package blockchain

import (
	"context"
	"math/big"
	"sync"

	"github.com/facebookgo/clock"
	"github.com/pkg/errors"

	"github.com/iotexproject/iotex-core/blockchain/action"
	"github.com/iotexproject/iotex-core/config"
	"github.com/iotexproject/iotex-core/crypto"
	"github.com/iotexproject/iotex-core/db"
	"github.com/iotexproject/iotex-core/iotxaddress"
	"github.com/iotexproject/iotex-core/logger"
	"github.com/iotexproject/iotex-core/pkg/hash"
	"github.com/iotexproject/iotex-core/pkg/keypair"
	"github.com/iotexproject/iotex-core/pkg/lifecycle"
	"github.com/iotexproject/iotex-core/state"
)

// Blockchain represents the blockchain data structure and hosts the APIs to access it
type Blockchain interface {
	lifecycle.StartStopper

	// Balance returns balance of an account
	Balance(addr string) (*big.Int, error)
	// Nonce returns the nonce if the account exists
	Nonce(addr string) (uint64, error)
	// CreateState adds a new State with initial balance to the factory
	CreateState(addr string, init uint64) (*state.State, error)
	// CommitStateChanges updates a State from the given actions
	CommitStateChanges(chainHeight uint64, tsf []*action.Transfer, vote []*action.Vote, executions []*action.Execution) error
	// Candidates returns the candidate list
	Candidates() (uint64, []*state.Candidate)
	// CandidatesByHeight returns the candidate list by a given height
	CandidatesByHeight(height uint64) ([]*state.Candidate, error)
	// For exposing blockchain states
	// GetHeightByHash returns Block's height by hash
	GetHeightByHash(h hash.Hash32B) (uint64, error)
	// GetHashByHeight returns Block's hash by height
	GetHashByHeight(height uint64) (hash.Hash32B, error)
	// GetBlockByHeight returns Block by height
	GetBlockByHeight(height uint64) (*Block, error)
	// GetBlockByHash returns Block by hash
	GetBlockByHash(h hash.Hash32B) (*Block, error)
	// GetTotalTransfers returns the total number of transfers
	GetTotalTransfers() (uint64, error)
	// GetTotalVotes returns the total number of votes
	GetTotalVotes() (uint64, error)
	// GetTotalExecutions returns the total number of executions
	GetTotalExecutions() (uint64, error)
	// GetTransfersFromAddress returns transaction from address
	GetTransfersFromAddress(address string) ([]hash.Hash32B, error)
	// GetTransfersToAddress returns transaction to address
	GetTransfersToAddress(address string) ([]hash.Hash32B, error)
	// GetTransfersByTransferHash returns transfer by transfer hash
	GetTransferByTransferHash(h hash.Hash32B) (*action.Transfer, error)
	// GetBlockHashByTransferHash returns Block hash by transfer hash
	GetBlockHashByTransferHash(h hash.Hash32B) (hash.Hash32B, error)
	// GetVoteFromAddress returns vote from address
	GetVotesFromAddress(address string) ([]hash.Hash32B, error)
	// GetVoteToAddress returns vote to address
	GetVotesToAddress(address string) ([]hash.Hash32B, error)
	// GetVotesByVoteHash returns vote by vote hash
	GetVoteByVoteHash(h hash.Hash32B) (*action.Vote, error)
	// GetBlockHashByVoteHash returns Block hash by vote hash
	GetBlockHashByVoteHash(h hash.Hash32B) (hash.Hash32B, error)
	// GetExecutionsFromAddress returns executions from address
	GetExecutionsFromAddress(address string) ([]hash.Hash32B, error)
	// GetExecutionsToAddress returns executions to address
	GetExecutionsToAddress(address string) ([]hash.Hash32B, error)
	// GetExecutionByExecutionHash returns execution by execution hash
	GetExecutionByExecutionHash(h hash.Hash32B) (*action.Execution, error)
	// GetBlockHashByExecutionHash returns Block hash by execution hash
	GetBlockHashByExecutionHash(h hash.Hash32B) (hash.Hash32B, error)
	// GetReceiptByExecutionHash returns the receipt by execution hash
	GetReceiptByExecutionHash(h hash.Hash32B) (*Receipt, error)
	// GetFactory returns the State Factory
	GetFactory() state.Factory
	// TipHash returns tip block's hash
	TipHash() (hash.Hash32B, error)
	// TipHeight returns tip block's height
	TipHeight() (uint64, error)
	// StateByAddr returns state of a given address
	StateByAddr(address string) (*state.State, error)

	// For block operations
	// MintNewBlock creates a new block with given actions
	// Note: the coinbase transfer will be added to the given transfers when minting a new block
	MintNewBlock(tsf []*action.Transfer, vote []*action.Vote, executions []*action.Execution, address *iotxaddress.Address, data string) (*Block, error)
	// TODO: Merge the MintNewDKGBlock into MintNewBlock
	// MintNewDKGBlock creates a new block with given actions and dkg keys
	MintNewDKGBlock(tsf []*action.Transfer, vote []*action.Vote, executions []*action.Execution,
		producer *iotxaddress.Address, dkgAddress *iotxaddress.DKGAddress, seed []byte, data string) (*Block, error)
	// MintDummyNewBlock creates a new dummy block, used for unreached consensus
	MintNewDummyBlock() *Block
	// CommitBlock validates and appends a block to the chain
	CommitBlock(blk *Block) error
	// ValidateBlock validates a new block before adding it to the blockchain
	ValidateBlock(blk *Block) error

	// For action operations
	// Validator returns the current validator object
	Validator() Validator
	// SetValidator sets the current validator object
	SetValidator(val Validator)

	// For smart contract operations
	// ExecuteContractRead runs a read-only smart contract operation, this is done off the network since it does not
	// cause any state change
	ExecuteContractRead(*action.Execution) ([]byte, error)
}

// blockchain implements the Blockchain interface
type blockchain struct {
	mu        sync.RWMutex // mutex to protect utk, tipHeight and tipHash
	dao       *blockDAO
	config    *config.Config
	genesis   *Genesis
	chainID   uint32
	tipHeight uint64
	tipHash   hash.Hash32B
	validator Validator
	lifecycle lifecycle.Lifecycle
	clk       clock.Clock

	// used by account-based model
	sf state.Factory
}

// Option sets blockchain construction parameter
type Option func(*blockchain, *config.Config) error

// DefaultStateFactoryOption sets blockchain's sf from config
func DefaultStateFactoryOption() Option {
	return func(bc *blockchain, cfg *config.Config) error {
		sf, err := state.NewFactory(cfg, state.DefaultTrieOption())
		if err != nil {
			return errors.Wrapf(err, "Failed to create state factory")
		}
		bc.sf = sf

		return nil
	}
}

// PrecreatedStateFactoryOption sets blockchain's state.Factory to sf
func PrecreatedStateFactoryOption(sf state.Factory) Option {
	return func(bc *blockchain, conf *config.Config) error {
		bc.sf = sf

		return nil
	}
}

// InMemStateFactoryOption sets blockchain's state.Factory as in memory sf
func InMemStateFactoryOption() Option {
	return func(bc *blockchain, cfg *config.Config) error {
		sf, err := state.NewFactory(cfg, state.InMemTrieOption())
		if err != nil {
			return errors.Wrapf(err, "Failed to create state factory")
		}
		bc.sf = sf

		return nil
	}
}

// PrecreatedDaoOption sets blockchain's dao
func PrecreatedDaoOption(dao *blockDAO) Option {
	return func(bc *blockchain, conf *config.Config) error {
		bc.dao = dao

		return nil
	}
}

// BoltDBDaoOption sets blockchain's dao with BoltDB from config.Chain.ChainDBPath
func BoltDBDaoOption() Option {
	return func(bc *blockchain, cfg *config.Config) error {
		bc.dao = newBlockDAO(db.NewBoltDB(cfg.Chain.ChainDBPath, &cfg.DB))

		return nil
	}
}

// InMemDaoOption sets blockchain's dao with MemKVStore
func InMemDaoOption() Option {
	return func(bc *blockchain, cfg *config.Config) error {
		bc.dao = newBlockDAO(db.NewMemKVStore())

		return nil
	}
}

// ClockOption overrides the default clock
func ClockOption(clk clock.Clock) Option {
	return func(bc *blockchain, conf *config.Config) error {
		bc.clk = clk

		return nil
	}
}

// NewBlockchain creates a new blockchain and DB instance
func NewBlockchain(cfg *config.Config, opts ...Option) Blockchain {
	// create the Blockchain
	chain := &blockchain{
		config:  cfg,
		genesis: Gen,
		clk:     clock.New(),
	}
	for _, opt := range opts {
		if err := opt(chain, cfg); err != nil {
			logger.Error().Err(err).Msgf("Failed to create blockchain option %s", opt)
			return nil
		}
	}
	chain.initValidator()
	if chain.dao != nil {
		chain.lifecycle.Add(chain.dao)
	}
	if chain.sf != nil {
		chain.lifecycle.Add(chain.sf)
	}
	if err := chain.Start(context.Background()); err != nil {
		logger.Error().Err(err).Msg("Failed to start blockchain")
		return nil
	}

	height, err := chain.TipHeight()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get blockchain height")
		return nil
	}
	if height > 0 {
		factoryHeight, err := chain.sf.Height()
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get factory's height")
			return nil
		}
		logger.Info().
			Uint64("blockchain height", height).Uint64("factory height", factoryHeight).
			Msg("Restarting blockchain")
		return chain
	}
	genesis := NewGenesisBlock(cfg)
	if genesis == nil {
		logger.Error().Msg("Cannot create genesis block.")
		return nil
	}
	// Genesis block has height 0
	if genesis.Header.height != 0 {
		logger.Error().
			Uint64("Genesis block has height", genesis.Height()).
			Msg("Expecting 0")
		return nil
	}
	// add producer into Trie
	if chain.sf != nil {
		if _, err := chain.sf.LoadOrCreateState(Gen.CreatorAddr, Gen.TotalSupply); err != nil {
			logger.Error().Err(err).Msg("Failed to add Creator into StateFactory")
			return nil
		}
	}
	// add Genesis block as very first block
	if err := chain.CommitBlock(genesis); err != nil {
		logger.Error().Err(err).Msg("Failed to commit Genesis block")
		return nil
	}
	return chain
}

func (bc *blockchain) initValidator() { bc.validator = &validator{sf: bc.sf} }

// Start starts the blockchain
func (bc *blockchain) Start(ctx context.Context) (err error) {
	if err = bc.lifecycle.OnStart(ctx); err != nil {
		return err
	}
	// get blockchain tip height
	bc.mu.Lock()
	defer bc.mu.Unlock()
	if bc.tipHeight, err = bc.dao.getBlockchainHeight(); err != nil {
		return err
	}
	if bc.tipHeight == 0 {
		return nil
	}
	// get blockchain tip hash
	if bc.tipHash, err = bc.dao.getBlockHash(bc.tipHeight); err != nil {
		return err
	}
	// populate state factory
	if bc.sf == nil {
		return errors.New("statefactory cannot be nil")
	}
	var startHeight uint64
	if factoryHeight, err := bc.sf.Height(); err == nil {
		if factoryHeight > bc.tipHeight {
			return errors.New("factory is higher than blockchain")
		}
		startHeight = factoryHeight + 1
	}
	// If restarting factory from fresh db, first create creator's state
	if startHeight == 0 {
		if _, err := bc.sf.LoadOrCreateState(Gen.CreatorAddr, Gen.TotalSupply); err != nil {
			return err
		}
	}
	for i := startHeight; i <= bc.tipHeight; i++ {
		blk, err := bc.GetBlockByHeight(i)
		if err != nil {
			return err
		}
		if err := bc.sf.CommitStateChanges(blk.Height(), blk.Transfers, blk.Votes, blk.Executions); err != nil {
			return err
		}
	}
	return nil
}

// Stop stops the blockchain.
func (bc *blockchain) Stop(ctx context.Context) error { return bc.lifecycle.OnStop(ctx) }

// Balance returns balance of address
func (bc *blockchain) Balance(addr string) (*big.Int, error) {
	return bc.sf.Balance(addr)
}

// Nonce returns the nonce if the account exists
func (bc *blockchain) Nonce(addr string) (uint64, error) {
	return bc.sf.Nonce(addr)
}

// CreateState adds a new State with initial balance to the factory
func (bc *blockchain) CreateState(addr string, init uint64) (*state.State, error) {
	return bc.sf.LoadOrCreateState(addr, init)
}

// CommitStateChanges updates a State from the given actions
func (bc *blockchain) CommitStateChanges(blockHeight uint64, tsf []*action.Transfer, vote []*action.Vote, executions []*action.Execution) error {
	return bc.sf.CommitStateChanges(blockHeight, tsf, vote, executions)
}

// Candidates returns the candidate list
func (bc *blockchain) Candidates() (uint64, []*state.Candidate) {
	return bc.sf.Candidates()
}

// CandidatesByHeight returns the candidate list by a given height
func (bc *blockchain) CandidatesByHeight(height uint64) ([]*state.Candidate, error) {
	return bc.sf.CandidatesByHeight(height)
}

// GetHeightByHash returns block's height by hash
func (bc *blockchain) GetHeightByHash(h hash.Hash32B) (uint64, error) {
	return bc.dao.getBlockHeight(h)
}

// GetHashByHeight returns block's hash by height
func (bc *blockchain) GetHashByHeight(height uint64) (hash.Hash32B, error) {
	return bc.dao.getBlockHash(height)
}

// GetBlockByHeight returns block from the blockchain hash by height
func (bc *blockchain) GetBlockByHeight(height uint64) (*Block, error) {
	hash, err := bc.GetHashByHeight(height)
	if err != nil {
		return nil, err
	}
	return bc.GetBlockByHash(hash)
}

// GetBlockByHash returns block from the blockchain hash by hash
func (bc *blockchain) GetBlockByHash(h hash.Hash32B) (*Block, error) {
	return bc.dao.getBlock(h)
}

// GetTotalTransfers returns the total number of transfers
func (bc *blockchain) GetTotalTransfers() (uint64, error) {
	return bc.dao.getTotalTransfers()
}

// GetTotalVotes returns the total number of votes
func (bc *blockchain) GetTotalVotes() (uint64, error) {
	return bc.dao.getTotalVotes()
}

// GetTotalExecutions returns the total number of executions
func (bc *blockchain) GetTotalExecutions() (uint64, error) {
	return bc.dao.getTotalExecutions()
}

// GetTransfersFromAddress returns transfers from address
func (bc *blockchain) GetTransfersFromAddress(address string) ([]hash.Hash32B, error) {
	return bc.dao.getTransfersBySenderAddress(address)
}

// GetTransfersToAddress returns transfers to address
func (bc *blockchain) GetTransfersToAddress(address string) ([]hash.Hash32B, error) {
	return bc.dao.getTransfersByRecipientAddress(address)
}

// GetTransferByTransferHash returns transfer by transfer hash
func (bc *blockchain) GetTransferByTransferHash(h hash.Hash32B) (*action.Transfer, error) {
	blkHash, err := bc.dao.getBlockHashByTransferHash(h)
	if err != nil {
		return nil, err
	}
	blk, err := bc.dao.getBlock(blkHash)
	if err != nil {
		return nil, err
	}
	for _, transfer := range blk.Transfers {
		if transfer.Hash() == h {
			return transfer, nil
		}
	}
	return nil, errors.Errorf("block %x does not have transfer %x", blkHash, h)
}

// GetBlockHashByTxHash returns Block hash by transfer hash
func (bc *blockchain) GetBlockHashByTransferHash(h hash.Hash32B) (hash.Hash32B, error) {
	return bc.dao.getBlockHashByTransferHash(h)
}

// GetVoteFromAddress returns votes from address
func (bc *blockchain) GetVotesFromAddress(address string) ([]hash.Hash32B, error) {
	return bc.dao.getVotesBySenderAddress(address)
}

// GetVoteToAddress returns votes to address
func (bc *blockchain) GetVotesToAddress(address string) ([]hash.Hash32B, error) {
	return bc.dao.getVotesByRecipientAddress(address)
}

// GetVotesByVoteHash returns vote by vote hash
func (bc *blockchain) GetVoteByVoteHash(h hash.Hash32B) (*action.Vote, error) {
	blkHash, err := bc.dao.getBlockHashByVoteHash(h)
	if err != nil {
		return nil, err
	}
	blk, err := bc.dao.getBlock(blkHash)
	if err != nil {
		return nil, err
	}
	for _, vote := range blk.Votes {
		if vote.Hash() == h {
			return vote, nil
		}
	}
	return nil, errors.Errorf("block %x does not have vote %x", blkHash, h)
}

// GetBlockHashByVoteHash returns Block hash by vote hash
func (bc *blockchain) GetBlockHashByVoteHash(h hash.Hash32B) (hash.Hash32B, error) {
	return bc.dao.getBlockHashByVoteHash(h)
}

// GetExecutionsFromAddress returns executions from address
func (bc *blockchain) GetExecutionsFromAddress(address string) ([]hash.Hash32B, error) {
	return bc.dao.getExecutionsByExecutorAddress(address)
}

// GetExecutionsToAddress returns executions to address
func (bc *blockchain) GetExecutionsToAddress(address string) ([]hash.Hash32B, error) {
	return bc.dao.getExecutionsByContractAddress(address)
}

// GetExecutionByExecutionHash returns execution by execution hash
func (bc *blockchain) GetExecutionByExecutionHash(h hash.Hash32B) (*action.Execution, error) {
	blkHash, err := bc.dao.getBlockHashByExecutionHash(h)
	if err != nil {
		return nil, err
	}
	blk, err := bc.dao.getBlock(blkHash)
	if err != nil {
		return nil, err
	}
	for _, execution := range blk.Executions {
		if execution.Hash() == h {
			return execution, nil
		}
	}
	return nil, errors.Errorf("block %x does not have execution %x", blkHash, h)
}

// GetBlockHashByExecutionHash returns Block hash by execution hash
func (bc *blockchain) GetBlockHashByExecutionHash(h hash.Hash32B) (hash.Hash32B, error) {
	return bc.dao.getBlockHashByExecutionHash(h)
}

// GetReceiptByExecutionHash returns the receipt by execution hash
func (bc *blockchain) GetReceiptByExecutionHash(h hash.Hash32B) (*Receipt, error) {
	return bc.dao.getReceiptByExecutionHash(h)
}

// GetFactory returns the State Factory
func (bc *blockchain) GetFactory() state.Factory {
	return bc.sf
}

// TipHash returns tip block's hash
func (bc *blockchain) TipHash() (hash.Hash32B, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.tipHash, nil
}

// TipHeight returns tip block's height
func (bc *blockchain) TipHeight() (uint64, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.tipHeight, nil
}

// ValidateBlock validates a new block before adding it to the blockchain
func (bc *blockchain) ValidateBlock(blk *Block) error {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if bc.validator == nil {
		panic("no block validator")
	}

	// replacement logic, used to replace a fake old dummy block
	if blk.Height() != 0 && blk.Height() <= bc.tipHeight {
		oldDummyBlock, err := bc.GetBlockByHeight(blk.Height())
		if err != nil {
			return errors.Wrapf(err, "The height of the new block is invalid")
		}
		if !oldDummyBlock.IsDummyBlock() {
			return errors.New("The replaced block is not a dummy block")
		}
		lastBlock, err := bc.GetBlockByHeight(blk.Height() - 1)
		if err != nil {
			return errors.Wrapf(err, "Failed to get the last block when replacing the dummy block")
		}
		return bc.validator.Validate(blk, lastBlock.Height(), lastBlock.HashBlock())
	}

	return bc.validator.Validate(blk, bc.tipHeight, bc.tipHash)
}

// MintNewBlock creates a new block with given actions
// Note: the coinbase transfer will be added to the given transfers
// when minting a new block
func (bc *blockchain) MintNewBlock(tsf []*action.Transfer, vote []*action.Vote, executions []*action.Execution,
	producer *iotxaddress.Address, data string) (*Block, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	tsf = append(tsf, action.NewCoinBaseTransfer(big.NewInt(int64(bc.genesis.BlockReward)), producer.RawAddress))

	blk := NewBlock(bc.chainID, bc.tipHeight+1, bc.tipHash, bc.clk, tsf, vote, executions)
	if producer.PrivateKey == keypair.ZeroPrivateKey {
		logger.Warn().Msg("Unsigned block...")
		return blk, nil
	}

	blk.Header.DKGID = []byte{}
	blk.Header.DKGPubkey = []byte{}
	blk.Header.DKGBlockSig = []byte{}

	blk.Header.Pubkey = producer.PublicKey
	blkHash := blk.HashBlock()
	blk.Header.blockSig = crypto.EC283.Sign(producer.PrivateKey, blkHash[:])
	return blk, nil
}

// MintNewDKGBlock creates a new block with given actions and dkg keys
// Note: the coinbase transfer will be added to the given transfers
// when minting a new block
func (bc *blockchain) MintNewDKGBlock(tsf []*action.Transfer, vote []*action.Vote, executions []*action.Execution,
	producer *iotxaddress.Address, dkgAddress *iotxaddress.DKGAddress, seed []byte, data string) (*Block, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	tsf = append(tsf, action.NewCoinBaseTransfer(big.NewInt(int64(bc.genesis.BlockReward)), producer.RawAddress))

	blk := NewBlock(bc.chainID, bc.tipHeight+1, bc.tipHash, bc.clk, tsf, vote, executions)
	if producer.PrivateKey == keypair.ZeroPrivateKey {
		logger.Warn().Msg("Unsigned block...")
		return blk, nil
	}

	blk.Header.DKGID = []byte{}
	blk.Header.DKGPubkey = []byte{}
	blk.Header.DKGBlockSig = []byte{}
	if len(dkgAddress.PublicKey) > 0 && len(dkgAddress.PrivateKey) > 0 && len(dkgAddress.ID) > 0 {
		blk.Header.DKGID = dkgAddress.ID
		blk.Header.DKGPubkey = dkgAddress.PublicKey
		var err error
		if _, blk.Header.DKGBlockSig, err = crypto.BLS.SignShare(dkgAddress.PrivateKey, seed); err != nil {
			return nil, errors.Wrap(err, "Failed to do DKG sign")
		}
	}

	blk.Header.Pubkey = producer.PublicKey
	blkHash := blk.HashBlock()
	blk.Header.blockSig = crypto.EC283.Sign(producer.PrivateKey, blkHash[:])
	return blk, nil
}

// MintDummyNewBlock creates a new dummy block, used for unreached consensus
func (bc *blockchain) MintNewDummyBlock() *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	blk := NewBlock(bc.chainID, bc.tipHeight+1, bc.tipHash, bc.clk, nil, nil, nil)
	blk.Header.Pubkey = keypair.ZeroPublicKey
	blk.Header.blockSig = []byte{}
	return blk
}

//  CommitBlock validates and appends a block to the chain
func (bc *blockchain) CommitBlock(blk *Block) error {
	if err := bc.ValidateBlock(blk); err != nil {
		return err
	}
	return bc.commitBlock(blk)
}

// StateByAddr returns the state of an address
func (bc *blockchain) StateByAddr(address string) (*state.State, error) {
	if bc.sf != nil {
		s, err := bc.sf.State(address)
		if err != nil {
			logger.Warn().Err(err).Str("Address", address)
			return nil, errors.New("account does not exist")
		}

		return s, nil
	}

	return nil, errors.New("state factory is nil")
}

// SetValidator sets the current validator object
func (bc *blockchain) SetValidator(val Validator) {
	bc.validator = val
}

// Validator gets the current validator object
func (bc *blockchain) Validator() Validator {
	return bc.validator
}

// ExecuteContractRead runs a read-only smart contract operation, this is done off the network since it does not
// cause any state change
func (bc *blockchain) ExecuteContractRead(ex *action.Execution) ([]byte, error) {
	// use latest block as carrier to run the offline execution
	// the block itself is not used
	h, _ := bc.TipHeight()
	blk, err := bc.GetBlockByHeight(h)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get block in ExecuteContractRead")
	}
	blk.Executions = nil
	blk.Executions = []*action.Execution{ex}
	blk.receipts = nil
	ExecuteContracts(blk, bc)
	// pull the results from receipt
	exHash := ex.Hash()
	receipt, ok := blk.receipts[exHash]
	if !ok {
		return nil, errors.Wrap(err, "failed to get receipt in ExecuteContractRead")
	}
	return receipt.ReturnValue, nil
}

//======================================
// private functions
//=====================================
// commitBlock commits a block to the chain
func (bc *blockchain) commitBlock(blk *Block) error {
	// write block into DB
	if err := bc.dao.putBlock(blk); err != nil {
		return err
	}
	// update tip hash and height
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.tipHeight = blk.Header.height
	bc.tipHash = blk.HashBlock()
	// update state factory
	if bc.sf != nil {
		ExecuteContracts(blk, bc)
		if err := bc.sf.CommitStateChanges(blk.Height(), blk.Transfers, blk.Votes, blk.Executions); err != nil {
			return err
		}
	}
	// write smart contract receipt into DB
	if err := bc.dao.putReceipts(blk); err != nil {
		return err
	}
	logger.Info().Uint64("height", blk.Header.height).Msg("commit a block")
	return nil
}

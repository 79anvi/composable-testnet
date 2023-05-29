package mint

import (
	"context"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/notional-labs/banksy/v2/x/mint/client/cli"
	"github.com/notional-labs/banksy/v2/x/mint/keeper"
	"github.com/notional-labs/banksy/v2/x/mint/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
	// TODO: simulation features
	// _ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic defines the basic application module used by the mint module.
type AppModuleBasic struct {
	cdc codec.Codec
}

var _ module.AppModuleBasic = AppModuleBasic{}

// Name returns the mint module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the mint module's types on the given LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(_ cdctypes.InterfaceRegistry) {}

// DefaultGenesis returns default genesis state as raw bytes for the mint
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the mint module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return types.ValidateGenesis(data)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the mint module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns no root tx command for the mint module.
func (AppModuleBasic) GetTxCmd() *cobra.Command { return nil }

// GetQueryCmd returns the root query command for the mint module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// AppModule implements an application module for the mint module.
type AppModule struct {
	AppModuleBasic

	keeper     keeper.Keeper
	authKeeper types.AccountKeeper

	// inflationCalculator is used to calculate the inflation rate during BeginBlock.
	// If inflationCalculator is nil, the default inflation calculation logic is used.
	inflationCalculator types.InflationCalculationFn
}

// NewAppModule creates a new AppModule object. If the InflationCalculationFn
// argument is nil, then the SDK's default inflation function will be used.
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper, ak types.AccountKeeper, ic types.InflationCalculationFn) AppModule {
	if ic == nil {
		ic = types.DefaultInflationCalculationFn
	}
	return AppModule{
		AppModuleBasic:      AppModuleBasic{cdc: cdc},
		keeper:              keeper,
		authKeeper:          ak,
		inflationCalculator: ic,
	}
}

// Name returns the mint module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants registers the mint module invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// RegisterServices registers a gRPC query service to respond to the
// module-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

// InitGenesis performs genesis initialization for the mint module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)

	am.keeper.InitGenesis(ctx, am.authKeeper, &genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the mint
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx, am.authKeeper)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock returns the begin blocker for the mint module.
func (am AppModule) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
	BeginBlocker(ctx, am.keeper, am.inflationCalculator)
}

// AppModuleSimulation functions
// TODO : add simulation fot this module
// // GenerateGenesisState creates a randomized GenState of the mint module.
// func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
// 	simulation.RandomizedGenState(simState)
// }

// // ProposalContents doesn't return any content functions for governance proposals.
// func (AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent {
// 	return nil
// }

// // RegisterStoreDecoder registers a decoder for mint module's types.
// func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
// 	sdr[types.StoreKey] = simulation.NewDecodeStore(am.cdc)
// }

// // WeightedOperations doesn't return any mint module operation.
// func (AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
// 	return nil
// }
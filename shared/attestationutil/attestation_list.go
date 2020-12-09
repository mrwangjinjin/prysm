package attestationutil

import (
	"context"
	"sort"

	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/blocks"
	stateTrie "github.com/prysmaticlabs/prysm/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/shared/params"
)

type List []*ethpb.Attestation

// SplitForProposer separates attestation list into two groups: valid and invalid attestations.
// The first group passes the all the required checks for attestation to be considered for proposing.
// And attestations from the second group should be deleted.
func (al List) SplitForProposer(ctx context.Context, state *stateTrie.BeaconState) (List, List) {
	return al.SplitValidate(func(att *ethpb.Attestation) bool {
		if _, err := blocks.ProcessAttestation(ctx, state, att); err == nil {
			return true
		}
		return false
	})
}

// SplitForProposal separates attestation list into two groups: valid and invalid attestations.
// Item validity is assessed using provided validation function.
func (al List) SplitValidate(validate func(att *ethpb.Attestation) bool) (List, List) {
	validAtts := make([]*ethpb.Attestation, 0, len(al))
	invalidAtts := make([]*ethpb.Attestation, 0, len(al))
	for _, att := range al {
		if validate(att) {
			validAtts = append(validAtts, att)
			continue
		}
		invalidAtts = append(invalidAtts, att)
	}
	return validAtts, invalidAtts
}

// SortByProfitability orders attestations by highest slot and by highest aggregation bit count.
func (al List) SortByProfitability() List {
	if len(al) < 2 {
		return al
	}
	sort.Slice(al, func(i, j int) bool {
		if al[i].Data.Slot == al[j].Data.Slot {
			return al[i].AggregationBits.Count() > al[j].AggregationBits.Count()
		}
		return al[i].Data.Slot > al[j].Data.Slot
	})
	return al
}

// LimitToMaxAttestations limits attestations to maximum attestations per block.
func (al List) LimitToMaxAttestations() List {
	if uint64(len(al)) > params.BeaconConfig().MaxAttestations {
		return al[:params.BeaconConfig().MaxAttestations]
	}
	return al
}

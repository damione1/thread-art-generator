package service

import (
	"testing"

	pbErrors "github.com/Damione1/thread-art-generator/core/errors"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/stretchr/testify/require"
)

func TestValidateCompositionParamsCrossField(t *testing.T) {
	t.Parallel()
	require.NoError(t, validateCompositionParams(&pb.Composition{
		NailsQuantity:     280,
		StartingNail:      0,
		MinimumDifference: 22,
	}))

	err := validateCompositionParams(&pb.Composition{
		NailsQuantity:     10,
		StartingNail:      10,
		MinimumDifference: 8,
	})
	require.Error(t, err)
	require.True(t, pbErrors.HasFieldViolation(err, "composition.starting_nail"), "got %v", pbErrors.ExtractFieldViolations(err))
	require.True(t, pbErrors.HasFieldViolation(err, "composition.minimum_difference"), "got %v", pbErrors.ExtractFieldViolations(err))
}

func TestNormalizeCompositionAlgorithmUsesRegistry(t *testing.T) {
	t.Parallel()
	require.Equal(t, pb.CompositionAlgorithm_COMPOSITION_ALGORITHM_VRELLIS, normalizeCompositionAlgorithm(0))
	require.Equal(t, pb.CompositionAlgorithm_COMPOSITION_ALGORITHM_L2, normalizeCompositionAlgorithm(pb.CompositionAlgorithm_COMPOSITION_ALGORITHM_L2))
}

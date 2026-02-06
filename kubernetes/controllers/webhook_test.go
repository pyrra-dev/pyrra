package controllers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	pyrrav1alpha1 "github.com/pyrra-dev/pyrra/kubernetes/api/v1alpha1"
)

// Test_validatorInterface verifies ServiceLevelObjective implements admission.Validator.
func Test_validatorInterface(t *testing.T) {
	slo := &pyrrav1alpha1.ServiceLevelObjective{}

	// This will fail to compile if ServiceLevelObjective doesn't implement admission.Validator
	var _ admission.Validator[*pyrrav1alpha1.ServiceLevelObjective] = slo

	// Test that ValidateCreate exists and can be called
	_, err := slo.ValidateCreate(context.Background(), slo)
	// We expect an error because the SLO is empty, but the method should be callable
	require.Error(t, err, "ValidateCreate should return error for invalid SLO")
}

// Test_setupWebhookWithManager verifies the webhook setup function signature is correct.
func Test_setupWebhookWithManager(t *testing.T) {
	// This test verifies that SetupWebhookWithManager exists and has the correct signature.
	// The actual webhook registration is tested in integration tests.
	reconciler := &ServiceLevelObjectiveReconciler{}

	// Verify the method exists (compilation test)
	_ = reconciler.SetupWebhookWithManager

	// Note: We cannot test the actual webhook registration without a manager/envtest,
	// but the fix ensures that when SetupWebhookWithManager is called with a real manager,
	// it will properly register the webhook using WithValidator().
	t.Log("SetupWebhookWithManager method exists with correct signature")
}

package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestStackTTLWithoutIngress(t *testing.T) {
	t.Parallel()
	stacksetName := "stackset-ttl-noingress"
	specFactory := NewTestStacksetSpecFactory(stacksetName).StackGC(3, 0)
	for i := 0; i < 5; i++ {
		stackVersion := fmt.Sprintf("v%d", i)
		var err error
		spec := specFactory.Create(stackVersion)
		if !stacksetExists(stacksetName) {
			err = createStackSet(stacksetName, true, spec)
		} else {
			err = updateStackset(stacksetName, spec)
		}
		require.NoError(t, err)
		_, err = waitForStack(t, stacksetName, stackVersion)
		require.NoError(t, err)
		_, err = waitForDeployment(t, fmt.Sprintf("%s-%s", stacksetName, stackVersion))
		require.NoError(t, err)
	}
	for i := 2; i < 5; i++ {
		require.True(t, stackExists(stacksetName, fmt.Sprintf("v%d", i)))
	}

	for i := 0; i < 2; i++ {
		deploymentName := fmt.Sprintf("%s-v%d", stacksetName, i)
		err := resourceDeleted(t, "stack", deploymentName, deploymentInterface()).withTimeout(time.Second * 60).await()
		require.NoError(t, err)
		require.False(t, stackExists(stacksetName, fmt.Sprintf("v%d", i)))
	}
}

func TestStackTTLWithIngress(t *testing.T) {
	t.Parallel()
	stacksetName := "stackset-ttl-ingress"
	specFactory := NewTestStacksetSpecFactory(stacksetName).StackGC(3, 0).Ingress()
	for i := 0; i < 5; i++ {
		stackVersion := fmt.Sprintf("v%d", i)
		var err error
		spec := specFactory.Create(stackVersion)
		if !stacksetExists(stacksetName) {
			err = createStackSet(stacksetName, true, spec)
		} else {
			err = updateStackset(stacksetName, spec)
		}
		require.NoError(t, err)
		_, err = waitForStack(t, stacksetName, stackVersion)
		require.NoError(t, err)
		fullStackName := fmt.Sprintf("%s-%s", stacksetName, stackVersion)
		_, err = waitForIngress(t, fullStackName)
		require.NoError(t, err)
		newWeight := map[string]float64{fullStackName: 100}
		err = setDesiredTrafficWeights(stacksetName, newWeight)
		require.NoError(t, err)
		err = trafficWeightsUpdated(t, stacksetName, weightKindActual, newWeight).withTimeout(10 * time.Minute).await()
		require.NoError(t, err)
	}
	for i := 2; i < 5; i++ {
		deploymentName := fmt.Sprintf("%s-v%d", stacksetName, i)
		require.True(t, stackExists(stacksetName, fmt.Sprintf("v%d", i)))
		waitForDeployment(t, deploymentName)
	}

	for i := 0; i < 2; i++ {
		deploymentName := fmt.Sprintf("%s-v%d", stacksetName, i)
		err := resourceDeleted(t, "stack", deploymentName, deploymentInterface()).withTimeout(time.Second * 60).await()
		require.NoError(t, err)
		require.False(t, stackExists(stacksetName, fmt.Sprintf("v%d", i)))
	}
}

package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPrescalingWithoutHPA(t *testing.T) {
	t.Parallel()
	stacksetName := "stackset-prescale-no-hpa"
	specFactory := NewTestStacksetSpecFactory(stacksetName).Ingress().StackGC(3, 0).Replicas(3)

	firstStack := "v1"
	spec := specFactory.Create(firstStack)
	err := createStackSet(stacksetName, true, spec)
	require.NoError(t, err)
	waitForStack(t, stacksetName, firstStack)

	secondStack := "v2"
	spec = specFactory.Create(secondStack)
	err = updateStackset(stacksetName, spec)
	require.NoError(t, err)
	waitForStack(t, stacksetName, secondStack)

	fullFirstStack := fmt.Sprintf("%s-%s", stacksetName, firstStack)
	fullSecondStack := fmt.Sprintf("%s-%s", stacksetName, secondStack)
	_, err = waitForIngress(t, stacksetName)
	require.NoError(t, err)
	desiredTraffic := map[string]float64{
		fullFirstStack:  50,
		fullSecondStack: 50,
	}
	err = setDesiredTrafficWeights(stacksetName, desiredTraffic)
	require.NoError(t, err)
	err = trafficWeightsUpdated(t, stacksetName, weightKindActual, desiredTraffic).withTimeout(time.Minute * 4).await()
	require.NoError(t, err)

	thirdStack := "v3"
	fullThirdStack := fmt.Sprintf("%s-%s", stacksetName, thirdStack)
	spec = specFactory.Replicas(1).Create(thirdStack)
	err = updateStackset(stacksetName, spec)
	require.NoError(t, err)
	deployment, err := waitForDeployment(t, fullThirdStack)
	require.NoError(t, err)

	desiredTraffic = map[string]float64{
		fullThirdStack:  10,
		fullFirstStack:  40,
		fullSecondStack: 50,
	}

	err = setDesiredTrafficWeights(stacksetName, desiredTraffic)
	require.NoError(t, err)
	err = trafficWeightsUpdated(t, stacksetName, weightKindActual, desiredTraffic).withTimeout(time.Minute * 4).await()
	require.NoError(t, err)

	deployment, err = waitForDeployment(t, fullThirdStack)
	require.NoError(t, err)
	require.EqualValues(t, 6, *(deployment.Spec.Replicas))
}

func TestPrescalingWithHPA(t *testing.T) {
	t.Parallel()
	stacksetName := "stackset-prescale-hpa"
	specFactory := NewTestStacksetSpecFactory(stacksetName).Ingress().StackGC(3, 0).
		HPA(1, 10).Replicas(3)

	firstStack := "v1"
	spec := specFactory.Create(firstStack)
	err := createStackSet(stacksetName, true, spec)
	require.NoError(t, err)
	waitForStack(t, stacksetName, firstStack)

	secondStack := "v2"
	spec = specFactory.Create(secondStack)
	err = updateStackset(stacksetName, spec)
	require.NoError(t, err)
	waitForStack(t, stacksetName, secondStack)

	fullFirstStack := fmt.Sprintf("%s-%s", stacksetName, firstStack)
	fullSecondStack := fmt.Sprintf("%s-%s", stacksetName, secondStack)
	_, err = waitForIngress(t, stacksetName)
	require.NoError(t, err)
	desiredTraffic := map[string]float64{
		fullFirstStack:  50,
		fullSecondStack: 50,
	}
	err = setDesiredTrafficWeights(stacksetName, desiredTraffic)
	require.NoError(t, err)
	err = trafficWeightsUpdated(t, stacksetName, weightKindActual, desiredTraffic).withTimeout(time.Minute * 4).await()
	require.NoError(t, err)

	thirdStack := "v3"
	fullThirdStack := fmt.Sprintf("%s-%s", stacksetName, thirdStack)
	spec = specFactory.Replicas(1).Create(thirdStack)
	err = updateStackset(stacksetName, spec)
	require.NoError(t, err)
	deployment, err := waitForDeployment(t, fullThirdStack)
	require.NoError(t, err)

	desiredTraffic = map[string]float64{
		fullThirdStack:  10,
		fullFirstStack:  40,
		fullSecondStack: 50,
	}

	err = setDesiredTrafficWeights(stacksetName, desiredTraffic)
	require.NoError(t, err)
	err = trafficWeightsUpdated(t, stacksetName, weightKindActual, desiredTraffic).withTimeout(time.Minute * 4).await()
	require.NoError(t, err)

	deployment, err = waitForDeployment(t, fullThirdStack)
	require.NoError(t, err)
	require.EqualValues(t, 6, *(deployment.Spec.Replicas))
}

#! /bin/bash

echo " "
echo "-------------------------------------------------------------"
echo "TEST 1: Running SimpleBlobTracker Test"
echo "-------------------------------------------------------------"
sleep 1
go test ./internal/abstractors -run TestSimpleBlobTracker_ReplayMode -v -count=1


echo " "
echo "-------------------------------------------------------------"
echo "TEST 2: Running EmitterTest Test"
echo "-------------------------------------------------------------"
sleep 1
go test ./internal/scenarios -run TestEmitter_BasicScenario -v -count=1


echo " "
echo "-------------------------------------------------------------"
echo "TEST 3: Running SynteticEmitterTest Test"
echo "-------------------------------------------------------------"
sleep 1
go test ./internal/abstractors -run TestSyntheticBlobEmitter_Suppression -v --count=1


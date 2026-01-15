// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2015 Mhd Sulhan <ms@kilabit.info>

package crf

import (
	"fmt"
	"os"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/dsv"
	"git.sr.ht/~shulhan/pakakeh.go/lib/mining/classifier"
	"git.sr.ht/~shulhan/pakakeh.go/lib/tabula"
)

func runCRF(t *testing.T, sampleFile, statFile, perfFile string, nstage, ntree int) {
	// read trainingset.
	samples := tabula.Claset{}
	_, e := dsv.SimpleRead(sampleFile, &samples)
	if e != nil {
		t.Fatal(e)
	}

	nbag := (samples.Len() * 63) / 100
	train, test, _, testListID := tabula.RandomPickRows(&samples, nbag, false)

	trainset := train.(tabula.ClasetInterface)
	testset := test.(tabula.ClasetInterface)

	crfRuntime := Runtime{
		Runtime: classifier.Runtime{
			StatFile: statFile,
			PerfFile: perfFile,
		},
		NStage: nstage,
		NTree:  ntree,
	}

	e = crfRuntime.Build(trainset)
	if e != nil {
		t.Fatal(e)
	}

	testset.RecountMajorMinor()
	fmt.Println("Testset:", testset)

	predicts, cm, probs := crfRuntime.ClassifySetByWeight(testset, testListID)

	fmt.Println("Confusion matrix:", cm)

	crfRuntime.Performance(testset, predicts, probs)
	e = crfRuntime.WritePerformance()
	if e != nil {
		t.Fatal(e)
	}
}

func TestPhoneme200_1(t *testing.T) {
	sampleFile := "../../testdata/phoneme/phoneme.dsv"
	perfFile := "phoneme_200_1.perf"
	statFile := "phoneme_200_1.stat"

	runCRF(t, sampleFile, statFile, perfFile, 200, 1)
}

func TestPhoneme200_10(t *testing.T) {
	sampleFile := "../../testdata/phoneme/phoneme.dsv"
	perfFile := "phoneme_200_10.perf"
	statFile := "phoneme_200_10.stat"

	runCRF(t, sampleFile, statFile, perfFile, 200, 10)
}

func TestMain(m *testing.M) {
	envTestCRF := os.Getenv("TEST_CRF")

	if len(envTestCRF) == 0 {
		os.Exit(0)
	}

	os.Exit(m.Run())
}

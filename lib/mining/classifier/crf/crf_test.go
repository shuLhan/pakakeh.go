// Copyright 2015 Mhd Sulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crf

import (
	"fmt"
	"os"
	"testing"

	"github.com/shuLhan/share/lib/dsv"
	"github.com/shuLhan/share/lib/mining/classifier"
	"github.com/shuLhan/share/lib/tabula"
)

var (
	SampleFile string
	PerfFile   string
	StatFile   string
	NStage     = 200
	NTree      = 1
)

func runCRF(t *testing.T) {
	// read trainingset.
	samples := tabula.Claset{}
	_, e := dsv.SimpleRead(SampleFile, &samples)
	if e != nil {
		t.Fatal(e)
	}

	nbag := (samples.Len() * 63) / 100
	train, test, _, testIds := tabula.RandomPickRows(&samples, nbag, false)

	trainset := train.(tabula.ClasetInterface)
	testset := test.(tabula.ClasetInterface)

	crfRuntime := Runtime{
		Runtime: classifier.Runtime{
			StatFile: StatFile,
			PerfFile: PerfFile,
		},
		NStage: NStage,
		NTree:  NTree,
	}

	e = crfRuntime.Build(trainset)
	if e != nil {
		t.Fatal(e)
	}

	testset.RecountMajorMinor()
	fmt.Println("Testset:", testset)

	predicts, cm, probs := crfRuntime.ClassifySetByWeight(testset, testIds)

	fmt.Println("Confusion matrix:", cm)

	crfRuntime.Performance(testset, predicts, probs)
	e = crfRuntime.WritePerformance()
	if e != nil {
		t.Fatal(e)
	}
}

func TestPhoneme200_1(t *testing.T) {
	SampleFile = "../../testdata/phoneme/phoneme.dsv"
	PerfFile = "phoneme_200_1.perf"
	StatFile = "phoneme_200_1.stat"

	runCRF(t)
}

func TestPhoneme200_10(t *testing.T) {
	SampleFile = "../../testdata/phoneme/phoneme.dsv"
	PerfFile = "phoneme_200_10.perf"
	StatFile = "phoneme_200_10.stat"
	NTree = 10

	runCRF(t)
}

func TestMain(m *testing.M) {
	envTestCRF := os.Getenv("TEST_CRF")

	if len(envTestCRF) == 0 {
		os.Exit(0)
	}

	os.Exit(m.Run())
}

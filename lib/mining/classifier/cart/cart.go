// Copyright 2015 Mhd Sulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package cart implement the Classification and Regression Tree by Breiman, et al.
// CART is binary decision tree.
//
// Breiman, Leo, et al. Classification and regression trees. CRC press,
// 1984.
//
// The implementation is based on Data Mining book,
//
// Han, Jiawei, Micheline Kamber, and Jian Pei. Data mining: concepts and
// techniques: concepts and techniques. Elsevier, 2011.
//
package cart

import (
	"fmt"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/mining/gain/gini"
	"github.com/shuLhan/share/lib/mining/tree/binary"
	"github.com/shuLhan/share/lib/numbers"
	libstrings "github.com/shuLhan/share/lib/strings"
	"github.com/shuLhan/share/lib/tabula"
)

const (
	// SplitMethodGini if defined in Runtime, the dataset will be splitted
	// using Gini gain for each possible value or partition.
	//
	// This option is used in Runtime.SplitMethod.
	SplitMethodGini = "gini"
)

const (
	// ColFlagParent denote that the column is parent/split node.
	ColFlagParent = 1
	// ColFlagSkip denote that the column would be skipped.
	ColFlagSkip = 2
)

//
// Runtime data for building CART.
//
type Runtime struct {
	// SplitMethod define the criteria to used for splitting.
	SplitMethod string `json:"SplitMethod"`
	// NRandomFeature if less or equal to zero compute gain on all feature,
	// otherwise select n random feature and compute gain only on selected
	// features.
	NRandomFeature int `json:"NRandomFeature"`
	// OOBErrVal is the last out-of-bag error value in the tree.
	OOBErrVal float64
	// Tree in classification.
	Tree binary.Tree
}

//
// New create new Runtime object.
//
func New(D tabula.ClasetInterface, splitMethod string, nRandomFeature int) (
	*Runtime, error,
) {
	runtime := &Runtime{
		SplitMethod:    splitMethod,
		NRandomFeature: nRandomFeature,
		Tree:           binary.Tree{},
	}

	e := runtime.Build(D)
	if e != nil {
		return nil, e
	}

	return runtime, nil
}

//
// Build will create a tree using CART algorithm.
//
func (runtime *Runtime) Build(D tabula.ClasetInterface) (e error) {
	// Re-check input configuration.
	switch runtime.SplitMethod {
	case SplitMethodGini:
		// Do nothing.
	default:
		// Set default split method to Gini index.
		runtime.SplitMethod = SplitMethodGini
	}

	runtime.Tree.Root, e = runtime.splitTreeByGain(D)

	return
}

//
// splitTreeByGain calculate the gain in all dataset, and split into two node:
// left and right.
//
// Return node with the split information.
//
func (runtime *Runtime) splitTreeByGain(D tabula.ClasetInterface) (
	node *binary.BTNode,
	e error,
) {
	node = &binary.BTNode{}

	D.RecountMajorMinor()

	// if dataset is empty return node labeled with majority classes in
	// dataset.
	nrow := D.GetNRow()

	if nrow <= 0 {
		if debug.Value >= 2 {
			fmt.Printf("[cart] empty dataset (%s) : %v\n",
				D.MajorityClass(), D)
		}

		node.Value = NodeValue{
			IsLeaf: true,
			Class:  D.MajorityClass(),
			Size:   0,
		}
		return node, nil
	}

	// if all dataset is in the same class, return node as leaf with class
	// is set to that class.
	single, name := D.IsInSingleClass()
	if single {
		if debug.Value >= 2 {
			fmt.Printf("[cart] in single class (%s): %v\n", name,
				D.GetColumns())
		}

		node.Value = NodeValue{
			IsLeaf: true,
			Class:  name,
			Size:   nrow,
		}
		return node, nil
	}

	if debug.Value >= 2 {
		fmt.Println("[cart] D:", D)
	}

	// calculate the Gini gain for each attribute.
	gains := runtime.computeGain(D)

	// get attribute with maximum Gini gain.
	MaxGainIdx := gini.FindMaxGain(&gains)
	MaxGain := gains[MaxGainIdx]

	// if maxgain value is 0, use majority class as node and terminate
	// the process
	if MaxGain.GetMaxGainValue() == 0 {
		if debug.Value >= 2 {
			fmt.Println("[cart] max gain 0 with target",
				D.GetClassAsStrings(),
				" and majority class is ", D.MajorityClass())
		}

		node.Value = NodeValue{
			IsLeaf: true,
			Class:  D.MajorityClass(),
			Size:   0,
		}
		return node, nil
	}

	// using the sorted index in MaxGain, sort all field in dataset
	tabula.SortColumnsByIndex(D, MaxGain.SortedIndex)

	if debug.Value >= 2 {
		fmt.Println("[cart] maxgain:", MaxGain)
	}

	// Now that we have attribute with max gain in MaxGainIdx, and their
	// gain dan partition value in Gains[MaxGainIdx] and
	// GetMaxPartValue(), we split the dataset based on type of max-gain
	// attribute.
	// If its continuous, split the attribute using numeric value.
	// If its discrete, split the attribute using subset (partition) of
	// nominal values.
	var splitV interface{}

	if MaxGain.IsContinu {
		splitV = MaxGain.GetMaxPartGainValue()
	} else {
		attrPartV := MaxGain.GetMaxPartGainValue()
		attrSubV := attrPartV.(libstrings.Row)
		splitV = attrSubV[0]
	}

	if debug.Value >= 2 {
		fmt.Println("[cart] maxgainindex:", MaxGainIdx)
		fmt.Println("[cart] split v:", splitV)
	}

	node.Value = NodeValue{
		SplitAttrName: D.GetColumn(MaxGainIdx).GetName(),
		IsLeaf:        false,
		IsContinu:     MaxGain.IsContinu,
		Size:          nrow,
		SplitAttrIdx:  MaxGainIdx,
		SplitV:        splitV,
	}

	dsL, dsR, e := tabula.SplitRowsByValue(D, MaxGainIdx, splitV)

	if e != nil {
		return node, e
	}

	splitL := dsL.(tabula.ClasetInterface)
	splitR := dsR.(tabula.ClasetInterface)

	// Set the flag to parent in attribute referenced by
	// MaxGainIdx, so it will not computed again in the next round.
	cols := splitL.GetColumns()
	for x := range *cols {
		if x == MaxGainIdx {
			(*cols)[x].Flag = ColFlagParent
		} else {
			(*cols)[x].Flag = 0
		}
	}

	cols = splitR.GetColumns()
	for x := range *cols {
		if x == MaxGainIdx {
			(*cols)[x].Flag = ColFlagParent
		} else {
			(*cols)[x].Flag = 0
		}
	}

	nodeLeft, e := runtime.splitTreeByGain(splitL)
	if e != nil {
		return node, e
	}

	nodeRight, e := runtime.splitTreeByGain(splitR)
	if e != nil {
		return node, e
	}

	node.SetLeft(nodeLeft)
	node.SetRight(nodeRight)

	return node, nil
}

// SelectRandomFeature if NRandomFeature is greater than zero, select and
// compute gain in n random features instead of in all features
func (runtime *Runtime) SelectRandomFeature(D tabula.ClasetInterface) {
	if runtime.NRandomFeature <= 0 {
		// all features selected
		return
	}

	ncols := D.GetNColumn()

	// count all features minus class
	nfeature := ncols - 1
	if runtime.NRandomFeature >= nfeature {
		// Do nothing if number of random feature equal or greater than
		// number of feature in dataset.
		return
	}

	// exclude class index and parent node index
	excludeIdx := []int{D.GetClassIndex()}
	cols := D.GetColumns()
	for x, col := range *cols {
		if (col.Flag & ColFlagParent) == ColFlagParent {
			excludeIdx = append(excludeIdx, x)
		} else {
			(*cols)[x].Flag |= ColFlagSkip
		}
	}

	// Select random features excluding feature in `excludeIdx`.
	var pickedIdx []int
	for x := 0; x < runtime.NRandomFeature; x++ {
		idx := numbers.IntPickRandPositive(ncols, false, pickedIdx,
			excludeIdx)
		pickedIdx = append(pickedIdx, idx)

		// Remove skip flag on selected column
		col := D.GetColumn(idx)
		col.Flag &^= ColFlagSkip
	}

	if debug.Value >= 1 {
		fmt.Println("[cart] selected random features:", pickedIdx)
		fmt.Println("[cart] selected columns        :", D.GetColumns())
	}
}

//
// computeGain calculate the gini index for each value in each attribute.
//
func (runtime *Runtime) computeGain(D tabula.ClasetInterface) (
	gains []gini.Gini,
) {
	if runtime.SplitMethod == SplitMethodGini {
		// create gains value for all attribute minus target class.
		gains = make([]gini.Gini, D.GetNColumn())
	}

	runtime.SelectRandomFeature(D)

	classVS := D.GetClassValueSpace()
	classIdx := D.GetClassIndex()
	classType := D.GetClassType()

	for x, col := range *D.GetColumns() {
		// skip class attribute.
		if x == classIdx {
			continue
		}

		// skip column flagged with parent
		if (col.Flag & ColFlagParent) == ColFlagParent {
			gains[x].Skip = true
			continue
		}

		// ignore column flagged with skip
		if (col.Flag & ColFlagSkip) == ColFlagSkip {
			gains[x].Skip = true
			continue
		}

		// compute gain.
		if col.GetType() == tabula.TReal {
			attr := col.ToFloatSlice()

			if classType == tabula.TString {
				target := D.GetClassAsStrings()
				gains[x].ComputeContinu(&attr, &target,
					&classVS)
			} else {
				targetReal := D.GetClassAsReals()
				classVSReal := libstrings.ToFloat64(classVS)

				gains[x].ComputeContinuFloat(&attr,
					&targetReal, &classVSReal)
			}
		} else {
			attr := col.ToStringSlice()
			attrV := col.ValueSpace

			if debug.Value >= 2 {
				fmt.Println("[cart] attr :", attr)
				fmt.Println("[cart] attrV:", attrV)
			}

			target := D.GetClassAsStrings()
			gains[x].ComputeDiscrete(&attr, &attrV, &target,
				&classVS)
		}

		if debug.Value >= 2 {
			fmt.Println("[cart] gain :", gains[x])
		}
	}
	return
}

//
// Classify return the prediction of one sample.
//
func (runtime *Runtime) Classify(data *tabula.Row) (class string) {
	node := runtime.Tree.Root
	nodev := node.Value.(NodeValue)

	for !nodev.IsLeaf {
		if nodev.IsContinu {
			splitV := nodev.SplitV.(float64)
			attrV := (*data)[nodev.SplitAttrIdx].Float()

			if attrV < splitV {
				node = node.Left
			} else {
				node = node.Right
			}
		} else {
			splitV := nodev.SplitV.([]string)
			attrV := (*data)[nodev.SplitAttrIdx].String()

			if libstrings.IsContain(splitV, attrV) {
				node = node.Left
			} else {
				node = node.Right
			}
		}
		nodev = node.Value.(NodeValue)
	}

	return nodev.Class
}

//
// ClassifySet set the class attribute based on tree classification.
//
func (runtime *Runtime) ClassifySet(data tabula.ClasetInterface) (e error) {
	nrow := data.GetNRow()
	targetAttr := data.GetClassColumn()

	for i := 0; i < nrow; i++ {
		class := runtime.Classify(data.GetRow(i))

		_ = targetAttr.Records[i].SetValue(class, tabula.TString)
	}

	return
}

//
// CountOOBError process out-of-bag data on tree and return error value.
//
func (runtime *Runtime) CountOOBError(oob tabula.Claset) (
	errval float64,
	e error,
) {
	// save the original target to be compared later.
	origTarget := oob.GetClassAsStrings()

	if debug.Value >= 2 {
		fmt.Println("[cart] OOB:", oob.Columns)
		fmt.Println("[cart] TREE:", &runtime.Tree)
	}

	// reset the target.
	oobtarget := oob.GetClassColumn()
	oobtarget.ClearValues()

	e = runtime.ClassifySet(&oob)

	if e != nil {
		// set original target values back.
		oobtarget.SetValues(origTarget)
		return
	}

	target := oobtarget.ToStringSlice()

	if debug.Value >= 2 {
		fmt.Println("[cart] original target:", origTarget)
		fmt.Println("[cart] classify target:", target)
	}

	// count how many target value is miss-classified.
	runtime.OOBErrVal, _, _ = libstrings.CountMissRate(origTarget, target)

	// set original target values back.
	oobtarget.SetValues(origTarget)

	return runtime.OOBErrVal, nil
}

//
// String yes, it will print it JSON like format.
//
func (runtime *Runtime) String() (s string) {
	s = fmt.Sprintf("NRandomFeature: %d\n"+
		" SplitMethod   : %s\n"+
		" Tree          :\n%v", runtime.NRandomFeature,
		runtime.SplitMethod,
		runtime.Tree.String())
	return s
}

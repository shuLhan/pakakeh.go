// Copyright 2015 Mhd Sulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package gini contain function to calculating Gini gain.
//
// Gini gain, which is an impurity-based criterion that measures the divergences
// between the probability distribution of the target attributes' values.
//
package gini

import (
	"fmt"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/numbers"
	libstrings "github.com/shuLhan/share/lib/strings"
)

//
// Gini contain slice of sorted index, slice of partition values, slice of Gini
// index, Gini value for all samples.
//
type Gini struct {
	// Skip if its true, the gain value would not be searched on this
	// instance.
	Skip bool
	// IsContinue define whether the Gini index came from continuous
	// attribute or not.
	IsContinu bool
	// Value of Gini index for all value in attribute.
	Value float64
	// MaxPartGain contain the index of partition which have the maximum
	// gain.
	MaxPartGain int
	// MaxGainValue contain maximum gain of index.
	MaxGainValue float64
	// MinIndexPart contain the index of partition which have the minimum
	// Gini index.
	MinIndexPart int
	// MinIndexGini contain minimum Gini index value.
	MinIndexValue float64
	// SortedIndex of attribute, sorted by values of attribute. This will
	// be used to reference the unsorted target attribute.
	SortedIndex []int
	// ContinuPart contain list of partition value for continuous attribute.
	ContinuPart []float64
	// DiscretePart contain the possible combination of discrete values.
	DiscretePart libstrings.Table
	// Index contain list of Gini Index for each partition.
	Index []float64
	// Gain contain information gain for each partition.
	Gain []float64
}

//
// ComputeDiscrete Given an attribute "src" with discrete value 'discval', and
// the target attribute "target" which contain n classes, compute the
// information gain of "src".
//
// The result is saved as gain value in MaxGainValue for each partition.
//
func (gini *Gini) ComputeDiscrete(src, discval, target, classes *[]string) {
	gini.IsContinu = false

	// create partition for possible combination of discrete values.
	gini.createDiscretePartition((*discval))

	if debug.Value >= 2 {
		fmt.Println("[gini] part :", gini.DiscretePart)
	}

	gini.Index = make([]float64, len(gini.DiscretePart))
	gini.Gain = make([]float64, len(gini.DiscretePart))
	gini.MinIndexValue = 1.0

	// compute gini index for all samples
	gini.Value = gini.compute(target, classes)

	gini.computeDiscreteGain(src, target, classes)
}

//
// computeDiscreteGain will compute Gini index and Gain for each partition.
//
func (gini *Gini) computeDiscreteGain(src, target, classes *[]string) {
	// number of samples
	nsample := float64(len(*src))

	if debug.Value >= 3 {
		fmt.Println("[gini] sample:", target)
		fmt.Printf("[gini] Gini(a=%s) = %f\n", (*src), gini.Value)
	}

	// compute gini index for each discrete values
	for i, subPart := range gini.DiscretePart {
		// check if sub partition has at least an element
		if len(subPart) == 0 {
			continue
		}

		sumGI := 0.0
		for _, part := range subPart {
			ndisc := 0.0
			var subT []string

			for _, el := range part {
				for t, a := range *src {
					if a != el {
						continue
					}

					// count how many sample with this discrete value
					ndisc++
					// split the target by discrete value
					subT = append(subT, (*target)[t])
				}
			}

			// compute gini index for subtarget
			giniIndex := gini.compute(&subT, classes)

			// compute probabilities of discrete value through all
			// samples
			p := ndisc / nsample

			probIndex := p * giniIndex

			// sum all probabilities times gini index.
			sumGI += probIndex

			if debug.Value >= 3 {
				fmt.Printf("[gini] subsample: %v\n", subT)
				fmt.Printf("[gini] Gini(a=%s) = %f/%f * %f = %f\n",
					part, ndisc, nsample,
					giniIndex, probIndex)
			}
		}

		gini.Index[i] = sumGI
		gini.Gain[i] = gini.Value - sumGI

		if debug.Value >= 3 {
			fmt.Printf("[gini] sample: %v\n", subPart)
			fmt.Printf("[gini] Gain(a=%s) = %f - %f = %f\n",
				subPart, gini.Value, sumGI,
				gini.Gain[i])
		}

		if gini.MinIndexValue > gini.Index[i] && gini.Index[i] != 0 {
			gini.MinIndexValue = gini.Index[i]
			gini.MinIndexPart = i
		}

		if gini.MaxGainValue < gini.Gain[i] {
			gini.MaxGainValue = gini.Gain[i]
			gini.MaxPartGain = i
		}
	}
}

//
// createDiscretePartition will create possible combination for discrete value
// in DiscretePart.
//
func (gini *Gini) createDiscretePartition(discval []string) {
	// no discrete values ?
	if len(discval) == 0 {
		return
	}

	// use set partition function to group the discrete values into two
	// subset.
	gini.DiscretePart = libstrings.Partition(discval, 2)
}

//
// ComputeContinu Given an attribute "src" and the target attribute "target"
// which contain n classes, compute the information gain of "src".
//
// The result of Gini partitions value, Gini Index, and Gini Gain is saved in
// ContinuPart, Index, and Gain.
//
func (gini *Gini) ComputeContinu(src *[]float64, target, classes *[]string) {
	gini.IsContinu = true

	// make a copy of attribute and target.
	A2 := make([]float64, len(*src))
	copy(A2, *src)

	T2 := make([]string, len(*target))
	copy(T2, *target)

	gini.SortedIndex = numbers.Floats64IndirectSort(A2, true)

	if debug.Value >= 1 {
		fmt.Println("[gini] attr sorted :", A2)
	}

	// sort the target attribute using sorted index.
	libstrings.SortByIndex(&T2, gini.SortedIndex)

	// create partition
	gini.createContinuPartition(&A2)

	// create holder for gini index and gini gain
	gini.Index = make([]float64, len(gini.ContinuPart))
	gini.Gain = make([]float64, len(gini.ContinuPart))
	gini.MinIndexValue = 1.0

	// compute gini index for all samples
	gini.Value = gini.compute(&T2, classes)

	gini.computeContinuGain(&A2, &T2, classes)
}

//
// createContinuPartition for dividing class and computing Gini index.
//
// This is assuming that the data `src` has been sorted in ascending order.
//
func (gini *Gini) createContinuPartition(src *[]float64) {
	l := len(*src)
	gini.ContinuPart = make([]float64, 0)

	// loop from first index until last index - 1
	for i := 0; i < l-1; i++ {
		sum := (*src)[i] + (*src)[i+1]
		med := sum / 2.0

		// If median is zero, its mean both left and right value is
		// zero. We are not allowing this, because it will result the
		// minimum Gini Index or maximum Gain value.
		if med == 0 {
			continue
		}

		// Reject if median is contained in attribute's value.
		// We use equality because if both src[i] and src[i+1] value
		// is equal, the median is equal to both of them.
		exist := false
		for j := 0; j <= i; j++ {
			if (*src)[j] == med {
				exist = true
				break
			}
		}
		if !exist {
			gini.ContinuPart = append(gini.ContinuPart, med)
		}
	}
}

//
// compute value for attribute T.
//
// Return Gini value in the form of,
//
// 1 - sum (probability of each classes in T)
//
func (gini *Gini) compute(target, classes *[]string) float64 {
	n := float64(len(*target))
	if n == 0 {
		return 0
	}

	classCount := libstrings.CountTokens(*target, *classes, true)

	var sump2 float64

	for x, v := range classCount {
		p := float64(v) / n
		sump2 += (p * p)

		if debug.Value >= 3 {
			fmt.Printf("[gini] compute (%s): (%d/%f)^2 = %f\n",
				(*classes)[x], v, n, p*p)
		}

	}

	return 1 - sump2
}

//
// computeContinuGain for each partition.
//
// The Gini gain formula we used here is,
//
// Gain(part,S) = Gini(S) - ((count(left)/S * Gini(left))
// 			+ (count(right)/S * Gini(right)))
//
// where,
// - left is sub-sample from S that is less than part value.
// - right is sub-sample from S that is greater than part value.
//
func (gini *Gini) computeContinuGain(src *[]float64, target, classes *[]string) {
	var gleft, gright float64
	var tleft, tright []string

	nsample := len(*src)

	if debug.Value >= 2 {
		fmt.Println("[gini] sorted data:", src)
		fmt.Println("[gini] Gini.Value:", gini.Value)
	}

	for p, contVal := range gini.ContinuPart {

		// find the split of samples between partition based on
		// partition value
		partidx := nsample
		for x, attrVal := range *src {
			if attrVal > contVal {
				partidx = x
				break
			}
		}

		nleft := partidx
		nright := nsample - partidx
		pleft := float64(nleft) / float64(nsample)
		pright := float64(nright) / float64(nsample)

		if partidx > 0 {
			tleft = (*target)[0:partidx]
			tright = (*target)[partidx:]

			gleft = gini.compute(&tleft, classes)
			gright = gini.compute(&tright, classes)
		} else {
			tleft = nil
			tright = (*target)[0:]

			gleft = 0
			gright = gini.compute(&tright, classes)
		}

		// count class in partition
		gini.Index[p] = ((pleft * gleft) + (pright * gright))
		gini.Gain[p] = gini.Value - gini.Index[p]

		if debug.Value >= 3 {
			fmt.Println("[gini] tleft:", tleft)
			fmt.Println("[gini] tright:", tright)

			fmt.Printf("[gini] GiniGain(%v) = %f - (%f * %f) + (%f * %f) = %f\n",
				contVal, gini.Value, pleft, gleft,
				pright, gright, gini.Gain[p])
		}

		if gini.MinIndexValue > gini.Index[p] && gini.Index[p] != 0 {
			gini.MinIndexValue = gini.Index[p]
			gini.MinIndexPart = p
		}

		if gini.MaxGainValue < gini.Gain[p] {
			gini.MaxGainValue = gini.Gain[p]
			gini.MaxPartGain = p
		}
	}
}

//
// GetMaxPartGainValue return the partition that have the maximum Gini gain.
//
func (gini *Gini) GetMaxPartGainValue() interface{} {
	if gini.IsContinu {
		return gini.ContinuPart[gini.MaxPartGain]
	}

	return gini.DiscretePart[gini.MaxPartGain]
}

//
// GetMaxGainValue return the value of partition which contain the maximum Gini
// gain.
//
func (gini *Gini) GetMaxGainValue() float64 {
	return gini.MaxGainValue
}

//
// GetMinIndexPartValue return the partition that have the minimum Gini index.
//
func (gini *Gini) GetMinIndexPartValue() interface{} {
	if gini.IsContinu {
		return gini.ContinuPart[gini.MinIndexPart]
	}

	return gini.DiscretePart[gini.MinIndexPart]
}

//
// GetMinIndexValue return the minimum Gini index value.
//
func (gini *Gini) GetMinIndexValue() float64 {
	return gini.MinIndexValue
}

//
// FindMaxGain find the attribute and value that have the maximum gain.
// The returned value is index of attribute.
//
func FindMaxGain(gains *[]Gini) (maxGainIdx int) {
	var gainValue float64
	var maxGainValue float64

	for i := range *gains {
		if (*gains)[i].Skip {
			continue
		}
		gainValue = (*gains)[i].GetMaxGainValue()
		if gainValue > maxGainValue {
			maxGainValue = gainValue
			maxGainIdx = i
		}
	}

	return
}

//
// FindMinGiniIndex return the index of attribute that have the minimum Gini index.
//
func FindMinGiniIndex(ginis *[]Gini) (minIndexIdx int) {
	var indexV float64
	minIndexV := 1.0

	for i := range *ginis {
		indexV = (*ginis)[i].GetMinIndexValue()
		if indexV > minIndexV {
			minIndexV = indexV
			minIndexIdx = i
		}
	}

	return
}

//
// String yes, it will print it JSON like format.
//
func (gini Gini) String() (s string) {
	s = fmt.Sprint("{\n",
		"  Skip          :", gini.Skip, "\n",
		"  IsContinu     :", gini.IsContinu, "\n",
		"  Index         :", gini.Index, "\n",
		"  Value         :", gini.Value, "\n",
		"  Gain          :", gini.Gain, "\n",
		"  MaxPartGain   :", gini.MaxPartGain, "\n",
		"  MaxGainValue  :", gini.MaxGainValue, "\n",
		"  MinIndexPart  :", gini.MinIndexPart, "\n",
		"  MinIndexValue :", gini.MinIndexValue, "\n",
		"  SortedIndex   :", gini.SortedIndex, "\n",
		"  ContinuPart   :", gini.ContinuPart, "\n",
		"  DiscretePart  :", gini.DiscretePart, "\n",
		"}")
	return
}

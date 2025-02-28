// Copyright 2022 The NLP Odyssey Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bart

import (
	"encoding/gob"

	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/mat/float"
	"github.com/nlpodyssey/spago/nn"
	"github.com/nlpodyssey/spago/nn/attention/multiheadattention"
	"github.com/nlpodyssey/spago/nn/normalization/layernorm"
)

var _ nn.Model = &CrossAttentionBlock{}

// ResidualNormCrossAttention is a cross-attention block with residual connection.
type ResidualNormCrossAttention interface {
	// Forward performs the forward pass.
	Forward(cache multiheadattention.Cache, seq1 []mat.Tensor, seq2 []mat.Tensor) ([]mat.Tensor, multiheadattention.Cache)
}

// CrossAttentionBlock implements a cross-attention block.
type CrossAttentionBlock struct {
	nn.Module
	Attention *multiheadattention.Model
	Norm      *layernorm.Model
}

func init() {
	gob.Register(&CrossAttentionBlock{})
}

// CrossAttentionBlockConfig provides configuration settings for a CrossAttentionBlock.
type CrossAttentionBlockConfig struct {
	Dim             int
	NumOfHeads      int
	NormalizeBefore bool
}

// NewCrossAttentionBlock returns a new CrossAttentionBlock.
func NewCrossAttentionBlock[T float.DType](c CrossAttentionBlockConfig) ResidualNormCrossAttention {
	block := &CrossAttentionBlock{
		Attention: multiheadattention.New[T](c.Dim, c.NumOfHeads, false, true),
		Norm:      layernorm.New[T](c.Dim, 1e-5),
	}
	if c.NormalizeBefore {
		return PreNormCrossAttentionBlock{block}
	}
	return PostNormCrossAttentionBlock{block}
}

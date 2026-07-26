package triage

import (
	"math"
	"regexp"
	"strings"
)

type TFIDFVectorizer struct {
	docFreq  map[string]int
	idf      map[string]float64
	vectors  [][]float64
	vocab    map[string]int
	tokenizer func(string) []string
}

var tokenSplitter = regexp.MustCompile(`[^a-zA-Z0-9\p{Han}]+`)

func defaultTokenizer(text string) []string {
	tokens := tokenSplitter.Split(strings.ToLower(text), -1)
	var result []string
	for _, t := range tokens {
		if len(t) >= 2 {
			result = append(result, t)
		}
	}
	return result
}

func NewTFIDFVectorizer(documents []string) *TFIDFVectorizer {
	v := &TFIDFVectorizer{
		docFreq:   make(map[string]int),
		idf:       make(map[string]float64),
		vocab:     make(map[string]int),
		tokenizer: defaultTokenizer,
	}
	v.buildVocab(documents)
	v.buildIDF(documents)
	v.vectors = make([][]float64, len(documents))
	for i, doc := range documents {
		v.vectors[i] = v.transform(doc)
	}
	return v
}

func (v *TFIDFVectorizer) buildVocab(documents []string) {
	df := make(map[string]int)
	idx := 0
	for _, doc := range documents {
		tokens := v.tokenizer(doc)
		seen := make(map[string]bool)
		for _, t := range tokens {
			if _, ok := v.vocab[t]; !ok {
				v.vocab[t] = idx
				idx++
			}
			if !seen[t] {
				df[t]++
				seen[t] = true
			}
		}
	}
	v.docFreq = df
}

func (v *TFIDFVectorizer) buildIDF(documents []string) {
	n := float64(len(documents))
	for term, df := range v.docFreq {
		v.idf[term] = math.Log((n+1)/(float64(df)+1)) + 1
	}
}

func (v *TFIDFVectorizer) transform(text string) []float64 {
	tokens := v.tokenizer(text)
	vec := make([]float64, len(v.vocab))
	tf := make(map[string]int)
	for _, t := range tokens {
		tf[t]++
	}
	for term, count := range tf {
		if idx, ok := v.vocab[term]; ok {
			idfVal := v.idf[term]
			vec[idx] = float64(count) * idfVal
		}
	}
	norm := 0.0
	for _, val := range vec {
		norm += val * val
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range vec {
			vec[i] /= norm
		}
	}
	return vec
}

func (v *TFIDFVectorizer) Transform(text string) []float64 {
	return v.transform(text)
}

func cosineSimilarity(a, b []float64) float64 {
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	if maxLen == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := 0; i < maxLen; i++ {
		var ai, bi float64
		if i < len(a) {
			ai = a[i]
		}
		if i < len(b) {
			bi = b[i]
		}
		dot += ai * bi
		normA += ai * ai
		normB += bi * bi
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
package triage

import (
	"math"
	"sync"
	"testing"
)

func TestDefaultClassifierIdentity(t *testing.T) {
	first := DefaultClassifier()
	for i := 0; i < 100; i++ {
		got := DefaultClassifier()
		if got != first {
			t.Errorf("第 %d 次调用返回不同的指针: 期望 %p, got %p", i+1, first, got)
		}
	}
}

func TestDefaultClassifierConcurrent(t *testing.T) {
	const n = 100
	ch := make(chan *TriageClassifier, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			ch <- DefaultClassifier()
		}()
	}
	wg.Wait()
	close(ch)

	first := DefaultClassifier()
	count := 0
	for got := range ch {
		count++
		if got != first {
			t.Errorf("goroutine 返回不同的指针: 期望 %p, got %p", first, got)
		}
	}
	if count != n {
		t.Errorf("期望收集 %d 个结果, got %d", n, count)
	}
}

func TestDefaultClassifierCorrectness(t *testing.T) {
	input := "CPU 使用率过高 redis"
	topK := 3

	defaultResults := DefaultClassifier().Classify(input, topK)
	freshResults := NewTriageClassifier(defaultSkills).Classify(input, topK)

	if len(defaultResults) != len(freshResults) {
		t.Fatalf("结果长度不同: DefaultClassifier=%d, NewTriageClassifier=%d",
			len(defaultResults), len(freshResults))
	}

	for i := range defaultResults {
		dr := defaultResults[i]
		fr := freshResults[i]
		if dr.Skill != fr.Skill {
			t.Errorf("第 %d 个结果 Skill 不同: DefaultClassifier=%s, NewTriageClassifier=%s",
				i, dr.Skill, fr.Skill)
		}
		if dr.Rank != fr.Rank {
			t.Errorf("第 %d 个结果 Rank 不同: DefaultClassifier=%d, NewTriageClassifier=%d",
				i, dr.Rank, fr.Rank)
		}
		tol := 1e-9
		if math.Abs(dr.Confidence-fr.Confidence) > tol {
			t.Errorf("第 %d 个结果 Confidence 不同: DefaultClassifier=%.10f, NewTriageClassifier=%.10f, 差值=%.10f",
				i, dr.Confidence, fr.Confidence, math.Abs(dr.Confidence-fr.Confidence))
		}
	}
}

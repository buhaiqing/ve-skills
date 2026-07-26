package triage

import (
	"math"
	"testing"
)

func TestCosineSimilarity(t *testing.T) {
	t.Run("orthogonal", func(t *testing.T) {
		a := []float64{1, 0, 0}
		b := []float64{0, 1, 0}
		sim := cosineSimilarity(a, b)
		if math.Abs(sim) > 1e-9 {
			t.Errorf("正交向量相似度应为0， got %f", sim)
		}
	})
	t.Run("identical", func(t *testing.T) {
		a := []float64{1, 2, 3}
		b := []float64{1, 2, 3}
		sim := cosineSimilarity(a, b)
		if math.Abs(sim-1.0) > 1e-9 {
			t.Errorf("相同向量相似度应为1， got %f", sim)
		}
	})
	t.Run("partial", func(t *testing.T) {
		a := []float64{1, 0, 0}
		b := []float64{1, 1, 0}
		sim := cosineSimilarity(a, b)
		if sim <= 0 || sim >= 1 {
			t.Errorf("部分相似向量相似度应在0到1之间， got %f", sim)
		}
	})
	t.Run("empty", func(t *testing.T) {
		sim := cosineSimilarity([]float64{}, []float64{})
		if sim != 0 {
			t.Errorf("空向量相似度应为0， got %f", sim)
		}
	})
	t.Run("unequal_lengths", func(t *testing.T) {
		a := []float64{1, 2}
		b := []float64{1, 2, 3}
		sim := cosineSimilarity(a, b)
		if sim < 0 || sim > 1 {
			t.Errorf("不等长向量相似度应在0到1之间， got %f", sim)
		}
	})
}

func TestTFIDFVectorizer(t *testing.T) {
	docs := []string{
		"Redis 缓存 慢查询",
		"MySQL 数据库 慢查询",
		"ECS 云服务器 实例 扩容",
	}
	v := NewTFIDFVectorizer(docs)

	if len(v.vocab) == 0 {
		t.Error("词汇表不应为空")
	}

	for _, vec := range v.vectors {
		if len(vec) != len(v.vocab) {
			t.Errorf("向量长度 %d 与词汇表大小 %d 不匹配", len(vec), len(v.vocab))
		}
	}

	t.Run("transform_new_text", func(t *testing.T) {
		vec := v.Transform("Redis 缓存")
		if len(vec) != len(v.vocab) {
			t.Errorf("转换向量长度 %d 与词汇表大小 %d 不匹配", len(vec), len(v.vocab))
		}
	})
}

func TestTriageClassifier(t *testing.T) {
	classifier := DefaultClassifier()

	tests := []struct {
		input    string
		expected string
	}{
		{"CPU 使用率过高 实例扩容", "ve-ecs-ops"},
		{"Redis 慢查询 内存不足", "ve-redis-ops"},
		{"MySQL 慢查询", "ve-rds-mysql-ops"},
		{"网络延迟 VPC 路由", "ve-vpc-ops"},
		{"Kafka 消息延迟", "ve-kafka-ops"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			results := classifier.Classify(tt.input, 3)
			if len(results) == 0 {
				t.Fatalf("分类结果不应为空")
			}
			if results[0].Skill != tt.expected {
				t.Errorf("期望首选技能为 %s， got %s (置信度: %.4f)",
					tt.expected, results[0].Skill, results[0].Confidence)
			}
		})
	}
}

func TestTriageClassifierTopK(t *testing.T) {
	classifier := DefaultClassifier()
	results := classifier.Classify("MySQL 慢查询 数据库", 2)
	if len(results) > 2 {
		t.Errorf("结果数量不应超过2， got %d", len(results))
	}
	if len(results) == 0 {
		t.Fatal("结果不应为空")
	}
}

func TestTriageClassifierFallback(t *testing.T) {
	classifier := DefaultClassifier()
	results := classifier.Classify("完全无关的随机输入 xyz abc", 3)
	if len(results) == 0 {
		t.Fatal("分类结果不应为空，应有回退结果")
	}
	if results[0].Skill != "ve-cms-ops" {
		t.Errorf("回退应返回 ve-cms-ops， got %s", results[0].Skill)
	}
}

func TestTriageClassifierConfidenceRange(t *testing.T) {
	classifier := DefaultClassifier()
	inputs := []string{
		"CPU 使用率过高 实例扩容",
		"Redis 慢查询 内存不足",
		"MySQL 慢查询",
		"网络延迟 VPC 路由",
		"Kafka 消息延迟",
		"完全无关的随机输入",
	}
	for _, input := range inputs {
		results := classifier.Classify(input, 3)
		for _, r := range results {
			if r.Confidence < 0 || r.Confidence > 1 {
				t.Errorf("置信度 %.4f 不在 [0,1] 范围内，输入: %s", r.Confidence, input)
			}
		}
	}
}

func TestDefaultSkillsCount(t *testing.T) {
	if len(defaultSkills) != 28 {
		t.Errorf("默认技能数量应为28， got %d", len(defaultSkills))
	}
}

func TestExplain(t *testing.T) {
	classifier := DefaultClassifier()
	result := ClassificationResult{Skill: "ve-ecs-ops", Confidence: 0.85, Rank: 1}
	explanation := classifier.Explain(result)
	if explanation == "" {
		t.Error("解释文本不应为空")
	}
	t.Logf("解释: %s", explanation)
}
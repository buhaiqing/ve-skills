package triage

import (
	"fmt"
	"sort"
)

type SkillDoc struct {
	Name        string
	Description string
	Keywords    []string
}

type ClassificationResult struct {
	Skill      string
	Confidence float64
	Rank       int
}

type TriageClassifier struct {
	vectorizer *TFIDFVectorizer
	skills     []SkillDoc
	vectors    [][]float64
}

var defaultSkills = []SkillDoc{
	{
		Name:        "ve-ecs-ops",
		Description: "ECS 弹性计算云服务器运维",
		Keywords:    []string{"ecs", "云服务器", "实例", "弹性计算", "cpu", "内存", "扩容", "缩容", "负载"},
	},
	{
		Name:        "ve-redis-ops",
		Description: "Redis 缓存数据库运维",
		Keywords:    []string{"redis", "缓存", "慢查询", "内存", "key", "过期", "集群", "哨兵"},
	},
	{
		Name:        "ve-rds-mysql-ops",
		Description: "RDS MySQL 关系型数据库运维",
		Keywords:    []string{"mysql", "rds", "数据库", "慢查询", "索引", "sql", "主从", "binlog"},
	},
	{
		Name:        "ve-vpc-ops",
		Description: "VPC 私有网络运维",
		Keywords:    []string{"vpc", "网络", "路由", "子网", "私有网络", "acl", "安全组", "nat"},
	},
	{
		Name:        "ve-kafka-ops",
		Description: "Kafka 消息队列运维",
		Keywords:    []string{"kafka", "消息队列", "topic", "consumer", "producer", "partition", "broker", "延迟"},
	},
	{
		Name:        "ve-mongodb-ops",
		Description: "MongoDB 文档数据库运维",
		Keywords:    []string{"mongodb", "mongo", "文档数据库", "collection", "索引", "聚合", "副本集", "分片"},
	},
	{
		Name:        "ve-elasticsearch-ops",
		Description: "Elasticsearch 搜索引擎运维",
		Keywords:    []string{"elasticsearch", "es", "搜索", "索引", "分片", "副本", "查询", "analyze"},
	},
	{
		Name:        "ve-cdn-ops",
		Description: "CDN 内容分发网络运维",
		Keywords:    []string{"cdn", "内容分发", "缓存", "加速", "域名", "回源", "带宽", "节点"},
	},
	{
		Name:        "ve-clb-ops",
		Description: "CLB 传统负载均衡运维",
		Keywords:    []string{"clb", "负载均衡", "四层", "七层", "转发", "后端", "健康检查", "监听器"},
	},
	{
		Name:        "ve-alb-ops",
		Description: "ALB 应用型负载均衡运维",
		Keywords:    []string{"alb", "应用负载均衡", "七层", "HTTP", "HTTPS", "路由", "转发规则", "云原生"},
	},
	{
		Name:        "ve-eip-ops",
		Description: "EIP 弹性公网IP运维",
		Keywords:    []string{"eip", "公网ip", "弹性公网", "带宽", "绑定", "解绑", "计费", "共享"},
	},
	{
		Name:        "ve-nat-ops",
		Description: "NAT 网关运维",
		Keywords:    []string{"nat", "网关", "地址转换", "snat", "dnat", "端口映射", "会话", "连接数"},
	},
	{
		Name:        "ve-vpn-ops",
		Description: "VPN 网关运维",
		Keywords:    []string{"vpn", "虚拟专用网", "ipsec", "ssl vpn", "隧道", "加密", "预共享密钥", "连通性"},
	},
	{
		Name:        "ve-dns-ops",
		Description: "DNS 域名解析运维",
		Keywords:    []string{"dns", "域名", "解析", "记录", "a记录", "cname", "mx", "ttl"},
	},
	{
		Name:        "ve-tos-ops",
		Description: "TOS 对象存储运维",
		Keywords:    []string{"tos", "对象存储", "bucket", "存储桶", "上传", "下载", "生命周期", "版本控制"},
	},
	{
		Name:        "ve-sls-ops",
		Description: "SLS 日志服务运维",
		Keywords:    []string{"sls", "日志服务", "log", "日志", "采集", "检索", "分析", "仪表盘"},
	},
	{
		Name:        "ve-iam-ops",
		Description: "IAM 访问管理运维",
		Keywords:    []string{"iam", "访问管理", "用户", "策略", "角色", "权限", "ak", "sk"},
	},
	{
		Name:        "ve-kms-ops",
		Description: "KMS 密钥管理运维",
		Keywords:    []string{"kms", "密钥管理", "加密", "解密", "密钥", "轮转", "cmk", "密钥对"},
	},
	{
		Name:        "ve-billing-ops",
		Description: "Billing 账单计费运维",
		Keywords:    []string{"billing", "账单", "计费", "费用", "成本", "预算", "告警", "退款"},
	},
	{
		Name:        "ve-vke-ops",
		Description: "VKE 容器服务运维",
		Keywords:    []string{"vke", "kubernetes", "k8s", "容器", "集群", "pod", "node", "namespace"},
	},
	{
		Name:        "ve-polar-mysql-ops",
		Description: "PolarDB MySQL 云原生数据库运维",
		Keywords:    []string{"polardb", "polar", "云原生", "mysql", "共享存储", "计算分离", "只读实例", "快速扩容"},
	},
	{
		Name:        "ve-rds-pg-ops",
		Description: "RDS PostgreSQL 关系型数据库运维",
		Keywords:    []string{"postgresql", "pg", "rds", "数据库", "慢查询", "索引", "主从", "扩展"},
	},
	{
		Name:        "ve-rds-ops",
		Description: "RDS 通用关系型数据库运维",
		Keywords:    []string{"rds", "数据库", "关系型", "备份", "恢复", "参数", "监控", "高可用"},
	},
	{
		Name:        "ve-nas-ops",
		Description: "NAS 网络附加存储运维",
		Keywords:    []string{"nas", "网络存储", "文件系统", "nfs", "cifs", "挂载", "容量", "性能"},
	},
	{
		Name:        "ve-cms-ops",
		Description: "CMS 监控服务运维",
		Keywords:    []string{"cms", "监控", "指标", "告警", "仪表盘", "metric", "阈值", "通知"},
	},
	{
		Name:        "ve-fg-ops",
		Description: "FG 函数计算运维",
		Keywords:    []string{"函数计算", "serverless", "fg", "事件驱动", "函数", "触发器", "运行时", "冷启动"},
	},
	{
		Name:        "ve-ark-ops",
		Description: "Ark 方舟AI模型服务运维",
		Keywords:    []string{"ark", "ai", "模型", "推理", "llm", "大模型", "方舟", "endpoint"},
	},
	{
		Name:        "ve-security-group-ops",
		Description: "安全组运维",
		Keywords:    []string{"安全组", "security group", "sg", "访问控制", "入站", "出站", "规则", "端口"},
	},
}

func NewTriageClassifier(skills []SkillDoc) *TriageClassifier {
	documents := make([]string, len(skills))
	for i, s := range skills {
		documents[i] = s.Description + " " + joinKeywords(s.Keywords)
	}
	vectorizer := NewTFIDFVectorizer(documents)
	return &TriageClassifier{
		vectorizer: vectorizer,
		skills:     skills,
		vectors:    vectorizer.vectors,
	}
}

func DefaultClassifier() *TriageClassifier {
	return NewTriageClassifier(defaultSkills)
}

func (c *TriageClassifier) Classify(input string, topK int) []ClassificationResult {
	inputVec := c.vectorizer.Transform(input)
	results := make([]ClassificationResult, len(c.skills))
	for i, skill := range c.skills {
		sim := cosineSimilarity(inputVec, c.vectors[i])
		results[i] = ClassificationResult{
			Skill:      skill.Name,
			Confidence: sim,
			Rank:       i + 1,
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Confidence > results[j].Confidence
	})
	for i := range results {
		results[i].Rank = i + 1
	}
	if topK > len(results) {
		topK = len(results)
	}
	results = results[:topK]
	if len(results) > 0 && results[0].Confidence < 0.05 {
		return []ClassificationResult{
			{Skill: "ve-cms-ops", Confidence: 0.1, Rank: 1},
		}
	}
	return results
}

func (c *TriageClassifier) Explain(result ClassificationResult) string {
	return fmt.Sprintf("技能 %s 匹配置信度为 %.4f，排名第 %d", result.Skill, result.Confidence, result.Rank)
}

func joinKeywords(keywords []string) string {
	result := ""
	for _, k := range keywords {
		result += k + " "
	}
	return result
}
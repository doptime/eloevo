package testout

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
	"time"

	_ "github.com/doptime/doptime/httpserve"
	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/redisdb"
)

// ========== 类型定义 ==========

type JudgementResult string

const (
	ResultPass      JudgementResult = "pass"       // ✅ 通过：找出谬误
	ResultErrorFree JudgementResult = "error_free" // 🟡 无错：命题本身正确
	ResultMisjudge  JudgementResult = "misjudge"   // 🟣 误判：证据不足或冲突
)

type ModalityType string

const (
	ModalityEmpirical   ModalityType = "empirical"   // 经验型（实验、观察）
	ModalityFormal      ModalityType = "formal"      // 形式型（逻辑、数学）
	ModalityTextual     ModalityType = "textual"     // 文本型（定义、引用）
	ModalityStatistical ModalityType = "statistical" // 统计型（数据、概率）
	ModalityComparative ModalityType = "comparative" // 对比型（AB测试、对照）
	ModalitySimulative  ModalityType = "simulative"  // 模拟型（仿真、推演）
	ModalityContextual  ModalityType = "contextual"  // 语境型（历史、文化）
)

type Evidence struct {
	ID            string       `json:"id"`
	ActionID      string       `json:"actionId"`
	Modality      ModalityType `json:"modality"`
	Score         float64      `json:"score"`
	Description   string       `json:"description"`
	Source        string       `json:"source"`
	Timestamp     int64        `json:"timestamp,omitempty"`
	ConflictsWith []string     `json:"conflictsWith,omitempty"`
}

type Claim struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	IsCorrect  bool   `json:"isCorrect"`
	GlitchHint string `json:"glitchHint,omitempty"`
	Correction string `json:"correction,omitempty"`
}

type Scene struct {
	ID               string                   `json:"id"`
	Title            string                   `json:"title"`
	Subject          string                   `json:"subject"`
	KnowledgePoint   string                   `json:"knowledgePoint"`
	Difficulty       int                      `json:"difficulty"`
	Claims           []Claim                  `json:"claims"`
	ActionsEnabled   []string                 `json:"actionsEnabled"`
	ModalityWeights  map[ModalityType]float64 `json:"modalityWeights"`
	TestoutThreshold struct {
		Score     float64 `json:"score"`
		Diversity int     `json:"diversity"`
	} `json:"testoutThreshold"`
	Hints []string `json:"hints"`
}

type Feedback struct {
	Strengths   []string `json:"strengths"`
	Suggestions []string `json:"suggestions"`
	NextHints   []string `json:"nextHints,omitempty"`
}

type Rewards struct {
	Points         int      `json:"points"`
	Achievements   []string `json:"achievements,omitempty"`
	UnlockedScenes []string `json:"unlockedScenes,omitempty"`
}

type JudgementResponse struct {
	Result    JudgementResult `json:"result"`
	Score     float64         `json:"score"`
	Diversity int             `json:"diversity"`
	Message   string          `json:"message"`
	Feedback  *Feedback       `json:"feedback,omitempty"`
	Rewards   *Rewards        `json:"rewards,omitempty"`
}

type EvidenceAnalysis struct {
	EvidenceID       string  `json:"evidenceId"`
	StrengthScore    float64 `json:"strengthScore"`
	RelevanceScore   float64 `json:"relevanceScore"`
	LogicalSoundness float64 `json:"logicalSoundness"`
	Reasoning        string  `json:"reasoning"`
}

type SmartFeedbackRequest struct {
	SceneID       string          `json:"sceneId"`
	Scene         Scene           `json:"scene"`
	Evidences     []Evidence      `json:"evidences"`
	CurrentResult JudgementResult `json:"currentResult"`
	MisjudgeCount int             `json:"misjudgeCount"`
	TimeElapsed   int             `json:"timeElapsed"`
}

type SmartFeedbackResponse struct {
	EnhancedFeedback Feedback           `json:"enhancedFeedback"`
	EvidenceAnalysis []EvidenceAnalysis `json:"evidenceAnalysis"`
	LearningPath     []string           `json:"learningPath"`
	Reasoning        string             `json:"reasoning"`
}

// ========== Redis Keys ==========

var (
	keyJudgementCache   = redisdb.NewHashKey[string, *JudgementResponse]()
	keyFeedbackAnalysis = redisdb.NewHashKey[string, *SmartFeedbackResponse]()
	keySceneData        = redisdb.NewHashKey[string, *Scene]()
)

// ========== Agent 定义 ==========

var AgentSmartFeedback = agent.Create(template.Must(template.New("SmartFeedbackAgent").Parse(`
You are an expert educational AI assistant for the TestOut learning game. Your role is to analyze student evidence and provide insightful, encouraging feedback that helps them improve their critical thinking skills.

<Scene Information>
Subject: {{.Scene.Subject}}
Knowledge Point: {{.Scene.KnowledgePoint}}
Difficulty: {{.Scene.Difficulty}}
Claim to Verify: {{.Scene.Claims}}
</Scene Information>

<Student's Evidence Stack>
{{range $idx, $evidence := .Evidences}}
Evidence {{$idx | plus1}}:
- ID: {{$evidence.ID}}
- Type: {{$evidence.Modality}}
- Score: {{$evidence.Score}}
- Description: {{$evidence.Description}}
- Source: {{$evidence.Source}}
{{end}}
</Student's Evidence Stack>

<Current Judgement>
Result: {{.CurrentResult}}
Misjudge Count: {{.MisjudgeCount}}
Time Elapsed: {{.TimeElapsed}} seconds
</Current Judgement>

# Your Tasks

1. Analyze each piece of evidence for:
   - Strength and relevance to the claim
   - Logical soundness
   - How it contributes to the overall argument

2. Provide personalized feedback that:
   - Highlights what the student did well (specific strengths)
   - Offers actionable suggestions for improvement
   - Encourages deeper thinking without giving away the answer
   - Matches the student's current level

3. Suggest a learning path based on:
   - Missing modalities of evidence
   - Conceptual gaps revealed by their approach
   - Next steps that would be most beneficial

Use tool calls to structure your analysis. Be encouraging and educational.
`))).WithToolCallMutextRun().WithTools(
	tool.NewTool("AnalyzeEvidence", "Analyze a specific piece of evidence", func(analysis *EvidenceAnalysis) {
		fmt.Printf("Analyzing evidence %s: Strength=%.2f, Relevance=%.2f, Logic=%.2f\n",
			analysis.EvidenceID, analysis.StrengthScore, analysis.RelevanceScore, analysis.LogicalSoundness)
	}),
	tool.NewTool("ProvideFeedback", "Generate comprehensive feedback", func(feedback *Feedback) {
		fmt.Printf("Generated feedback with %d strengths and %d suggestions\n",
			len(feedback.Strengths), len(feedback.Suggestions))
	}),
	tool.NewTool("SuggestLearningPath", "Suggest next steps for learning", func(path []string) {
		fmt.Printf("Learning path: %v\n", path)
	}),
)

var AgentEvidenceGenerator = agent.Create(template.Must(template.New("EvidenceGeneratorAgent").Parse(`
You are an evidence generation assistant for the TestOut learning game. Based on the student's action, generate realistic evidence that reflects what they would discover.

<Scene Context>
Subject: {{.Scene.Subject}}
Knowledge Point: {{.Scene.KnowledgePoint}}
Claim: {{.Claim.Text}}
Is Claim Correct: {{.Claim.IsCorrect}}
{{if .Claim.Correction}}Correction: {{.Claim.Correction}}{{end}}
</Scene Context>

<Action Taken>
Action ID: {{.ActionID}}
Action Category: {{.ActionCategory}}
Student's Query: {{.Query}}
</Action Taken>

Generate evidence that:
1. Is scientifically accurate and appropriate for the subject
2. Reflects what would realistically be discovered through this action
3. Helps students learn, but doesn't directly give away the answer
4. Has appropriate modality type and strength score

Use GenerateEvidence tool to create the evidence.
`))).WithToolCallMutextRun().WithTools(
	tool.NewTool("GenerateEvidence", "Generate a new evidence object", func(evidence *Evidence) {
		evidence.Timestamp = time.Now().Unix()
		fmt.Printf("Generated evidence: %s (Modality: %s, Score: %.2f)\n",
			evidence.ID, evidence.Modality, evidence.Score)
	}),
)

// ========== Service 方法 ==========

type TestOutService struct {
	modelList *models.ModelList
}

func NewTestOutService() *TestOutService {
	return &TestOutService{
		modelList: models.NewModelList("Qwen3Next80b",
			models.Qwen3B235Thinking2507,
			models.Qwen3Next80B),
	}
}

// GetSmartFeedback 生成智能反馈
func (s *TestOutService) GetSmartFeedback(req *SmartFeedbackRequest) (*SmartFeedbackResponse, error) {
	// 检查缓存
	cacheKey := fmt.Sprintf("%s:%s:%d", req.SceneID, req.CurrentResult, len(req.Evidences))
	feedbackCache := keyFeedbackAnalysis.ConcatKey(req.SceneID)

	if cached, err := feedbackCache.HGet(cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// 准备LLM调用参数
	params := map[string]any{
		agent.UseModel:  s.modelList.SequentialPick(),
		"Scene":         req.Scene,
		"Evidences":     req.Evidences,
		"CurrentResult": req.CurrentResult,
		"MisjudgeCount": req.MisjudgeCount,
		"TimeElapsed":   req.TimeElapsed,
	}

	// 调用LLM Agent
	response := &SmartFeedbackResponse{
		EvidenceAnalysis: []EvidenceAnalysis{},
		LearningPath:     []string{},
	}

	// 使用闭包捕获结果
	analysisResults := []EvidenceAnalysis{}
	var feedbackResult Feedback
	var learningPath []string

	// 重新定义工具来捕获结果
	tempAgent := agent.Create(template.Must(template.New("SmartFeedbackAgent").Parse(`
You are an expert educational AI assistant for the TestOut learning game. Your role is to analyze student evidence and provide insightful, encouraging feedback that helps them improve their critical thinking skills.

<Scene Information>
Subject: {{.Scene.Subject}}
Knowledge Point: {{.Scene.KnowledgePoint}}
Difficulty: {{.Scene.Difficulty}}
Claim to Verify: {{range .Scene.Claims}}{{.Text}}{{end}}
</Scene Information>

<Student's Evidence Stack>
{{range $idx, $evidence := .Evidences}}
Evidence {{add $idx 1}}:
- ID: {{$evidence.ID}}
- Type: {{$evidence.Modality}}
- Score: {{$evidence.Score}}
- Description: {{$evidence.Description}}
- Source: {{$evidence.Source}}
{{end}}
</Student's Evidence Stack>

<Current Judgement>
Result: {{.CurrentResult}}
Misjudge Count: {{.MisjudgeCount}}
Time Elapsed: {{.TimeElapsed}} seconds
</Current Judgement>

# Your Tasks

1. Use AnalyzeEvidence for EACH piece of evidence to evaluate its strength, relevance, and logical soundness.

2. Use ProvideFeedback to generate comprehensive, encouraging feedback with specific strengths and actionable suggestions.

3. Use SuggestLearningPath to recommend next steps based on the student's approach and missing elements.

Be specific, educational, and supportive in your analysis.
`))).WithToolCallMutextRun().WithTools(
		tool.NewTool("AnalyzeEvidence", "Analyze a specific piece of evidence", func(analysis *EvidenceAnalysis) {
			analysisResults = append(analysisResults, *analysis)
		}),
		tool.NewTool("ProvideFeedback", "Generate comprehensive feedback", func(feedback *Feedback) {
			feedbackResult = *feedback
		}),
		tool.NewTool("SuggestLearningPath", "Suggest next steps for learning", func(path []string) {
			learningPath = path
		}),
	)

	// 执行Agent调用
	result := tempAgent.Call(params)

	response.EvidenceAnalysis = analysisResults
	response.EnhancedFeedback = feedbackResult
	response.LearningPath = learningPath
	response.Reasoning = result.Content

	// 缓存结果
	feedbackCache.HSet(cacheKey, response)

	return response, nil
}

// GenerateEvidence 生成证据
func (s *TestOutService) GenerateEvidence(sceneID string, scene *Scene, claim *Claim, actionID, actionCategory, query string) (*Evidence, error) {
	params := map[string]any{
		agent.UseModel:   s.modelList.SequentialPick(),
		"Scene":          scene,
		"Claim":          claim,
		"ActionID":       actionID,
		"ActionCategory": actionCategory,
		"Query":          query,
	}

	var generatedEvidence *Evidence

	// 创建临时agent来捕获生成的证据
	tempAgent := agent.Create(template.Must(template.New("EvidenceGeneratorAgent").Parse(`
You are an evidence generation assistant for the TestOut learning game. Based on the student's action, generate realistic evidence that reflects what they would discover.

<Scene Context>
Subject: {{.Scene.Subject}}
Knowledge Point: {{.Scene.KnowledgePoint}}
Claim: {{.Claim.Text}}
Is Claim Correct: {{.Claim.IsCorrect}}
{{if .Claim.Correction}}Correction: {{.Claim.Correction}}{{end}}
</Scene Context>

<Action Taken>
Action ID: {{.ActionID}}
Action Category: {{.ActionCategory}}
Student's Query: {{.Query}}
</Action Taken>

Generate ONE piece of evidence by calling GenerateEvidence tool. The evidence should:
1. Be scientifically accurate and appropriate for the subject
2. Reflect what would realistically be discovered through this action
3. Help students learn without directly giving away the answer
4. Have appropriate modality type and strength score (0.0-1.0)
5. Include a clear description and source

Choose the modality based on action:
- controlled_test, experiment → empirical
- logical_check, proof → formal
- definition_lookup → textual
- data_analysis → statistical
- comparison → comparative
- simulation → simulative
- historical_context → contextual
`))).WithToolCallMutextRun().WithTools(
		tool.NewTool("GenerateEvidence", "Generate a new evidence object", func(evidence *Evidence) {
			evidence.Timestamp = time.Now().Unix()
			generatedEvidence = evidence
		}),
	)

	tempAgent.Call(params)

	if generatedEvidence == nil {
		return nil, fmt.Errorf("failed to generate evidence")
	}

	return generatedEvidence, nil
}

// EvaluateJudgement 评估判定结果（可选的AI增强）
func (s *TestOutService) EvaluateJudgement(scene *Scene, evidences []Evidence) (*JudgementResponse, error) {
	// 这里可以调用LLM来提供额外的判定洞察
	// 基础判定逻辑应该在前端的JudgementService中
	// 这里只是提供AI增强的反馈

	req := &SmartFeedbackRequest{
		SceneID:   scene.ID,
		Scene:     *scene,
		Evidences: evidences,
	}

	feedback, err := s.GetSmartFeedback(req)
	if err != nil {
		return nil, err
	}

	return &JudgementResponse{
		Result:    ResultMisjudge, // 实际判定应该由客户端完成
		Score:     0,
		Diversity: 0,
		Message:   "AI分析完成",
		Feedback:  &feedback.EnhancedFeedback,
	}, nil
}

// ========== HTTP Handlers ==========

// HandleGetSmartFeedback HTTP处理器
func (s *TestOutService) HandleGetSmartFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SmartFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	response, err := s.GetSmartFeedback(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Service error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGenerateEvidence HTTP处理器
func (s *TestOutService) HandleGenerateEvidence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SceneID        string `json:"sceneId"`
		Scene          Scene  `json:"scene"`
		Claim          Claim  `json:"claim"`
		ActionID       string `json:"actionId"`
		ActionCategory string `json:"actionCategory"`
		Query          string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	evidence, err := s.GenerateEvidence(
		req.SceneID,
		&req.Scene,
		&req.Claim,
		req.ActionID,
		req.ActionCategory,
		req.Query,
	)

	if err != nil {
		http.Error(w, fmt.Sprintf("Service error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(evidence)
}

// SetupRoutes 设置路由
func (s *TestOutService) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/testout/feedback", s.HandleGetSmartFeedback)
	mux.HandleFunc("/api/testout/evidence", s.HandleGenerateEvidence)
}

// ========== 辅助函数 ==========

func init() {
	// 注册模板函数
	funcMap := template.FuncMap{
		"plus1": func(i int) int { return i + 1 },
		"add":   func(a, b int) int { return a + b },
	}

	// 更新agent模板
	_ = funcMap
}

// Example 使用示例
func Example() {
	// 创建服务
	service := NewTestOutService()

	// 方式1: 直接调用方法
	req := &SmartFeedbackRequest{
		SceneID: "physics_001",
		Scene: Scene{
			ID:             "physics_001",
			Title:          "自由落体实验",
			Subject:        "物理",
			KnowledgePoint: "重力加速度",
			Difficulty:     3,
			Claims: []Claim{
				{
					ID:         "claim_1",
					Text:       "重物比轻物落得更快",
					IsCorrect:  false,
					Correction: "忽略空气阻力时，所有物体下落加速度相同",
				},
			},
		},
		Evidences: []Evidence{
			{
				ID:          "ev_1",
				ActionID:    "controlled_test",
				Modality:    ModalityEmpirical,
				Score:       0.8,
				Description: "在真空中测试，发现羽毛和铁球同时落地",
				Source:      "实验观察",
			},
		},
		CurrentResult: ResultMisjudge,
		MisjudgeCount: 1,
		TimeElapsed:   120,
	}

	feedback, err := service.GetSmartFeedback(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Feedback: %+v\n", feedback)

	// 方式2: 启动HTTP服务器
	mux := http.NewServeMux()
	service.SetupRoutes(mux)

	fmt.Println("TestOut Service running on :8080")
	// http.ListenAndServe(":8080", mux)
}

package models

import (
	"os"
	"sync"
	"time"

	"github.com/doptime/eloevo/utils"
	openai "github.com/sashabaranov/go-openai"
)

// Model represents an OpenAI model with its associated client and model name.
type Model struct {
	Client          *openai.Client
	Name            string
	TopP            float32
	TopK            int
	Temperature     float32
	ToolInPrompt    *ToolInPrompt
	avgResponseTime time.Duration
	lastReceived    time.Time
	requestPerMin   float64
	mutex           sync.RWMutex
}

func (model *Model) ResponseTime(duration ...time.Duration) time.Duration {
	if len(duration) == 0 {
		return model.avgResponseTime
	}
	model.mutex.Lock()
	defer model.mutex.Unlock()
	alpha := 0.1
	model.avgResponseTime += time.Duration(int64(float64(time.Duration(int64(duration[0]-model.avgResponseTime))) * alpha))
	model.requestPerMin += (60000000.0/float64(time.Since(model.lastReceived).Microseconds()+100) - model.requestPerMin) * 0.01
	model.lastReceived = time.Now()
	return model.avgResponseTime
}

// NewModel initializes a new Model with the given baseURL, apiKey, and modelName.
// It configures the OpenAI client to use a custom base URL if provided.
func NewModel(baseURL, apiKey, modelName string) *Model {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}
	client := openai.NewClientWithConfig(config)
	return &Model{
		Client:          client,
		Name:            modelName,
		avgResponseTime: 120 * time.Second,
	}
}
func (m *Model) WithToolsInSystemPrompt() *Model {
	m.ToolInPrompt = &ToolInPrompt{InSystemPrompt: true}
	return m
}
func (m *Model) WithToolsInUserPrompt() *Model {
	m.ToolInPrompt = &ToolInPrompt{InUserPrompt: true}
	return m
}
func (m *Model) WithTopP(topP float32) *Model {
	m.TopP = topP
	return m
}
func (m *Model) WithTopK(topK int) *Model {
	m.TopK = topK
	return m
}
func (m *Model) WithTemperature(temperature float32) *Model {
	m.Temperature = temperature
	return m
}

const (
	EndPoint8010   = "http://gpu.lan:8010/v1"
	EndPoint8009   = "http://gpu.lan:8009/v1"
	EndPoint8008   = "http://gpu.lan:8008/v1"
	EndPoint8007   = "http://gpu.lan:8007/v1"
	EndPoint8006   = "http://gpu.lan:8006/v1"
	EndPoint8003   = "http://gpu.lan:8003/v1"
	ApiKey         = "token-deaf"
	ApiKeyDeepseek = "sk-2d9e2689120c4544820485740ea2f36c"
	NameQwen32B    = "Qwen/Qwen2.5-32B-Instruct-AWQ"

	NameQwen32BCoder      = "Qwen/Qwen2.5-Coder-32B-Instruct-AWQ"
	NameQwen32BCoderLocal = "/home/deaf/.cache/huggingface/hub/models--Qwen--Qwen2.5-32B-Instruct-AWQ/snapshots/5c7cb76a268fc6cfbb9c4777eb24ba6e27f9ee6c"

	NameQwen72B      = "Qwen/Qwen2.5-72B-Instruct-AWQ"
	NameQwen72BLocal = "/home/deaf/.cache/huggingface/hub/models--Qwen--Qwen2.5-72B-Instruct-AWQ/snapshots/698703eae6604af048a3d2f509995dc302088217"
	//NameQwen14B = "Qwen/Qwen2.5-14B-Instruct-AWQ"
	NameQwen7B         = "Qwen/Qwen2.5-7B-Instruct-AWQ"
	NameGemma          = "neuralmagic/gemma-2-9b-it-quantized.w4a16"
	NameMistralNemo    = "shuyuej/Mistral-Nemo-Instruct-2407-GPTQ"
	NameMistralSmall   = "AMead10/Mistral-Small-Instruct-2409-awq"
	NameMistralNemoAwq = "casperhansen/mistral-nemo-instruct-2407-awq"
	NameLlama38b       = "neuralmagic/Meta-Llama-3.1-8B-Instruct-quantized.w4a16"
	NameMarcoo1        = "AIDC-AI/Marco-o1"
	NamePhi4           = "/home/deaf/.cache/huggingface/hub/models--Orion-zhen--phi-4-awq/snapshots/bc73c60ec9d246127dff940b3331c5464f18442e"

	NameLlama33_70b = "casperhansen/llama-3.3-70b-instruct-awq"
	NameDeepseek    = "deepseek-chat"

	//NameQwQ32B = "/home/deaf/.cache/huggingface/hub/models--KirillR--QwQ-32B-Preview-AWQ/snapshots/b082e5c095a17c50cc78fc6fe43a0eae326bd203"
)

// Initialize all models with their corresponding endpoints and names.
var (
	ModelQwen32B = NewModel(EndPoint8008, ApiKey, NameQwen32B)

	ModelQwen32BCoder      = NewModel(EndPoint8007, ApiKey, NameQwen32BCoder)
	ModelQwen32BCoderLocal = NewModel(EndPoint8007, ApiKey, NameQwen32BCoderLocal)

	ModelQwen72B      = NewModel(EndPoint8007, ApiKey, NameQwen72B)
	ModelQwen72BLocal = NewModel(EndPoint8007, ApiKey, NameQwen72BLocal)

	ModelQwenQvq72B = NewModel(EndPoint8007, ApiKey, "/home/deaf/.cache/huggingface/hub/models--kosbu--QVQ-72B-Preview-AWQ/snapshots/9f763dc5a3bf51ed157aee12a8aae4ae8e7c1926")

	ModelQwen14B        = NewModel("http://gpu.lan:1214/v1", ApiKey, "/home/deaf/.cache/huggingface/hub/models--Qwen--Qwen2.5-14B-Instruct-AWQ/snapshots/539535859b135b0244c91f3e59816150c8056698")
	ModelQwen7B         = NewModel(EndPoint8007, ApiKey, NameQwen7B)
	ModelPhi3           = NewModel(EndPoint8006, ApiKey, "neuralmagic/Phi-3-medium-128k-instruct-quantized.w4a16")
	ModelGemma          = NewModel(EndPoint8006, ApiKey, NameGemma)
	ModelMistralNemo    = NewModel(EndPoint8003, ApiKey, NameMistralNemo)
	ModelMistralSmall   = NewModel(EndPoint8003, ApiKey, NameMistralSmall)
	ModelMistralNemoAwq = NewModel(EndPoint8003, ApiKey, NameMistralNemoAwq)
	ModelLlama38b       = NewModel(EndPoint8007, ApiKey, NameLlama38b)
	ModelMarcoo1        = NewModel(EndPoint8008, ApiKey, NameMarcoo1)
	ModelQwen32B12K     = NewModel(EndPoint8008, ApiKey, NameQwen32B)
	ModelLlama33_70b    = NewModel(EndPoint8007, ApiKey, NameLlama33_70b)
	//ModelDeepseek       = NewModel(EndPointDeepseek, ApiKeyDeepseek, NameDeepseek)
	ModelQwen2_1d5B    = NewModel("http://gpu.lan:8215/v1", ApiKey, "/home/deaf/.cache/huggingface/hub/models--Qwen--Qwen2.5-1.5B-Instruct/snapshots/989aa7980e4cf806f80c7fef2b1adb7bc71aa306")
	ModelQwen2_7B      = NewModel("http://gpu.lan:1207/v1", ApiKey, "/home/deaf/.cache/huggingface/hub/models--Qwen--Qwen2.5-7B-Instruct-AWQ/snapshots/b25037543e9394b818fdfca67ab2a00ecc7dd641")
	DeepSeekR1_Qwen_14 = NewModel("http://gpu.lan:3214/v1", ApiKey, "/home/deaf/.cache/huggingface/hub/models--casperhansen--deepseek-r1-distill-qwen-14b-awq/snapshots/1874537e80f451042f7993dfa2b21fd25b4e7223")
	DeepSeekR132B      = NewModel("http://gpu.lan:4733/v1", ApiKey, "DeepSeek-R1-Distill-Qwen-32B-AWQ").WithTopP(0.6)
	DSV3Baidu          = NewModel("https://qianfan.baidubce.com/v2", os.Getenv("BDAPIKEY"), "deepseek-v3").WithTopP(0.6)
	DeepSeekV3         = NewModel("https://api.deepseek.com/", utils.TextFromFile("/Users/yang/eloevo/.vscode/DSAPIKEY.txt"), "deepseek-chat").WithTopP(0.6).WithToolsInSystemPrompt()
	//https://tbnx.plus7.plus/token
	DeepSeekV3TB     = NewModel("https://tbnx.plus7.plus/v1", os.Getenv("DSTB"), "deepseek-chat").WithTopP(0.6)
	GeminiTB         = NewModel("https://tao.plus7.plus/v1", os.Getenv("geminitb"), "gemini-2.0-flash-exp").WithTopP(0.8).WithToolsInUserPrompt()
	GeminiFlashLight = NewModel("https://tao.plus7.plus/v1", "gemini-2.0-flash-lite", "gemini-2.0-flash-light").WithTopP(0.8).WithToolsInUserPrompt()
	GPT41Mini        = NewModel("https://tao.plus7.plus/v1", os.Getenv("geminitb"), "gpt-4.1-mini").WithTopP(0.8)

	DolphinR1Mistral24B = NewModel("http://gpu.lan:4733/v1", ApiKey, "Dolphin3.0-R1-Mistral-24B-AWQ").WithToolsInSystemPrompt()
	FuseO1              = NewModel("http://gpu.lan:4732/v1", ApiKey, "FuseO1").WithTopP(0.92).WithTemperature(0.6).WithTopK(40)
	Qwq32B              = NewModel("http://gpu.lan:1232/v1", ApiKey, "QwQ-32B").WithTopP(0.92).WithTemperature(0.6) //.WithTopK(40)
	Qwen32B             = NewModel("http://gpu.lan:1232/v1", ApiKey, "Qwen25B32")

	GLM32B = NewModel("http://gpu.lan:19732/v1", ApiKey, "glmz132b").WithTemperature(0.6).WithTopP(0.95) //.WithTopK(40) .WithToolInPrompt(true)

	Gemma3B27           = NewModel("http://gpu.lan:5527/v1", ApiKey, "gemma3").WithTopP(0.92).WithTemperature(0.9).WithToolsInUserPrompt()
	Gemma3B12           = NewModel("http://gpu.lan:5527/v1", ApiKey, "gemma3b12").WithTopP(0.95).WithTemperature(1.0).WithToolsInSystemPrompt()
	Qwen30BA3           = NewModel("http://gpu.lan:12303/v1", ApiKey, "qwen30ba3").WithTemperature(0.7).WithTopP(0.8)
	Qwen3B14            = NewModel("http://gpu.lan:1214/v1", ApiKey, "qwen3b14").WithTemperature(0.7).WithTopP(0.8)
	Qwen3B32Nonthinking = NewModel("http://gpu.lan:1214/v1", ApiKey, "qwen3b32").WithTemperature(0.7).WithTopP(0.8)
	Qwen3B32Thinking    = NewModel("http://gpu.lan:1214/v1", ApiKey, "qwen3b32").WithTemperature(0.6).WithTopP(0.95)

	Reka3 = NewModel("http://gpu.lan:43813/v1", ApiKey, "reka3").WithTopP(0.92).WithTemperature(0.9).WithToolsInSystemPrompt()
	Phi4  = NewModel("http://gpu.lan:7214/v1", ApiKey, "phi4").WithToolsInSystemPrompt()

	//ModelDefault        = ModelQwen32BCoderLocal
	ModelDefault = ModelQwen72BLocal
)

package learnbychoose

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"text/template"
	"time"

	// "github.com/yourbasic/graph"

	"github.com/doptime/doptime/api"
	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/redisdb"
	"github.com/remeh/sizedwaitgroup"
)

var AgentGetWordLearningData = agent.NewAgent().WithTemplate(template.Must(template.New("GetWordLearningData").Parse(`
给定一个词语 或 问题  或  一个知识点
{{.Word}}

现在的目标是1)生成一张图片作为学习材料。2)生成若干Bullet notes 作为学习材料。学习材料是对这个词的深度理解和创造性的直观应用。

更进一步来说
## 对图片来的要求是：
- 捕获语义的重要实质: 通过查看该图片，用户能够触及对这个词的实质的理解。
- 比对性原则：为区别于容易混淆的概念。需要对这个概念的核心特征给与特写。如果必要需要引入2张子图或4张子图，对核心特征进行强调和凸出。
- 感性直观: 以感性直观的呈现核心概念。
- 注意力引导：通过注意力引导，确保用户注意力能停留在关键细节上。
- 符合Gestalt原则：确保不同文化背景的人，不同年龄段的人，都能有相同的理解。
- 图片应该是本地生成的，并且提供的是本地的图片URL；目前已经测试网络图片的Url 都是幻觉，都是不存在的。

## 对Bullet notes的要求是：
- 视觉化的表征倾向: 倾向于优先使用emoji , ASCII Art等视觉化的文本表征
- 压缩文字: 使用少量但是必要的文字，以Bullet notes 方式呈现。
- 每一行一种语义或者是理解: 不同的行之间用换行符分隔。 
- 行数限制: 3-7行
- 符合Gestalt原则，尽可能使得不同文化背景的人，不同年龄段的人，都能有相同的理解。

TODO： 
1. 生成一个30-80个符号左右的学习材料
2. 生成最终的图片

先完成TODO, 并再次调用ToolCall 来保存学习材料

`))).WithToolCallMutextRun().WithTools(tool.NewTool("SaveWordLearningData", "For learning purpose, given a word, generate a bunch of related learning materials for that keywords ", func(newItem *WordLearningData) {
	if newItem.Word == "" || newItem.ImageURL != "" {
		return
	}
	// download image
	resp, err := http.Get(newItem.ImageURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		// 检查 HTTP 状态码
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("请求失败，HTTP 状态码: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		}
		return
	}
	defer resp.Body.Close() // 确保在函数结束时关闭响应体
	// 读取响应体内容
	newItem.ImageRawData, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("读取响应体失败: %v", err)
		return
	}
	keyWordLearningData.HSet(newItem.Word, newItem)
}))

type WordLearningData struct {
	Word                           string    `description:"string, The word or knowledge point to learn about"`
	AssociativeLearningBulletNotes string    `description:"string, Associative learning notes related to the word"`
	ImageURL                       string    `description:"URL of the image related to the word" msgpack:"-"`
	ImageRawData                   []byte    `description:"-"`
	CreatedAt                      time.Time `description:"-" `
}

var keyWordLearningData = redisdb.NewHashKey[string, *WordLearningData]()
var ApiWordLearningData = api.Api(func(in []string) (out []*WordLearningData, err error) {
	var params []interface{}
	for _, word := range in {
		params = append(params, word)
	}
	out, _ = keyWordLearningData.HMGET(params...)

	MaxThreads := 1
	swg := sizedwaitgroup.New(MaxThreads)

	for i, word := range in {
		if out[i] == nil && MaxThreads > 0 && word != "" {
			swg.Add()
			MaxThreads--

			go func(idx int, w string) {
				defer swg.Done()

				err := AgentGetWordLearningData.WithModels(models.Gemini20FlashImageAigpt).Call(context.Background(), map[string]any{
					"Word": w,
				})
				if err != nil {
					fmt.Printf("Error calling AgentGetWordLearningData for word %s: %v\n", w, err)
					return
				}
				out[idx], _ = keyWordLearningData.HGet(w)
			}(i, word)
		}
	}

	swg.Wait()
	return out, nil
}, api.WithApiKey("WordLearningData"))

func Debug() {
	fmt.Println("Debugging AgentGetWordLearningData...")
}

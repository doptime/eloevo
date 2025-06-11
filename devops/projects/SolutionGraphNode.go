package projects

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/doptime/redisdb"
	"github.com/samber/lo"
)

type SolutionGraphNode struct {
	Id                string `description:"Required when update. Id, string, unique." milvus:"PK,in,out"`
	Pathname          string `description:"Ascii pathname of current node。pathname is multi-level, using bullet name to denodes node's modualized intention. extension name such as .md ... is needed"`
	BulletDescription string `msgpack:"alias:BulletName" description:"Required when create. item of the solution. Bullet Description of Module, Constraints, Guidelines, Architecturals, Nexus or Specifications."`
	Content           string `msgpack:"alias:Detail" description:"Complete description or implementation details of the solution item. Content be stored in file. For SuperEdge nodes, this includes architectural principles, constraints, or guidelines. For Module nodes, this includes the specific solution implementation, algorithms, or technical specifications."`

	SuperEdge bool `description:"bool,true if the item is super edge of the solution graph. super edge 描述节点之间的协议,约定,约束,标准,规范,想法,技术路线,时间限制,资源限制,法律客户需求,反馈限制、层次化约束等 "`

	//SuperEdgeNodes []string `description:"array of Ids. If this node is super edge. here lists the child nodes that belongs to this SuperEdge. SuperEdgeNodes不能包含超边节点，因为超边节点实际上是图的边而不是图的节点，超边包含超边会破坏图结构. \nRequired by SuperEdge item. update each time super edge revised. "`

	Importance int64 `description:"int, value >=- 1 and value<= 10, Importance. \nRequired when create; optional when update. making Importance < 0  to Remove the item."`
	Priority   int64 `description:"int, value >= 0 and value <= 10 . \n Required for module node. use in Gatt chart to determin the priority of the item. the lower the higher the priority."`

	EmbedingVector []float32 `description:"-" milvus:"dim=1024,index" `

	SelfMemo string `description:"Self memo information. Such as missing elements, constrains collision, or what to do in further iterations."`

	//被人类专家标记为锁定的条目。Locked = true. 不能被删除和修改
	Locked bool `description:"-"`

	HashKey redisdb.HashKey[string, *SolutionGraphNode] `description:"-" msgpack:"-"`
}

// 	//初始添加的时候得分为0，Elo 后产生Elo分数
// 	Elo            float64 `description:"-"`
// func (u *SolutionGraphNode) ScoreAccessor(delta ...int) int {
// 	if len(delta) > 0 {
// 		u.Elo += float64(delta[0])
// 		KeyBusinessDronebot.HSet(u.Id, u)
// 	}

//		return int(u.Elo)
//	}
func (u *SolutionGraphNode) GetId() string {
	return u.Id
}

func (u *SolutionGraphNode) Embed(embed ...[]float32) []float32 {
	if len(embed) > 0 {
		u.EmbedingVector = embed[0]
		u.HashKey.HSet(u.Id, u)
	}
	return u.EmbedingVector
}

func (u *SolutionGraphNode) String(layer ...int) string {
	numLayer := append(layer, 0)[0]
	indence := strings.Repeat("\t", numLayer)
	SelfMemo := lo.Ternary(u.SelfMemo != "", " SelfMemo: "+u.SelfMemo, "")
	Detail := lo.Ternary(u.Content != "", " Detail: "+u.Content, "")
	return fmt.Sprint(indence, "BulletDescription: ", u.BulletDescription, Detail, SelfMemo, " [Id:", u.Id, lo.Ternary(u.SuperEdge, " SuperEdge", ""), " importance:", strconv.Itoa(int(u.Importance)), " priority:", strconv.Itoa(int(u.Priority)), "]\n")
	//return fmt.Sprint(indence, "Id:", u.Id, " Importance:", u.Importance, communityCore, " Priority:", u.Priority, "\n", u.Item, "\n\n")
}

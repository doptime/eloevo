package scrum

type ProductGoal string

var ProductGoalAirplane ProductGoal = `系统愿景:实现AI时代，无人机作为基础物流平台。确保方案和应用的极致的简单、可靠、低成本、高效用。
你的核心目标是基于第一性原理的工程学实现，构建一个在无人机平台和机器人应用领域具有高价值、高可行性的项目模块矩阵，这些项目应能在未来的世界中产生最大的联合商业效用和社会效用。

## 涉及的目标行业包括：
- AI-Driven Bobotic Development
- Robotic As a Service
- Drone Technology & Solution
- Suppy Chain & Drone & Logistics Technology
    -如从农业产地直供最终零售点
- Sustainable Packaging Technology
- Sustainable Transportation Infrastructure

## 部分愿景:
- 借助外接电源或超高速放电电池，垂直起飞的固定翼无人机
- 它是一个非常便利的载具平台。可以提供各种机器人的投送和收回服务
- 由于极高的滑翔比。它的物流成本只有汽车的1/10和船运的1/2. 可以在全球内完成有中继的长途运输
- 它可以借助地形和动态风向变化，实现能量的节约。
- 联合多机器人和多飞机。送外卖，入户医疗检查。在户外部署就餐，住宿，岗哨，它能做很多。
- 在未来的世界中，基于无人机平台的全球即时物流和资源分配系统，是最重要的基础设施。
- AI 驱动的自主机器人团队，能够高效协作完成复杂任务。`

var ProductGoalAntiAging ProductGoal = `系统愿景:实现AI时代，抗衰老和健康管理的日常方案。确保方案和应用的极致的简单、可靠、低成本、高效用。	
## 其它重要约束:
- 采用市场可以购买的药品、检测设备等
- 实现关键衰老相关的份子指标的改善
- 在器官层级上实现衰老指征的改善
- 具体到日程级别，剂量、时间、动作级别的可操作性
- 基于对衰老的第一性原理的理解。
- 最终求解得到的解决方案是给个人日常使用。并不用于商业化运作。

`
var ProductGoalUniLearning ProductGoal = `
# 系统愿景: 
## 实现一个学习、面试、培训一体化的平台；
### 对任何特定的知识，先要构建一个非常具象化的场景；
    - 本主题对应的场景是一个单一的微场景 / 单一场景的游戏。
    - 有可交互的情境。
    - 有适当的环节需要用到特定的知识作为关键的决策判断。
    - 用户的交互操作非常直观、流畅的方式进行。 比如选择、堆放等。
    - 交互要能够通过摄像头，用手势指向来完成。以便用户可以不需要在屏幕前操作，避免近视。
    - 用户不需要非常直接地了解该知识的背景和细节。用户只需要在交互当中，能部分感受到知识的内容。这样可以容纳更高的知识密度。要求该场景可以最有效体现知识的关键作用。
    - 全程在用户参与的情况下，全过程录像可以作为YTB的一个精彩的节目。通过节目推广平台是一个重要策略。要提高场景和互动的节目效果。
    - 场景游戏应该体现格式塔心理学原则。避免要求用户具有额外的的背景知识、或者是额外的宗教、文化背景。

## 软件架构与实施约束:
  - 所有的场景中的元素使用svg或其它方式，每个元素分别定义。
  - 代码必须以全量代码而不是增量代码的形式给出。不能只给出修改的部分。因为Refine只支持代码的全量替换。
  - 现在的项目是更大的react next.js项目的一个子项目。 不需要关心总项目的配置。只需要关注该子项目的实现。
  - 涉及的声音，尽量通过tts引擎，以及 Tone.js 、@tonejs/midi 来实现。避免使用音频文件。
  - 前端用tsx,tailwindcss、Zustand、 Framer Motion 来实现。
  - 代码必须用全量的方式给出，不然会丢失现有的代码。不允响应当中使用 许其它的代码保持不变。
  - 需要启用'use client';
  - 全部的文件都保存在当前目录下。 用前缀替代必要的目录层级，SolutionFileNode的FileName 格式如 components-*.ts 或是 store-*.ts 或是 utils-audio.ts, 不包含目录名称；
  - 除了gestureStore 采用import { useGestureStore } from '../components/guesture/gestureStore'; 其它组件要指向当前路径，也就是引用的时候同样应该使用 import { xx } from './yyy'; 的方式来引用。
  - 手势支持相关文档
核心机制:基于Zustand全局状态管理，通过useGestureStore订阅手势
手势以及payload列表
point {x,y}: 食指移动，触发光标跟踪
click {x,y,targetId}: 食指拇指快速捏合，执行主操作
dragstart {x,y,targetId}: 捏住>200ms，开始拖拽
drag {x,y,dx,dy}: 拖拽移动
dragend {x,y}: 释放捏合，结束拖拽
contextmenu {x,y,targetId}: 长按>1s，触发菜单
swipe {direction}: 手掌快速挥动，场景切换
cancel: 掌心正对>500ms，取消操作
transformstart {distance,angle}: 双手捏合
transform {scale,rotation}: 双手缩放/旋转
transformend: 单手释放
消费示例 / typescript
import { useGestureStore } from '../../components/guesture/gestureStore';
const gesture = useGestureStore();
if(gesture.type === 'click' && gesture.payload.targetId === 'my-btn'){
  // 处理点击
}

  - 要使用手势:import { useGestureStore } from '../../components/guesture/gestureStore'; 其定义为：
export const useGestureStore = create<GestureStore>((set) => ({
gesture: { type: 'idle', payload: null, timestamp: Date.now() },
setGesture: (newGesture) => set({ gesture: newGesture }),
}));
`

// ## step1
// 当前的主题是 直观感受比较1-20数字的大小
// 请为这个主题生成一个好玩，体验顺畅的游戏方案

// ## step2
// 很好。以这个方案为基准。请进一步提出解决方案文件的元数据列表，作为下一步高质量实现解决方案的预备。

// 解决方案的每个文件对应有一个元数据：

// type SolutionGraphNode struct {

//     Pathname          string `description:"Ascii pathname of current node。pathname is multi-level, using bullet name to denodes node's modualized intention. extension name such as .md ... is needed"`

//     BulletDescription string `msgpack:"alias:BulletDescription" description:"BulletDescription  是文件内容的摘要。描述和文件的模块化的意图。实现的细节"`

//     FileContent           string `msgpack:"alias:Detail" description:"file stored in disk. compile and running by npm"`

// }

// 解决方案将以文件的方式构建。具体的建构将通过另外的Agent 进行。这些文件内容将根据文件的元描述来创建并完善。请深度分析，创建符合目标数据结构的文件元数据（但请不要填写FileContent）。请确保提出的元数据可以 制导后续文件内容 以模块化、高度洗练、粒度适当的方式 完成对解决方案的建构。

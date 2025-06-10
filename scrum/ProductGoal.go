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
    - 该场景使用背景色作为背景。
    - 所有的场景中的元素使用svg或其它方式，每个元素分别定义。
    - 有可交互的情境。
    - 有适当的环节可以展示知识。
    - 用户的交互操作非常直观、流畅的方式进行。 比如站队、选择、堆放等。
    - 交互要能够通过摄像头，用手势指向来完成。以便用户可以不需要在屏幕前操作，避免近视。
    - 用户不需要非常直接地了解该知识的背景和细节。用户只需要在交互当中，能部分感受到知识的内容。这样可以容纳更高的知识密度。要求该场景可以最有效体现知识的关键作用。
    - 一个微小的主题对应一个确切的微场景。通过海量的场景覆盖所有的需求。
    - 全程在用户参与的情况下，全过程录像可以作为YTB的一个精彩的节目。通过节目推广平台是一个重要策略。要提高场景和互动的节目效果。
    - 其它一些核心的问题，飞过的弹幕乌鸦展示可能的重要疑问，如果点击可以看到相关的回答。
    - 手势 DOM 响应需求 
gesture:point {x, y}: 食指移动；屏幕显示跟随食指的视觉指针，悬停元素高亮。
gesture:enter / gesture:leave: 光标进/出元素，触发悬停 (x,y,target)。
gesture:click {x, y, target}: 食指拇指快速捏合；对高亮元素执行主要操作，元素有点击反馈。
gesture:dragstart {x, y, target}: 食指拇指捏住；元素被“抓住”，视觉上轻微抬升。
gesture:drag {x, y, dx, dy}: 捏住手移动；“抓住”的元素实时平移。
gesture:dragend {x, y}: 捏合手释放；拖拽结束，元素停在当前位置。
gesture:contextmenu {x, y, target}: 食指拇指长按 (>1s)；弹出上下文菜单或详细信息弹窗。
gesture:swipe {direction}: 手掌快速挥动；场景切换或内容滚动。
gesture:cancel: 掌心正对摄像头 (>500ms)；关闭弹窗，取消操作，或返回上级。
gesture:transformstart {distance, angle}: 两手同时进入并捏合；屏幕显示双手指示器。
gesture:transform {scale, rotation}: 两手间距/角度变化；实时缩放/旋转场景或物体。
gesture:transformend: 任意手离开；缩放/旋转结束，物体/场景固定。

## 其它重要约束:
  - 前端用tsx,tailwindcss、Zustand、 Framer Motion 来实现。
`

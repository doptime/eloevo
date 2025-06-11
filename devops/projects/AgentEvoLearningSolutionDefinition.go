package projects

var PlanForTopic = `核心构想：神谕天平 (The Oracle's Scale)
理念: 将数字大小的比较，从“从零开始构建”转变为“对一个已有的状态进行微调与修正”。这极大地缩短了单次游戏的用时，将玩家的注意力集中在“判断差异、快速修正”的核心技能上，形成一个紧凑、上瘾的“观察-操作-验证”循环。

场景与核心元素 (单一场景)
场景: 依然是Aero玻璃效果的“数字神庙”，但所有焦点都集中在中央那个巨大、华丽、响应灵敏的神谕天平上。
左侧托盘 (命题端 The Challenge):
这里是“问题”的展示区。
形式一 (数字命题): 一个巨大的、发光的数字悬浮在托盘上方，例如 “19”。
形式二 (图形命题): 预先排列好的能量球整齐地出现在托盘里，例如一个 4x4 的方阵（代表16）。这需要用户进行快速的视觉估算。
右侧托盘 (解答端 The Workspace):
这里是用户的操作区。
游戏开始时，这里会随机生成一堆数量与左侧有偏差的能量球。例如，左侧是19，右侧可能初始为16或21。
托盘上方有一个实时更新的、醒目的计数器。
操作面板 (The Modifiers):
位于解答端托盘的下方，或用户手部最容易触及的区域。
包含四个玻璃质感的交互按钮：+1, -1, +3, -3。它们是用户调整数量的唯一工具。
审判按钮 (The Judgment Button):
位于天平正下方，一个醒目的圆形按钮，初始文字为“开始审判”。当玩家认为两侧平衡时，将由它来触发最终的验证动画。
完整交互流程 (A Tighter Loop)
序白与开始 (The Prologue)

开场动画：神谕天平在光雾中缓缓现身。柔和的文字浮现：“宇宙万物，皆在均衡。神谕天平，将考验你对数量的感知。”
一个“开始挑战”按钮亮起。
手势: 用户用食指指向并 gesture:click 该按钮。
设定挑战 (The Challenge is Set)

瞬间发生:
左侧 (命题): 一个数字（如“14”）或一堆能量球（如3x5的阵列）瞬间出现在托盘上。
右侧 (解答): 一堆数量有偏差的能量球（如16个）“哗啦”一下落入托盘，上方的计数器显示为 “16”。
下方: +1, -1, +3, -3 四个操作按钮和“开始审判”按钮同时亮起，游戏正式开始。
调整与修正 (The Core Gameplay)

目标: 用户需要通过点击操作按钮，使右侧托盘里的能量球数量与左侧命题匹配。
操作示例:
左侧是“14”，右侧是“16”。用户需要减少2个。
用户用视觉指针 (gesture:point) 指向 -3 按钮，它会高亮。
执行 gesture:click。
即时反馈 (重点!):
视觉: 3个能量球从右侧的堆里“飞出”并消散，或者被吸入-3按钮中。能量球堆肉眼可见地变少。
听觉: 伴随一个“咻”的收缩音效。
数据: 右侧计数器立刻从“16”变为 “13”。
用户发现减多了，于是指向并 gesture:click +1 按钮。
即时反馈:
视觉: 1个新的能量球从+1按钮“发射”出来，带着优雅的弧线落入右侧的球堆中，并有一个Q弹的动画。
听觉: 伴随一个清脆的“叮”声。
数据: 计数器从“13”变为 “14”。
完成状态: 当右侧计数器变为“14”，与左侧命题一致时，“开始审判”按钮会开始发出呼吸灯一般的柔和光芒，强烈暗示用户可以进行下一步。
最终审判 (The Grand Reveal)

用户指向并发光中的“开始审判”按钮，执行 gesture:click。
操作按钮暂时隐藏，所有注意力集中到天平上。
如果平衡 (Correct):
如果左侧是数字命题，那么相应数量的能量球会如流星雨般落下填满左侧托盘，与右侧形成视觉上的对称。
天平的指针精确地指向中央的“0”刻度，然后整个天平爆发出璀璨的金色光芒。
响起胜利、和谐的圣歌音效。屏幕上出现“完美均衡 (Perfect Equilibrium)”的祝贺信息。
如果失衡 (Incorrect):
天平会伴随着沉重的“嘎吱”声和物理惯性，猛地向更重的一方倾斜。
更重一侧的托盘会发出警告性的红光，并响起一个略带滑稽的“失败”音效（比如铜管乐器走调的声音）。
“开始审判”按钮变为“再次尝试”，操作按钮重新出现，用户可以立刻进行修正。
循环 (The Loop)

成功或失败后，一个“新的挑战”按钮会取代中央的提示信息，用户点击后，天平复位，瞬间开始下一轮完全随机的挑战。
为何这个版本更好玩、更流畅？
高节奏，高密度: 砍掉了准备阶段，玩家每一秒都在进行有意义的“判断-操作”行为。从看到问题到动手解决几乎是零延迟。
从“繁”到“巧”: 用户不再需要进行1-20次单调的拖拽，而是通过+3/-3等按钮进行更高效、更具策略性的调整。这减少了重复劳作，增加了思考的乐趣。
即时反馈的爽快感: 每一次点击按钮，都有“一整套”精心设计的视听反馈（球的飞入/飞出、计数器变化、音效），这种即时满足感是游戏体验的核心。Framer Motion 在这里至关重要。
悬念与揭晓的乐趣: “开始审判”环节是每一轮游戏的情感高潮。用户从“我觉得对了”到“系统验证我真的对了”的过程，充满了期待感和成就感，这个华丽的验证动画就是对玩家努力的最高奖赏。
低挫败感: 即使失败，反馈也很有趣且无惩罚性，玩家可以立刻修正，而不是从头再来。这鼓励玩家持续尝试。
技术要点补充
Zustand (状态管理): store 的设计会更简单，主要管理 challengeValue (左侧), currentValue (右侧), 和 gameState (e.g., adjusting, judging)。点击+1按钮，就直接调用 store.getState().increase(1)。
Framer Motion (动画):
使用 AnimatePresence 来处理能量球的增减动画，可以非常轻松地实现它们飞入和飞出的效果。
计数器的数字变化可以使用 motion.h1 并结合 animate 属性，实现平滑的滚动或淡入淡出效果，而不是生硬地跳变。`

var FileItems = `
[
    {
        "Pathname": "package.json",
        "BulletDescription": "项目核心依赖与脚本定义。声明React, Vite, TypeScript基础环境，并明确集成 Zustand, Framer Motion, Tailwind CSS。特别指出需要集成 @mediapipe/tasks-vision 用于手势识别。"
    },
    {
        "Pathname": "vite.config.ts",
        "BulletDescription": "Vite构建工具的配置文件。配置React插件(@vitejs/plugin-react)，确保支持TSX语法和快速热更新。"
    },
    {
        "Pathname": "tailwind.config.js",
        "BulletDescription": "Tailwind CSS配置文件。定义项目的主题颜色、字体，并配置 ‘backdrop-blur‘ 等工具类以支持Aero玻璃效果。"
    },
    {
        "Pathname": "src/main.tsx",
        "BulletDescription": "应用程序的入口文件。负责在DOM中渲染根组件 ‘App‘，并引入全局样式表 ‘index.css‘。"
    },
    {
        "Pathname": "src/index.css",
        "BulletDescription": "全局CSS样式文件。包含Tailwind CSS的‘@tailwind‘指令，并可定义全局背景、字体等基础样式。"
    },
    {
        "Pathname": "src/App.tsx",
        "BulletDescription": "应用的根组件。负责初始化全局服务（如手势识别），并渲染主游戏场景 ‘OracleScaleScene‘。作为整个应用的顶层容器。"
    },
    {
        "Pathname": "src/core/store/gameStore.ts",
        "BulletDescription": "Zustand全局状态管理。定义并导出‘useGameStore‘ hook。Store需包含状态(State)：‘gameState‘ ('prologue'|'challenging'|'judging'|'result'), ‘challenge‘ ({type, value}), ‘workspaceValue‘。以及操作(Actions)：‘startNewChallenge‘, ‘adjustWorkspaceValue‘, ‘judge‘, ‘reset‘。"
    },
    {
        "Pathname": "src/core/gestures/useGestureControls.ts",
        "BulletDescription": "封装手势识别逻辑的React Custom Hook。该Hook负责初始化MediaPipe HandLandmarker，启动摄像头，处理视频流，并将识别到的手势（捏合、移动等）转换为自定义DOM事件（如‘gesture:click‘），同时返回屏幕指针的实时坐标 ‘{x, y}‘。"
    },
    {
        "Pathname": "src/core/gestures/GestureEventDispatcher.ts",
        "BulletDescription": "手势事件分发器模块。被 ‘useGestureControls‘ 调用，负责将原始的手部关键点数据，解析为具体的、语义化的自定义事件（如‘gesture:click‘, ‘gesture:point‘等），并将其分发到 window 或指定的目标元素上。"
    },
    {
        "Pathname": "src/components/scene/OracleScaleScene.tsx",
        "BulletDescription": "核心游戏场景的主组件。调用‘useGestureControls‘启动手势识别，渲染‘GesturePointer‘视觉指针。订阅‘gameStore‘的状态，根据‘gameState‘来动态编排和渲染‘BalanceAltar‘、‘GameMessage‘等场景元素，是所有游戏组件的“导演”。"
    },
    {
        "Pathname": "src/components/scene/BalanceAltar.tsx",
        "BulletDescription": "神谕天平的3D视觉组件。接收props如‘leftValue‘, ‘rightValue‘, ‘isJudging‘。内部包含左右两个‘AltarPan‘。核心功能是使用Framer Motion，根据‘isJudging‘状态和左右重量差，执行精准、富有物理感的倾斜动画。"
    },
    {
        "Pathname": "src/components/scene/AltarPan.tsx",
        "BulletDescription": "天平托盘组件。可复用于左右两侧。接收‘count‘或‘challenge‘作为prop。负责渲染指定数量的‘EnergyOrb‘能量球，并使用Framer Motion的‘AnimatePresence‘来处理球的动态增减动画。同时可能显示托盘上方的数字标签。"
    },
    {
        "Pathname": "src/components/scene/EnergyOrb.tsx",
        "BulletDescription": "能量球的基础组件。使用SVG或带有样式的div实现，确保视觉效果精致。利用Framer Motion实现独立的入场（如弹性掉落）和出场（如溶解消失）动画。它的动画效果是提升游戏体验的关键。"
    },
    {
        "Pathname": "src/components/ui/GesturePointer.tsx",
        "BulletDescription": "屏幕上的手势视觉指针。一个简单的组件，其屏幕位置(x, y)由‘useGestureControls‘ hook的返回值驱动，为用户提供关于其手势指向位置的即时视觉反馈。"
    },
    {
        "Pathname": "src/components/ui/ModifierButton.tsx",
        "BulletDescription": "可复用的操作按钮（+1, -1, +3, -3）。接收‘value‘ (number) 和 ‘onClick‘ 回调作为props。按钮的样式会响应手势悬停(‘gesture:enter‘)而高亮，并响应‘gesture:click‘事件来触发‘onClick‘回调。"
    },
    {
        "Pathname": "src/components/ui/JudgmentButton.tsx",
        "BulletDescription": "“开始审判”按钮。其显示文本和视觉状态（如呼吸灯闪烁效果）由‘gameStore‘驱动。例如，当左右平衡时，按钮变为激活状态并开始发光，吸引用户点击。"
    },
    {
        "Pathname": "src/components/ui/GameMessage.tsx",
        "BulletDescription": "游戏信息展示组件。用于显示开场的序白、胜利的祝贺（“完美均衡”）或失败提示。使用Framer Motion实现优雅的淡入淡出或打字机动画效果。"
    },
    {
        "Pathname": "src/assets/sounds/addOrb.mp3",
        "BulletDescription": "增加能量球时的音效文件。应为清脆、积极的声效。"
    },
    {
        "Pathname": "src/assets/sounds/removeOrb.mp3",
        "BulletDescription": "减少能量球时的音效文件。应为收缩、魔幻的声效。"
    },
    {
        "Pathname": "src/assets/sounds/judgmentSuccess.mp3",
        "BulletDescription": "审判成功、天平平衡时的音效。应为和谐、胜利的圣歌或钟声。"
    },
    {
        "Pathname": "src/assets/sounds/judgmentFail.mp3",
        "BulletDescription": "审判失败、天平失衡时的音效。应为沉重、略带滑稽的失衡声。"
    }
]`

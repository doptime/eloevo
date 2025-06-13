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
    "Filename": "App.tsx",
    "BulletDescription": "Root component that sets up the game environment and orchestrates the main game flow. It will render the 'Oracle's Scale' scene and manage high-level game states like 'adjusting' or 'judging'. It will leverage Zustand for global state management and Framer Motion for scene transitions."
  },
  {
    "Filename": "components-OracleScale.tsx",
    "BulletDescription": "The central component representing the 'Oracle's Scale' itself. This component will handle the rendering of the left and right trays, the balance pointer, and the 'Judgment Button'. It will receive data from the global state and use Framer Motion for all its visual animations, such as tilting and glowing."
  },
  {
    "Filename": "components-Tray.tsx",
    "BulletDescription": "A reusable component for rendering the left and right trays of the scale. It will display either a numerical challenge or a collection of energy balls. It will manage the visual representation of energy balls entering, exiting, or accumulating within the tray, using SVG for elements and Framer Motion for their animations."
  },
  {
    "Filename": "components-EnergyBall.tsx",
    "BulletDescription": "A single energy ball component, likely an SVG, with its own animation properties defined by Framer Motion. This component will be responsible for rendering an individual energy ball and animating its movement (e.g., flying in, flying out, or settling in the tray)."
  },
  {
    "Filename": "components-ModifierButton.tsx",
    "BulletDescription": "A reusable component for the '+1, -1, +3, -3' interaction buttons. Each button will have visual feedback on hover/focus (using Tailwind CSS for styling) and will trigger state updates via the global Zustand store on 'gesture:click'. It will also incorporate Framer Motion for click animations."
  },
  {
    "Filename": "components-JudgmentButton.tsx",
    "BulletDescription": "The 'Start Judgment' button component. It will have a distinct visual style (glass-morphic, glowing). Its appearance will be controlled by the game state, and it will trigger the final judgment sequence when clicked via 'gesture:click'. Framer Motion will be used for its glowing and scaling animations."
  },
  {
    "Filename": "components-NumberCounter.tsx",
    "BulletDescription": "A component to display the real-time count of energy balls on the right tray. This component will receive its value from the Zustand store and use Framer Motion's 'animate' property to achieve smooth, visually appealing transitions (e.g., rolling or fading) when the number changes."
  },
  {
    "Filename": "components-FeedbackOverlay.tsx",
    "BulletDescription": "A component to display visual feedback such as 'Perfect Equilibrium' or 'Try Again' messages. It will appear as an overlay and use Framer Motion for its entrance and exit animations (e.g., fade-in/out, scale-up/down). The content will depend on the game's outcome."
  },
  {
    "Filename": "components-StartChallengeButton.tsx",
    "BulletDescription": "The initial 'Start Challenge' button that appears at the beginning of the game to initiate the first challenge. It will respond to 'gesture:click' to transition the game into the 'challenge set' state."
  },
  {
    "Filename": "store-gestureStore.ts",
    "BulletDescription": "Zustand store for managing global gesture state. This file will define the store and its actions to subscribe to and update gesture inputs (point, click, dragstart, drag, dragend, contextmenu, swipe, cancel, transformstart, transform, transformend) as per the provided specification. This is a critical component for integrating hand gestures."
  },
  {
    "Filename": "store-gameStore.ts",
    "BulletDescription": "Zustand store for managing the core game state. This will include challengeValue (the target number/quantity), currentValue (the user's current quantity on the right tray), and gameState (e.g., 'idle', 'adjusting', 'judging', 'correct', 'incorrect'). It will also define actions to modify these states, such as 'increase', 'decrease', 'setChallenge', and 'triggerJudgment'."
  },
  {
    "Filename": "utils-audio.ts",
    "BulletDescription": "Utility file for managing sound effects using TTS engine and Tone.js. It will contain functions to play specific sounds for actions like adding/removing energy balls, a 'ding' for adding, a 'swoosh' for removing, and distinct sounds for 'correct' and 'incorrect' judgments. This will avoid audio files and rely on programmatically generated sounds."
  },
  {
    "Filename": "assets-svg-scale-icon.svg",
    "BulletDescription": "SVG file for the main scale icon and its components (trays, pointer). This will be imported and used within the OracleScale component."
  },
  {
    "Filename": "assets-svg-energy-ball.svg",
    "BulletDescription": "SVG file for the energy ball visual element. This will be imported and used within the EnergyBall component."
  },
  {
    "Filename": "styles-index.css",
    "BulletDescription": "Main Tailwind CSS file for global styles and utility classes. This will include custom styles for glass-morphic effects, button states, and general layout."
  },
  {
    "Filename": "hooks-useGestureHandler.ts",
    "BulletDescription": "A custom React hook that integrates with the "useGestureStore" to translate raw gesture events into specific game actions. This hook will listen to changes in the gesture store and dispatch actions to the "gameStore" (e.g., on 'click' of a ModifierButton, update "currentValue")."
  }
]
  `

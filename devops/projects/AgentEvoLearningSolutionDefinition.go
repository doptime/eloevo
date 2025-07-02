package projects

var PlanForWordSensation = `
认识单词游戏:
- 开始新游戏，系统随机从潜在单词列表当中选择一个目标单词。这个单词是我们要认识的单词。
  潜在单词列表从网页参数wordlist中获取。wordList采用逗号分隔的字符串。 默认空白列表替换为apple,window,fiction,开心，快乐。

- 系统朗读目标单词。一次正常速度，一次0.3倍速。
- 然后给定3个选项卡，要求从里面选择一个选项卡作为答案，每个选项卡上有相应的提示语。
    选项卡的提示内容包括中、英、日、西班牙语，emoji. 以及目标单词的词根、联想、svg 等。尽可能多地和这个单词建立关联。提词内容必须凑到60个符号左右。
    选项看的内容应该是一个文本字符串，包含多种语言的提示语。但是需要引入换行符，以便以多行显式不同组别的内容，以使得内容美观，容易理解。
- 选项卡选择后，翻面揭示选项卡背后的答案文字，并比对，告知答案是否正确。正确答案的选项卡背后的答案文字就是我们要认识的单词。

- 选项卡的下方有1个发音喇叭按钮（不包含文字），短按钮的内容会自动发音，一次正常发音，一次0.3倍速慢速播放题目。
- 题目选择结束后，也应该使用语音播报结果。
- 当点击选项卡后。该按钮转变成为新的游戏按钮

- 选择正确答案后翻转成镜像对称，此时应该继续再进行一次镜像翻转，翻转成原先的样子（以便能看清），并且放大显示（用来表明选中）。
- 朗读发音的时候，滴声播放结束后才能播放文本的语音。
- 分离数据和数据操作到单独的文件


这个方案的核心是一个名为LiteracyGame的组件。它管理着整个游戏的流程：
游戏初始化：随机选择一个目标单词，并生成三个选项卡，其中一个包含正确答案。
交互与反馈：您可以通过模拟的手势（或鼠标点击）来选择卡片。卡片会以3D翻转的动画效果揭示背后的答案。
多维度提示：每个卡片的正面都集成了多语言、Emoji、词根、联想和SVG图形等多种提示信息，以最丰富的方式帮助建立认知关联。
动态按钮：下方的控制按钮会根据游戏状态在“发音”和“新游戏”之间切换。
手势光标：屏幕上有一个虚拟光标，它会响应point手势，模拟真实的手势指向操作。

单词发音: 使用 Web Speech API 
背面卡片的内容: 使用doptime-client从服务端读取。


完整的游戏逻辑：使用Zustand进行清晰的状态管理。
丰富的交互动画：使用Framer Motion实现卡片翻转和元素进出场动画，视觉效果流畅。
手势交互集成：代码中包含了监听和响应useGestureStore的逻辑，并提供了一个鼠标模拟层，以便在没有真实手势输入设备时进行测试。
音频反馈：集成了基于Web API的TTS发音和Tone.js的音效，符合不使用音频文件的约束。
模块化和可扩展性：代码结构清晰，通过注释划分了不同的“虚拟文件”，方便您未来将其拆分到真实的文件结构中。单词库wordDatabase也很容易扩展。
对接真实手势输入：将模拟的useGestureStore替换为项目中实际的import。确保您的手势识别模块能够正确地更新Zustand store中的gesture状态。
扩展词库：在wordDatabase中添加更多的单词和对应的多维度提示信息。您可以考虑将这个数据结构从后端API获取，使其更具动态性。
优化提示内容：目前提示语是硬编码的。您可以设计一个算法，根据目标单词的特性自动生成或组合更有趣的提示，确保其长度和吸引力。
增加游戏模式：可以增加计时挑战、连续答对奖励等更多游戏化元素，提高趣味性和重玩价值。
录制与分享：考虑集成录屏功能，让用户可以方便地将他们的游戏过程录制下来，一键分享到社交平台，这会是推广产品的一个绝佳途径。
点击答案后的视觉反馈：1. 被点击的卡片 镜面翻转；2。如果是正确答案，则再次镜面翻转，并且被放大; 3.正确答案卡，需要放大；其它不是正确答案且未被点击的选项，不需要视觉反馈。

`
var PlanForNumberSensation = `核心构想：神谕天平 (The Oracle's Scale)
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
计数器的数字变化可以使用 motion.h1 并结合 animate 属性，实现平滑的滚动或淡入淡出效果，而不是生硬地跳变。



- 整个屏幕应该尽可能使用一种俯视视角，使得球可以以更大面积更趋向平铺在屏幕上。
- 左右两个球所在托盘区域不应该重叠（影响分辨属于哪一边）；但也不应该过于分离。
- 球应该设置多种显眼，容易区分的颜色
- 侧边应该采用隐形挡板，能够把球回弹到可见区域，且不影响看清楚球。
- 如果用户的球数和目标数字不匹配，不应该视为失败。和目标数字的差别在3之内都应该视为成功，给与成功提示音，偏差越小，响应越是热烈。
- 游戏结束后能显示游戏结果，同时能顺利继续游戏。


需要做的改进列表:
调整目标:左侧有一半的概率出现数字命题，而不是只有图形命题。现在用数字命题时候，无法显示数字内容，应该修复此Bug。(修复 isNumericChallenge ===true 时，显示数字内容无法显示， 或者是 isNumericChallenge 状态异常)
调整目标: 重构并且简化现有的代码
重新开始游戏后缺乏”开始审判“。无法查看结果。调整目标:重新开始游戏后，能够显示”开始审判“按钮，并且能够查看结果。
`

var FileItems = `[
  {
    "Filename": "page.tsx",
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
    "BulletDescription": "A custom React hook that integrates with the 'useGestureStore' to translate raw gesture events into specific game actions. This hook will listen to changes in the gesture store and dispatch actions to the 'gameStore' (e.g., on 'click' of a ModifierButton, update 'currentValue')."
  }
]`

# 业务需求说明书（BRD, Business Requirement Document 从业务的角度描述系统需要解决的问题、业务目标和期望的商业价值）：
# 功能需求规格说明书（FRS, Functional Requirement Specification，包括用户期望的功能、行为和约束）： 
  - 系统可以根据requirement.md中的需求说明，自动编写/重写 设计文档.evolab/designs.md。
  - 设计文档应当以.evolab/designs.templete.md为模板，包括相应的内容。

# 非功能需求说明（NFR, Non-functional Requirement Document,系统的性能、安全性、可扩展性等非功能方面的需求）：
  - 场景需求定义在本文件中定义。需求被看做是严格的硬性约束。
  - designs.md 的迭代 采用OopenAI o1 / Gemini 2.0 Flash Thinking来生成
  - feedbacks_xxxx.md 由 Qwen2.5-72B 生成。 或者是由编译器生成
  - 具体修改系统文件/编码 由DeepSeek V3生成。DeepSeek v3 输入64K。 输出8K。成本是OpenAI 4o 的1/20， 但是效果显著优于OpenAI 4o.
  - 目标系统重的实现目标的数据和代码结构被看做是真正的目标。限定迭代后的目标和之前的目标版本具有内容上的稳定性。

# 用户场景和用例文档 (实际使用场景和用例，展示功能如何满足用户需求；每个用户故事应包含明确的输入、输出和验收标准)： 
  - 本系统名为evolab。用来迭代式地自动完成软件系统或其它系统的开发



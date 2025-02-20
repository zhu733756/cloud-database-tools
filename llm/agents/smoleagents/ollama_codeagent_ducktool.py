from smolagents import CodeAgent, LiteLLMModel, DuckDuckGoSearchTool

# 配置 Ollama 模型
model = LiteLLMModel(
    model_id="ollama_chat/deepseek-r1:7b",
    base_url="http://192.168.137.112:11434",
)

# 添加工具
tools = [DuckDuckGoSearchTool()]  # 添加 DuckDuckGo 搜索工具

# 初始化 SmolAgents 的 CodeAgent
agent = CodeAgent(tools=tools, model=model, add_base_tools=True)

# 测试智能体
output = agent.run("What is the latest news about AI?")
print("Final output:")
print(output)


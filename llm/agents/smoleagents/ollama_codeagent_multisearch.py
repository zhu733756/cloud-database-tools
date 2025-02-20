from smolagents import (
    CodeAgent,
    LiteLLMModel,
    DuckDuckGoSearchTool,
    ToolCallingAgent,
    VisitWebpageTool,
)


model = LiteLLMModel(
    model_id="ollama_chat/deepseek-r1:7b", # 格式是ollama_chat/xxx
    base_url="http://192.168.137.13:11434",
    api_key=None, # 默认为 None
    num_ctx=8192
)

search_agent = ToolCallingAgent(
    tools=[DuckDuckGoSearchTool(), VisitWebpageTool()],
    model=model,
    name="search_agent",
    description="This is an agent that can do web search.",
    max_steps=10,
)

manager_agent = CodeAgent(
    tools=[],
    model=model,
    managed_agents=[search_agent],
)

out = manager_agent.run("If the US keeps it 2024 growth rate, how many years would it take for the GDP to double?")
print(out)

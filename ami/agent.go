package ami

// Agents lists agents and their status.
func Agents(client Client, actionID string) (Response, error) {
	return send(client, "Agents", actionID, nil)
}

// AgentLogoff sets an agent as no longer logged in.
func AgentLogoff(client Client, actionID, agent string, soft bool) (Response, error) {
	return send(client, "AgentLogoff", actionID, map[string]interface{}{
		"Agent": agent,
		"Soft":  soft,
	})
}

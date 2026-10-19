package screens

// Advice is display-only. The current value always comes from the entry's
// UserExperience projection, never from this recommendation or keyboard focus.
func shellExperienceRecommendation(entry ShellExperienceEntry) (string, string) {
	switch entry.Control {
	case ShellExperiencePersona:
		switch entry.Choice {
		case "neutral":
			return "Gentleman", "Use Neutral when you prefer neutral conversational wording."
		case "unmanaged":
			return "Gentleman", "Use Custom when you already manage your own persona."
		default:
			return "Gentleman", "It provides the complete Gentle-Shell ODD experience."
		}
	case ShellExperienceBackground:
		return "ON", "Run independent work without blocking your terminal."
	case ShellExperienceTDD:
		return "ON", "Prove behavior with a failing test before implementation."
	case ShellExperienceRDD:
		return "ON", "Independent review preserves evidence; it applies globally and costs extra reviewer-agent tokens."
	case ShellExperienceCodeGraph:
		return "ON", "Use only with a qualified private Linux amd64 Separate runtime; project overrides remain."
	case ShellExperienceContext7:
		return "ON", "Use current library references through MCP."
	case ShellExperienceSDK:
		if entry.Choice == "main-preview" {
			return "Stable", "Use SDK Main only on a qualified Linux amd64 fresh Separate install; provider and native pins stay stock."
		}
		return "Stable", "Keep the stock SDK source, provider and native pins."
	case ShellExperienceGGA:
		return "OFF", "Opt in only with an explicit provider; no automatic hooks or provider inference."
	case ShellExperienceProvider:
		return "Explicit choice", "Type your provider; it is never inferred from Pi configuration."
	default:
		return "", ""
	}
}

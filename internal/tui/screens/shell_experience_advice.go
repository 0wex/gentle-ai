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
		return "ON", "Use only on Linux amd64; the runtime stays private to this install (Separate or launcher-owned Shared). Project overrides remain."
	case ShellExperienceContext7:
		return "ON", "Use current library references through MCP."
	case ShellExperienceInstallTarget:
		if entry.Choice == "existing" {
			return "Configure existing", "Prefills this install's selections; continuing replaces only its target and keeps the previous tree."
		}
		if entry.Choice == "shared" {
			return "Separate install", "Reuse a detected Pi only when you want the defaults applied over it; Main needs Separate."
		}
		return "Separate install", "Keeps your personal Pi untouched and installs its own prefix, agent and launchers."
	case ShellExperienceSDK:
		if entry.Choice == "main-preview" {
			return "Stable", "Use SDK Main only on a qualified Linux amd64 fresh Separate install; provider and native pins stay stock."
		}
		return "Stable", "Keep the stock SDK source, provider and native pins."
	case ShellExperienceGGA:
		return "OFF", "Opt in with or without a provider; GGA configures it later when absent. No automatic hooks or provider inference."
	case ShellExperienceProvider:
		return "Explicit choice", "Type your provider if you want; leave it empty and GGA configures it later. It is never inferred from Pi."
	default:
		return "", ""
	}
}

package ui

// Theme represents a UI theme
type Theme struct {
	Name      string      `json:"name"`
	Colors    ThemeColors `json:"colors"`
	Fonts     ThemeFonts  `json:"fonts"`
	CustomCSS string      `json:"customCss,omitempty"`
}

// ThemeColors defines theme color palette
type ThemeColors struct {
	Primary    string `json:"primary"`
	Secondary  string `json:"secondary"`
	Background string `json:"background"`
	Surface    string `json:"surface"`
	Text       string `json:"text"`
	TextMuted  string `json:"textMuted"`
	Border     string `json:"border"`
	Success    string `json:"success"`
	Warning    string `json:"warning"`
	Error      string `json:"error"`
	Info       string `json:"info"`
}

// ThemeFonts defines theme fonts
type ThemeFonts struct {
	Body string `json:"body"`
	Code string `json:"code"`
}

// PredefinedThemes contains built-in themes
var PredefinedThemes = map[string]Theme{
	"purple": {
		Name: "purple",
		Colors: ThemeColors{
			Primary:    "#8B5CF6",
			Secondary:  "#A78BFA",
			Background: "#0F0F23",
			Surface:    "#1A1A2E",
			Text:       "#FFFFFF",
			TextMuted:  "#9CA3AF",
			Border:     "#374151",
			Success:    "#10B981",
			Warning:    "#F59E0B",
			Error:      "#EF4444",
			Info:       "#3B82F6",
		},
		Fonts: ThemeFonts{
			Body: "Inter, system-ui, sans-serif",
			Code: "JetBrains Mono, monospace",
		},
	},
	"blue": {
		Name: "blue",
		Colors: ThemeColors{
			Primary:    "#3B82F6",
			Secondary:  "#60A5FA",
			Background: "#0F172A",
			Surface:    "#1E293B",
			Text:       "#F8FAFC",
			TextMuted:  "#94A3B8",
			Border:     "#334155",
			Success:    "#22C55E",
			Warning:    "#EAB308",
			Error:      "#EF4444",
			Info:       "#0EA5E9",
		},
		Fonts: ThemeFonts{
			Body: "Inter, system-ui, sans-serif",
			Code: "Fira Code, monospace",
		},
	},
	"green": {
		Name: "green",
		Colors: ThemeColors{
			Primary:    "#10B981",
			Secondary:  "#34D399",
			Background: "#0D1117",
			Surface:    "#161B22",
			Text:       "#F0F6FC",
			TextMuted:  "#8B949E",
			Border:     "#30363D",
			Success:    "#3FB950",
			Warning:    "#D29922",
			Error:      "#F85149",
			Info:       "#58A6FF",
		},
		Fonts: ThemeFonts{
			Body: "Inter, system-ui, sans-serif",
			Code: "Source Code Pro, monospace",
		},
	},
	"light": {
		Name: "light",
		Colors: ThemeColors{
			Primary:    "#6366F1",
			Secondary:  "#818CF8",
			Background: "#FFFFFF",
			Surface:    "#F9FAFB",
			Text:       "#111827",
			TextMuted:  "#6B7280",
			Border:     "#E5E7EB",
			Success:    "#059669",
			Warning:    "#D97706",
			Error:      "#DC2626",
			Info:       "#2563EB",
		},
		Fonts: ThemeFonts{
			Body: "Inter, system-ui, sans-serif",
			Code: "JetBrains Mono, monospace",
		},
	},
}

// GetTheme returns a predefined theme by name
func GetTheme(name string) (Theme, bool) {
	theme, exists := PredefinedThemes[name]
	return theme, exists
}

// ToCSS converts a theme to CSS custom properties
func (t Theme) ToCSS() string {
	return `
:root {
	--color-primary: ` + t.Colors.Primary + `;
	--color-secondary: ` + t.Colors.Secondary + `;
	--color-background: ` + t.Colors.Background + `;
	--color-surface: ` + t.Colors.Surface + `;
	--color-text: ` + t.Colors.Text + `;
	--color-text-muted: ` + t.Colors.TextMuted + `;
	--color-border: ` + t.Colors.Border + `;
	--color-success: ` + t.Colors.Success + `;
	--color-warning: ` + t.Colors.Warning + `;
	--color-error: ` + t.Colors.Error + `;
	--color-info: ` + t.Colors.Info + `;
	--font-body: ` + t.Fonts.Body + `;
	--font-code: ` + t.Fonts.Code + `;
}
` + t.CustomCSS
}

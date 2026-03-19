package css

type StyleSheet struct {
	Rules []Rule
}

type Rule struct {
	Selectors    []string
	Declarations []Declaration
}

type Declaration struct {
	Property string
	Value    string
}

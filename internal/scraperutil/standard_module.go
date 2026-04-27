package scraperutil

type StandardModule struct {
	ScraperName        string
	ScraperDescription string
	ScraperOptions     []any
	ScraperDefaults    any
	ScraperPriority    int
	ConfigType         func() ScraperConfigInterface
	NewScraperFunc     any
	FlatOverrides      FlattenOverrides
	FlatBuilder        SettingsBuilder
	FlatBuilderRaw     SettingsBuilderWithRaw
	UseRawBuilder      bool
}

func (m *StandardModule) Name() string        { return m.ScraperName }
func (m *StandardModule) Description() string { return m.ScraperDescription }
func (m *StandardModule) Options() any        { return m.ScraperOptions }
func (m *StandardModule) Defaults() any       { return m.ScraperDefaults }
func (m *StandardModule) Priority() int       { return m.ScraperPriority }

func (m *StandardModule) Constructor() any {
	return m.NewScraperFunc
}

func (m *StandardModule) Validator() any {
	return ValidatorFunc(func(a any) error {
		cfg := m.ConfigType()
		type validator interface {
			ValidateConfig(any) error
		}
		if v, ok := cfg.(validator); ok {
			return v.ValidateConfig(a)
		}
		return nil
	})
}

func (m *StandardModule) ConfigFactory() any {
	return ConfigFactory(func() any { return m.ConfigType() })
}

func (m *StandardModule) FlattenFunc() any {
	if m.UseRawBuilder && m.FlatBuilderRaw != nil {
		return DefaultFlattenConfigWithRaw(m.FlatOverrides, m.FlatBuilderRaw)
	}
	if m.FlatBuilder != nil {
		return DefaultFlattenConfig(m.FlatOverrides, m.FlatBuilder)
	}
	return nil
}

var _ ScraperModule = (*StandardModule)(nil)

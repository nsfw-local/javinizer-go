package migrations

// Future migration placeholder for v3 → v4
//
// To implement this migration when needed:
// 1. Uncomment the code below
// 2. Update CurrentConfigVersion in config.go to 4
// 3. Implement the Migrate() function with the transformation logic
//
// Example implementation:
//
// func init() {
//     config.RegisterMigration(&v3ToV4Migration{})
// }
//
// type v3ToV4Migration struct{}
//
// func (m *v3ToV4Migration) FromVersions() []int { return []int{3} }
// func (m *v3ToV4Migration) ToVersion() int      { return 4 }
// func (m *v3ToV4Migration) Description() string { return "Add new field X" }
// func (m *v3ToV4Migration) Migrate(cfg *config.Config) error {
//     cfg.NewField = "default_value"
//     return nil
// }

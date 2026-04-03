# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- SettingsHash() method in TranslationConfig for translation cache invalidation based on configuration changes
- Hash computation includes all output-affecting settings (provider, languages, models, modes) while excluding infrastructure settings (api_key, base_url, timeout)
- Provider-specific settings in hash: OpenAI.Model, OpenAICompatible.Model, Anthropic.Model, DeepL.Mode, Google.Mode
- CLI scrape command now applies configured translations to scraped metadata

### Fixed

- Frontend provider name consistency: changed `openai_compatible` to `openai-compatible` for better UX
- Added validation for openai-compatible provider requiring model field to be set
- Made ApplyConfiguredTranslation() public for use across packages

### Changed

- Translation configuration provider value standardized to `openai-compatible` (frontend now matches backend JSON tag)
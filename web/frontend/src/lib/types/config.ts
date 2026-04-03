/**
 * Shared TypeScript interfaces for Javinizer configuration
 * These should be kept in sync with the Go backend types
 */

export interface BrowserConfig {
	enabled: boolean;
	binary_path?: string;
	timeout: number;
	max_retries: number;
	headless: boolean;
	stealth_mode: boolean;
	window_width: number;
	window_height: number;
	slow_mo: number;
	block_images: boolean;
	block_css: boolean;
	user_agent?: string;
	debug_visible: boolean;
}

export interface ScrapersConfig {
	browser?: BrowserConfig;
	scrape_actress?: boolean;
	[key: string]: any;
}

export interface Config {
	scrapers?: ScrapersConfig;
	[key: string]: any;
}

export interface ApiToken {
	id: string;
	token_prefix: string;
	name: string;
	created_at: string;
	last_used_at: string | null;
	revoked_at: string | null;
}

export interface CreateTokenResponse {
	token: string;
	id: string;
	name: string;
	created_at: string;
}

export interface TokenListResponse {
	tokens: ApiToken[];
	count: number;
}

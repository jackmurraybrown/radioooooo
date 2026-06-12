// ✮⋆‧°—°‧⋆✮ clean re-exports from the generated spec — edit types.gen.ts by running pnpm generate-types
import type { components } from './types.gen'

type S = components['schemas']

export type Channel = S['Channel']
export type Episode = S['Episode']
export type EpisodeBody = S['EpisodeBody']
export type Media           = S['Media']
export type MediaCreateBody = S['MediaCreateBody']
export type MediaUpdateBody = S['MediaUpdateBody']
export type Playlist           = S['Playlist']
export type PlaylistItem       = S['PlaylistItem']
export type PlaylistCreateBody = S['PlaylistCreateBody']
export type PlaylistUpdateBody = S['PlaylistUpdateBody']
// ⋆˙⟡ logoUrl added in migration 007 — extend until pnpm generate-types is re-run against the updated spec
export type Station = S['Station'] & { logoUrl?: string | null }
export type User = S['User']

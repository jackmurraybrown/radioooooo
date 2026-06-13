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
export type Station = S['Station']
export type User = S['User']

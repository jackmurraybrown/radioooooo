// ✮⋆‧°—°‧⋆✮ clean re-exports from the generated spec — edit types.gen.ts by running pnpm generate-types
import type { components } from './types.gen'

type S = components['schemas']

export type Channel = S['Channel']
export type Episode = S['Episode']
export type EpisodeBody = S['EpisodeBody']
export type Media = S['Media']
export type Playlist = S['Playlist']
export type PlaylistItem = S['PlaylistItem']
export type Station = S['Station']
export type User = S['User']

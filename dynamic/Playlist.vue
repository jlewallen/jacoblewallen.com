<template>
<div class="playlist container">
	<div class="header row">
		<div class="col-md-12">
		</div>
	</div>
	<div v-if="playlist.playlist">
		<div class="header row">
			<div class="cover col-md-3">
				<img :src="playlist.playlist.images[1].url"></img>
			</div>
			<div class="details col-md-9">
				<h2>
					<a href="" v-on:click="openPlaylists($event)">Playlists</a> / {{ playlist.playlist.name }}
				</h2>
			</div>
		</div>
		<div class="row">
		</div>
		<div class="row">
			<div class="tracks col-md-12">
				<table>
					<thead>
						<tr>
							<th class="track" style="width: 33%;">Track</th>
							<th class="artist" style="width: 33%;">Artist</th>
							<th class="album" style="width: 33%;">Album</th>
						</tr>
					</thead>
					<tbody>
						<tr v-for="track in playlist.tracks" class="track">
							<td class="track">{{ track.track.name }}</td>
							<td class="artist">{{ track.track.artists | joinArtists }}</td>
							<td class="album">{{ track.track.album.name }}</td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>
	</div>
</div>
</template>
<script>
import _ from 'lodash'
import moment from 'moment'

import { Playlists } from "./playlists"

export default {
	name: 'Playlist',
	props: {
		query: {
			required: true
		}
	},
	data: () => {
		return {
			playlist: {}
		}
	},
	created() {
		this.refresh()
	},
	filters: {
		prettyTime(value) {
			return moment(value).format("MMM Do YYYY")
		},
		joinArtists(value) {
			return _(value).map("name").join(", ")
		}
	},
	methods: {
		refresh() {
			new Playlists().playlist(this.$props.query.id).then(playlist => {
				this.playlist = playlist
			});
		},
		openPlaylists(ev) {
			ev.preventDefault()
			this.$emit('navigate', '')
		}
	}
}
</script>
<style>
</style>

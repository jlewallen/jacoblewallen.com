<template>
<div v-masonry item-selector=".item" class="playlists">
	<div v-masonry-tile class="folder" v-for="folder in folders">
		<h3>{{folder.name}}</h3>
		<ul>
			<li v-for="playlist in folder.playlists" class="playlist" :class="{ 'recently-modified': playlist.recentlyModified }">
				<a :href="'?id=' + playlist.id">{{playlist.name}}</a>
				<span class="details"> {{playlist.numberOfTracks}} tracks, modified {{playlist.lastModified | prettyTime}}</span>
			</li>
		</ul>
	</div>
</div>
</template>
<script>
import moment from 'moment'

import { Playlists } from "./playlists"

export default {
	name: 'Home',
	data: () => {
		return {
			folders: []
		}
	},
	created() {
		this.refresh()
	},
	filters: {
		prettyTime(value) {
			return moment(value).format("MMM Do YYYY");
		}
	},
	methods: {
		refresh() {
			new Playlists().playlists().then(data => {
				this.folders = data.folders
			});
		}
	}
}
</script>
<style>
</style>

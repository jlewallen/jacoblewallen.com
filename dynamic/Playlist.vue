<template>
<div>
</div>
</template>
<script>
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
		console.log(this.$props)
		this.refresh()
	},
	filters: {
		prettyTime(value) {
			return moment(value).format("MMM Do YY");
		}
	},
	methods: {
		refresh() {
			new Playlists().playlist(this.$props.query.id).then(playlist => {
				console.log(playlist)
				this.playlist = playlist
			});
		}
	}
}
</script>
<style>
</style>
